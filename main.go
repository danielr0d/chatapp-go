package main

import (
	"log"
	"os"

	"chatapp/pkg/httpserver"
	"chatapp/pkg/redisrepo"

	"github.com/joho/godotenv"
)

func main() {
	// .env is optional, env vars may come from the environment itself
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	redisrepo.InitialiseRedis()
	redisrepo.CreateFetchChatBetweenIndex()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	httpserver.StartHTTPServer(":" + port)
}
