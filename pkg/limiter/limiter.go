package limiter

import (
	"context"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/shareed2k/go_limiter"
)

var Limiter *go_limiter.Limiter

func InitLimter() error {
	redisURI := os.Getenv("LENSQUERY_REDIS_URI")
	option, err := redis.ParseURL(redisURI)
	if err != nil {
		log.Fatal(err)
		return err
	}

	client := redis.NewClient(option)

	cmd := client.Ping(context.Background())
	if cmd.Err() != nil {
		log.Fatal("Redis error: ", redisURI)
		return cmd.Err()
	}
	log.Println("Redis connected", cmd)

	Limiter = go_limiter.NewLimiter(client)
	return nil
}
