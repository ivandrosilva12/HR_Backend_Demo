package redisdb

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var (
	Client *redis.Client
	Ctx    = context.Background()
)

func InitRedis() {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	password := os.Getenv("REDIS_PASSWORD") // vazio por padrão
	db := 0

	Client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	if _, err := Client.Ping(Ctx).Result(); err != nil {
		log.Fatalf("Erro ao conectar ao Redis: %v", err)
	}

	log.Println("✅ Redis conectado com sucesso")
}
