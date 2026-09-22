package configs

import (
	"auth/internal/repositories"
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/ajaka/nexus-shared/outbox"
)

func InitializeKafkaOutbox(ctx context.Context, repo *repositories.Repository, logger *slog.Logger) {
	outbox, err := outbox.NewEngine(repo, "stuff", outbox.Config{
		PollInterval: 100 * time.Millisecond,
		BatchSize:    20,
		Topic:        "Some topic",
	}, logger)
	if err != nil {
		logger.Error("Failed to init outbox engine", "err", err.Error())
		os.Exit(1)
	}
	outbox.Start(ctx)
}
