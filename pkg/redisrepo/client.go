package redisrepo

import (
	"context"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
)

var redisClient *redis.Client

func initialiseRedis() *redis.Client {
	conn := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_CONNECTION_STRING"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	pong, err := conn.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal("Redis Connection Failed", err)
	}

	log.Println("Redis Connected.", "Ping", pong)

	redisClient = conn
	return redisClient
}
