package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/viranchils96/simple-text-searcher/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Path        string
	Query       string
	Shards      int
	MaxResults  int
	ProfilePort int
	Timeout     time.Duration
}

func main() {
	cfg := loadConfig()
	logger := initLogger()
	defer logger.Sync()

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	setupProfiling(cfg.ProfilePort, logger)
	registerShutdownHandler(cancel, logger)

	// Document streaming
	docsCh, errCh := utils.StreamDocuments(ctx, cfg.Path)
	var docs []utils.Document

	// Index initialization
	index := utils.NewIndex(cfg.Shards)
	batch := make([]utils.Document, 0, 1000)

	// Processing pipeline
	logger.Info("Starting processing pipeline")
	start := time.Now()

Processing:
	for {
		select {
		case doc, ok := <-docsCh:
			if !ok {
				break Processing
			}
			docs = append(docs, doc)
			batch = append(batch, doc)

			if len(batch) >= 1000 {
				index.Add(batch)
				batch = batch[:0]
			}

		case err := <-errCh:
			if err != nil {
				logger.Fatal("Pipeline failed", zap.Error(err))
			}

		case <-ctx.Done():
			logger.Warn("Processing cancelled")
			return
		}
	}

	// Process remaining documents
	if len(batch) > 0 {
		index.Add(batch)
	}

	logger.Info("Indexing completed",
		zap.Int("documents", len(docs)),
		zap.Duration("duration", time.Since(start)),
	)

	// Execute search
	start = time.Now()
	results := index.Search(cfg.Query, cfg.MaxResults, docs)
	logger.Info("Search completed",
		zap.Int("results", len(results)),
		zap.Duration("duration", time.Since(start)),
	)

	printResults(results, cfg.MaxResults)
}

func loadConfig() Config {
	var cfg Config
	flag.StringVar(&cfg.Path, "path", "enwiki-latest-abstract1.xml.gz", "Document path")
	flag.StringVar(&cfg.Query, "query", "Small wild cat", "Search query")
	flag.IntVar(&cfg.Shards, "shards", 8, "Index shards")
	flag.IntVar(&cfg.MaxResults, "max", 10, "Max results")
	flag.IntVar(&cfg.ProfilePort, "profile", 6060, "Profile port")
	flag.DurationVar(&cfg.Timeout, "timeout", 5*time.Minute, "Timeout")
	flag.Parse()
	return cfg
}

func initLogger() *zap.Logger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, _ := config.Build()
	return logger
}

func setupProfiling(port int, logger *zap.Logger) {
	go func() {
		addr := fmt.Sprintf("localhost:%d", port)
		logger.Info("Profiling enabled", zap.String("address", addr))
		http.ListenAndServe(addr, nil)
	}()
}

func registerShutdownHandler(cancel context.CancelFunc, logger *zap.Logger) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		logger.Info("Shutting down", zap.String("signal", sig.String()))
		cancel()
	}()
}

func printResults(results []utils.SearchResult, max int) {
	fmt.Printf("\nTop %d results:\n", max)
	for i, result := range results {
		if i >= max {
			break
		}
		fmt.Printf("%d. [Score: %.3f] ID: %d\n%s\n\n",
			i+1, result.Score, result.ID, result.Text)
	}
}
