package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/michaelboegner/interviewer/internal/server"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load(".env.dev"); err != nil {
			log.Printf("Error loading .env file: %v", err)
		}
	}

	srv := server.NewServer()
	srv.StartServer()
}
