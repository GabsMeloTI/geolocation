package pkg

import (
	"os"

	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
)

var (
	Rdb       *redis.Client
	Ctx       = context.Background()
	SaveRedis bool
)

func InitRedis(environment string, saveRedis bool) {
	SaveRedis = saveRedis
	if environment == "PROD" {
		redisAddr := os.Getenv("REDIS_URL")
		if redisAddr == "" {
			panic("REDIS_URL não foi definido corretamente. Verifique as variáveis de ambiente.")
		}

		Rdb = redis.NewClient(&redis.Options{
			Addr:     redisAddr,
			Password: os.Getenv("REDIS_PASSWORD"),
		})

		_, err := Rdb.Ping(Ctx).Result()
		if err != nil {
			panic("Não foi possível conectar ao Redis: " + err.Error())
		}

		return
	}
}

func CacheGet(ctx context.Context, key string) (string, error) {
	if !SaveRedis || Rdb == nil {
		return "", redis.Nil
	}
	return Rdb.Get(ctx, key).Result()
}

func CacheSet(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if !SaveRedis || Rdb == nil {
		return nil
	}
	return Rdb.Set(ctx, key, value, expiration).Err()
}
