package main

import (
	"database/sql"
	"log/slog"
	"os"
	"strings"

	"github.com/Danztee/shazam-build/internal/env"
	_ "github.com/go-sql-driver/mysql"
)

func main() {

	dsn := env.GetString("DATABASE_URL", "root:@tcp(127.0.0.1:3306)/shazam?parseTime=true")
	if strings.HasPrefix(dsn, "postgres://") {
		dsn = "root:@tcp(127.0.0.1:3306)/shazam?parseTime=true"
	}

	cfg := Config{
		addr: ":" + env.GetString("PORT", "8080"),
		db: dbConfig{
			dsn: dsn,
		},
	}

	// logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Database
	db, err := sql.Open("mysql", cfg.db.dsn)
	if err != nil {
		logger.Error("failed to open database", "error", err, "dsn", cfg.db.dsn)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Error("failed to connect to MySQL database", "error", err, "dsn", cfg.db.dsn)
		os.Exit(1)
	}

	logger.Info("Connected to MySQL database")

	api := application{
		config: cfg,
		db:     db,
	}

	if err := api.run(api.mount()); err != nil {
		slog.Error("server failed to start", "error", err)
		os.Exit(1)
	}

}
