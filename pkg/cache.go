package pkg

import (
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
)

var (
	Rdb *redis.Client
	Ctx = context.Background()
)

func InitRedis(environment string) {
	if environment == "PROD" {
		redisAddr := os.Getenv("REDIS_URL")
		if redisAddr == "" {
			panic("REDIS_URL não foi definido corretamente. Verifique as variáveis de ambiente.")
		}

		Rdb = redis.NewClient(&redis.Options{
			Addr: redisAddr,
		})

		_, err := Rdb.Ping(Ctx).Result()
		if err != nil {
			panic("Não foi possível conectar ao Redis: " + err.Error())
		}

		fmt.Println("Conectado ao Redis em:", redisAddr)
	}
	log.Println("redis pass...")
}
