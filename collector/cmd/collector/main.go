package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"

	"github.com/cti-dashboard/collector/internal/processor"
	"github.com/cti-dashboard/collector/internal/repository"
	"github.com/cti-dashboard/collector/internal/scheduler"
	"github.com/cti-dashboard/collector/internal/scraper"
)

type Config struct {
	DBHost             string
	DBPort             string
	DBName             string
	DBUser             string
	DBPassword         string
	LogLevel           string
	CollectionInterval time.Duration
	SourcesFile        string
}

func loadConfig() *Config {
	intervalStr := getEnv("COLLECTION_INTERVAL", "5m")
	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		log.WithError(err).Warn("Invalid COLLECTION_INTERVAL, using default 5m")
		interval = 5 * time.Minute
	}

	return &Config{
		DBHost:             getEnv("DB_HOST", "postgres"),
		DBPort:             getEnv("DB_PORT", "5432"),
		DBName:             getEnv("DB_NAME", "cti_db"),
		DBUser:             getEnv("DB_USER", "cti_user"),
		DBPassword:         getEnv("DB_PASSWORD", ""),
		LogLevel:           getEnv("LOG_LEVEL", "info"),
		CollectionInterval: interval,
		SourcesFile:        getEnv("SOURCES_FILE", "/app/sources.json"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func initDB(cfg *Config) (*sql.DB, error) {
	log.WithFields(log.Fields{
		"host":     cfg.DBHost,
		"port":     cfg.DBPort,
		"database": cfg.DBName,
	}).Info("Initializing database connection")

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Debug("Testing database connection...")

	maxRetries := 5
	for i := 1; i <= maxRetries; i++ {
		err = db.Ping()
		if err == nil {
			log.Info("Database connection established successfully")
			return db, nil
		}

		log.WithFields(log.Fields{
			"attempt": i,
			"max":     maxRetries,
			"error":   err.Error(),
		}).Warn("Failed to ping database, retrying...")

		if i < maxRetries {
			time.Sleep(time.Duration(i) * time.Second)
		}
	}

	return nil, fmt.Errorf("ping failed after %d attempts: %w", maxRetries, err)
}

func loadSources(filename string) ([]scraper.Source, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var sources []scraper.Source
	if err := json.Unmarshal(data, &sources); err != nil {
		return nil, err
	}

	return sources, nil
}

type Collector struct {
	webScraper   *scraper.WebScraper
	rssFetcher   *scraper.RSSFetcher
	forumScraper *scraper.ForumScraper
	titleGen     *processor.RuleBasedGenerator
	scorer       *processor.CriticalityScorer
	repository   *repository.ContentRepository
}

func NewCollector(db *sql.DB, sources []scraper.Source) *Collector {
	return &Collector{
		webScraper:   scraper.NewWebScraper(sources),
		rssFetcher:   scraper.NewRSSFetcher(),
		forumScraper: scraper.NewForumScraper(),
		titleGen:     processor.NewRuleBasedGenerator(),
		scorer:       processor.NewCriticalityScorer(),
		repository:   repository.NewContentRepository(db),
	}
}

func (c *Collector) Collect(sources []scraper.Source) error {
	log.Info("Starting collection cycle")

	successCount := 0
	errorCount := 0

	for _, source := range sources {
		log.WithField("source", source.Name).Info("Scraping source")
		
		var items []scraper.ScrapedContent
		var err error
		
		// handle .onion sites differently
		if strings.Contains(source.URL, ".onion") {
			log.WithField("source", source.Name).Info("Deep scraping forum")
			items, err = c.forumScraper.ScrapeForumDeep(source)
			if err != nil {
				log.WithFields(log.Fields{
					"source": source.Name,
					"error":  err.Error(),
				}).Error("Failed to deep scrape forum")
				errorCount++
				continue
			}
		} else {
			// try RSS first for feeds
			if strings.Contains(source.URL, "/feed") || 
			   strings.Contains(source.URL, "/rss") || 
			   strings.Contains(source.URL, ".xml") {
				log.WithField("source", source.Name).Debug("Attempting RSS fetch")
				items, err = c.rssFetcher.FetchFeed(source)
			}
			
			// fallback to HTML scraping
			if err != nil || len(items) == 0 {
				log.WithField("source", source.Name).Debug("Attempting HTML scraping")
				scraped, scrapeErr := c.webScraper.ScrapeSource(source)
				if scrapeErr != nil {
					log.WithFields(log.Fields{
						"source": source.Name,
						"error":  scrapeErr.Error(),
					}).Error("Failed to scrape source")
					errorCount++
					continue
				}
				items = []scraper.ScrapedContent{scraped}
			}
		}

		log.WithFields(log.Fields{
			"source": source.Name,
			"items":  len(items),
		}).Info("Scraped items")

		for _, scraped := range items {
			// generate title
			title := c.titleGen.Generate(scraped.Content)
			if title == "" {
				title = source.Name
			}

			// calculate criticality score
			criticalityScore := c.scorer.Calculate(scraped.Content, []string{})
			
			// auto-categorize content
			categories := c.scorer.AutoCategorize(scraped.Content)

			item := &repository.ContentItem{
				Title:            title,
				SourceName:       source.Name,
				SourceURL:        scraped.Source.URL,
				Content:          scraped.Content,
				PublishedAt:      scraped.PublishedAt,
				CriticalityScore: criticalityScore,
				Categories:       categories,
			}

			// check if already exists
			exists, err := c.repository.ExistsByURL(scraped.Source.URL)
			if err != nil {
				log.WithError(err).Warn("Failed to check if content exists")
			}
			if exists {
				log.WithField("url", scraped.Source.URL).Debug("Content already exists, skipping")
				continue
			}

			// insert into database
			err = c.repository.Insert(item)
			if err != nil {
				log.WithFields(log.Fields{
					"title":  title,
					"source": source.Name,
					"error":  err.Error(),
				}).Error("Failed to insert content item")
				errorCount++
				continue
			}

			log.WithFields(log.Fields{
				"title":  title,
				"source": source.Name,
				"score":  criticalityScore,
			}).Info("Content item inserted successfully")

			successCount++
		}
	}

	log.WithFields(log.Fields{
		"success": successCount,
		"errors":  errorCount,
	}).Info("Collection cycle completed")

	return nil
}

func main() {
	cfg := loadConfig()

	level, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.WithError(err).Warn("Invalid log level, using INFO")
		level = log.InfoLevel
	}
	log.SetLevel(level)
	log.SetFormatter(&log.JSONFormatter{})

	log.WithFields(log.Fields{
		"version":             "1.0.0",
		"log_level":           cfg.LogLevel,
		"collection_interval": cfg.CollectionInterval,
		"sources_file":        cfg.SourcesFile,
	}).Info("Starting CTI Data Collector (Real Scraping Mode)")

	sources, err := loadSources(cfg.SourcesFile)
	if err != nil {
		log.WithError(err).Fatal("Failed to load sources")
	}

	if len(sources) == 0 {
		log.Fatal("No sources configured. Please add sources to sources.json")
	}

	log.WithField("count", len(sources)).Info("Loaded sources")

	db, err := initDB(cfg)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize database")
	}
	defer func() {
		log.Info("Closing database connection")
		db.Close()
	}()

	collector := NewCollector(db, sources)

	collectionFunc := func() error {
		return collector.Collect(sources)
	}

	sched := scheduler.NewScheduler(cfg.CollectionInterval, collectionFunc)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go sched.Start(ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	sig := <-sigChan
	log.WithField("signal", sig.String()).Info("Shutdown signal received")

	sched.Stop()

	log.Info("CTI Data Collector shutdown complete")
}
