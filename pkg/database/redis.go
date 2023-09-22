package database

import (
	"context"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client

func ConnectRedis() error {
	redisURI := os.Getenv("LENSQUERY_REDIS_URI")
	option, err := redis.ParseURL(redisURI)
	if err != nil {
		log.Fatal(err)
		return err
	}

	RedisClient = redis.NewClient(option)

	cmd := RedisClient.Ping(context.Background())
	if cmd.Err() != nil {
		log.Fatal("Redis error: ", redisURI)
		return cmd.Err()
	}
	log.Println("Redis connected", cmd)

	return nil
}
