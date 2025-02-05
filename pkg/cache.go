package pkg

import (
	"os"

	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
)

var (
	Rdb *redis.Client
	Ctx = context.Background()
)

func InitRedis() {
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "3.238.112.146:6379"
	}

	Rdb = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	_, err := Rdb.Ping(Ctx).Result()
	if err != nil {
		panic("Não foi possível conectar ao Redis: " + err.Error())
	}
}
