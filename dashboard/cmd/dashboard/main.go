package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/cti-dashboard/dashboard/internal/auth"
	"github.com/cti-dashboard/dashboard/internal/config"
	"github.com/cti-dashboard/dashboard/internal/database"
	"github.com/cti-dashboard/dashboard/internal/handlers"
	"github.com/cti-dashboard/dashboard/internal/middleware"
	"github.com/cti-dashboard/dashboard/internal/repository"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func main() {
	initLogger()

	log.Info("Starting CTI Dashboard...")

	cfg, err := config.Load()
	if err != nil {
		log.WithError(err).Fatal("Failed to load configuration")
	}

	db, err := database.Connect(database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		Name:     cfg.Database.Name,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
	})
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	router := setupRouter(db, cfg)

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.WithField("port", cfg.Server.Port).Info("Starting HTTP server")
	if err := router.Run(addr); err != nil {
		log.WithError(err).Fatal("Failed to start server")
	}
}

func setupRouter(db *sql.DB, cfg *config.Config) *gin.Engine {
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	router.Use(middleware.ErrorHandler())
	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger())
	router.Use(auth.InitSessionStore(cfg.Session.Secret))

	userRepo := repository.NewUserRepository(db)
	contentRepo := repository.NewContentRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	authHandler := handlers.NewAuthHandler(userRepo)
	contentHandler := handlers.NewContentHandler(contentRepo)
	categoryHandler := handlers.NewCategoryHandler(categoryRepo)

	router.Static("/static", "./web/static")
	router.LoadHTMLGlob("web/templates/*.html")

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"service": "cti-dashboard",
		})
	})

	router.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/login")
	})
	
	router.GET("/login", func(c *gin.Context) {
		c.HTML(200, "login.html", nil)
	})
	
	router.GET("/dashboard", middleware.RequireAuth(), func(c *gin.Context) {
		c.HTML(200, "dashboard.html", nil)
	})
	
	router.GET("/detail/:id", middleware.RequireAuth(), func(c *gin.Context) {
		c.HTML(200, "detail.html", nil)
	})
	
	router.GET("/categories", middleware.RequireAdmin(), func(c *gin.Context) {
		c.HTML(200, "categories.html", nil)
	})

	api := router.Group("/api")
	{
		authRoutes := api.Group("/auth")
		{
			authRoutes.POST("/login", authHandler.Login)
			authRoutes.POST("/logout", authHandler.Logout)
			authRoutes.GET("/session", authHandler.GetSession)
		}

		contentRoutes := api.Group("/contents")
		{
			contentRoutes.GET("", contentHandler.List)
			contentRoutes.GET("/stats", contentHandler.GetStats)
			contentRoutes.GET("/:id", contentHandler.GetByID)
		}

		categoryRoutes := api.Group("/categories")
		{
			categoryRoutes.GET("", categoryHandler.List)
			categoryRoutes.POST("", middleware.RequireAdmin(), categoryHandler.Create)
			categoryRoutes.PUT("/:id", middleware.RequireAdmin(), categoryHandler.Update)
			categoryRoutes.DELETE("/:id", middleware.RequireAdmin(), categoryHandler.Delete)
		}

		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})
	}

	return router
}

func initLogger() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)

	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "WARN":
		log.SetLevel(log.WarnLevel)
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
}
