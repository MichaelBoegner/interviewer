package main

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/michaelboegner/interviewer/internal/server"
)

func main() {
	var handler slog.Handler
	if os.Getenv("ENV") == "production" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		_ = godotenv.Load(".env.dev")
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}

	logger := slog.New(handler)

	srv, err := server.NewServer(logger)
	if err != nil {
		logger.Error("server initialization failed", "error", err)
		os.Exit(1)
	}

	logger.Info("starting server...")
	srv.StartServer()
}
