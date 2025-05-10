package redis

import(
	"github.com/redis/go-redis/v9"
)

func InitRedis() *redis.Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	return redisClient
}
