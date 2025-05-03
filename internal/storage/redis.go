package storage

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
)

var Rdb *redis.Client

// Redis 연결을 초기화하는 함수
func InitRedis() {
	Rdb = redis.NewClient(&redis.Options{
		Addr:     "redis:6379", // Docker Compose에서 서비스 이름(redis)을 호스트로 사용
		Password: "",           // 기본 비밀번호가 없으면 빈 문자열
		DB:       0,            // 기본 DB(0번)
	})

	// Redis 연결 테스트
	ctx := context.Background()
	_, err := Rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	} else {
		fmt.Println("Connected to Redis successfully!")
	}
}

func CloseRedis() {
	err := Rdb.Close()
	if err != nil {
		log.Fatalf("Could not close Redis connection: %v", err)
	}
}

func SetRoomChat(roomID string, message interface{}) error {
	ctx := context.Background()
	return Rdb.RPush(ctx, roomID, message).Err()
}

func GetRoomChat(roomID string) ([]string, error) {
	ctx := context.Background()
	return Rdb.LRange(ctx, roomID, 0, -1).Result()
}
