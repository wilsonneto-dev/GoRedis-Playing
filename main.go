package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client *redis.ClusterClient
	ctx    context.Context
}

func NewRedisStore(password string, db int) *RedisStore {
	rdb := redis.NewFailoverClusterClient(&redis.FailoverOptions{
		MasterName: "mymaster",
		SentinelAddrs: []string{
			"sentinel1:26379",
			"sentinel1:26379",
			"sentinel1:26379"},
		Password: password,
		DB:       db,
		RouteRandomly: true,
	})

	return &RedisStore{
		client: rdb,
		ctx:    context.Background(),
	}
}

var store *RedisStore

func (rs *RedisStore) Save(key string, value string) error {
	fmt.Println("Saving key: ", key, " value: ", value)
	return rs.client.Set(rs.ctx, key, value, 0).Err()
}

func (rs *RedisStore) Retrieve(key string) (string, error) {
	fmt.Println("Retrieving key: ", key)
	return rs.client.Get(rs.ctx, key).Result()
}

type KeyValuePayload struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func setHandler(c *gin.Context) {
	var payload KeyValuePayload

	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to bind JSON"})
		return
	}

	key := payload.Key
	value := payload.Value

	err := store.Save(key, value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set value in Redis"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func getHandler(c *gin.Context) {
	key := c.Param("key")
	value, err := store.Retrieve(key)

	if err == redis.Nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Key not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get value from Redis", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"value": value})
}

func main() {
	router := gin.Default()
	store = NewRedisStore("", 0)

	router.POST("/set", setHandler)
	router.GET("/get/:key", getHandler)

	router.Run("0.0.0.0:8080")
}
