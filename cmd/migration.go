package main

import (
	"context"
	"os"

	"github.com/irabeny89/gbege/internal/logger"
	"github.com/irabeny89/gbege/migrate"
	"github.com/irabeny89/gosqlitex"
)

func handleMigrations(ctx context.Context, db *gosqlitex.DbClient) error {
	migDir, ok := os.LookupEnv("MIG_DIR")
	if !ok {
		migDir = "./migrations"
	}
	sep, ok := os.LookupEnv("MIG_SEP")
	if !ok {
		sep = "_"
	}
	logger.Log.Info("Running migrations", "dir", migDir, "sep", sep)
	return migrate.RunMigrations(ctx, migDir, sep, db)
}
