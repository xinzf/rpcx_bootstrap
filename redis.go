package bootstrap

import (
	"errors"
	"github.com/go-redis/redis"
)

var Redis *_redis
var _redisInit bool

type redisConfig struct {
	Addr string `mapstructure:"addr"`
	Pswd string `mapstructure:"pswd"`
	DB   int    `mapstructure:"db"`
}

type _redis struct {
	client *redis.Client
}

func (this *_redis) init() error {
	if _redisInit {
		return nil
	}

	_redisInit = true
	Redis = &_redis{}
	if Config.Redis.Addr == "" {
		return errors.New("redis config's addr is empty")
	}

	Redis.client = redis.NewClient(&redis.Options{
		Addr:     Config.Redis.Addr,
		Password: Config.Redis.Pswd,
		DB:       Config.Redis.DB,
	})

	_, err := Redis.client.Ping().Result()
	if err != nil {
		return errors.New("redis connect failed,err: " + err.Error())
	}
	return nil
}

func (this *_redis) Client() *redis.Client {
	return this.client
}
