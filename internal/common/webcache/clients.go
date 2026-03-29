package webcache

import "github.com/armylong/go-library/service/redis"

var RedisWslClient *redis.Client
var RedisClient *redis.Client

func init() {
	RedisWslClient = redis.GetClient(`wsl`)
	RedisClient = redis.GetClient(`default`)
}
