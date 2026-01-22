package scheduler

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
)

type CollectionFunc func() error

type Scheduler struct {
	interval       time.Duration
	collectionFunc CollectionFunc
	stopChan       chan struct{}
	doneChan       chan struct{}
}

func NewScheduler(interval time.Duration, collectionFunc CollectionFunc) *Scheduler {
	return &Scheduler{
		interval:       interval,
		collectionFunc: collectionFunc,
		stopChan:       make(chan struct{}),
		doneChan:       make(chan struct{}),
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	log.WithField("interval", s.interval).Info("Starting collection scheduler")

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	s.runCollection()

	for {
		select {
		case <-ticker.C:
			s.runCollection()
		case <-s.stopChan:
			log.Info("Scheduler stop signal received")
			close(s.doneChan)
			return
		case <-ctx.Done():
			log.Info("Context cancelled, stopping scheduler")
			close(s.doneChan)
			return
		}
	}
}

func (s *Scheduler) Stop() {
	log.Info("Stopping scheduler...")
	close(s.stopChan)
	<-s.doneChan
	log.Info("Scheduler stopped")
}

func (s *Scheduler) runCollection() {
	start := time.Now()
	log.Info("Starting collection cycle")

	err := s.collectionFunc()
	dur := time.Since(start)
	
	if err != nil {
		log.WithFields(log.Fields{
			"error":    err.Error(),
			"duration": dur,
		}).Error("Collection cycle failed")
	} else {
		log.WithField("duration", dur).Info("Collection cycle completed successfully")
	}
}
