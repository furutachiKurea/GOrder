package redis

import (
	"fmt"
	"time"

	"github.com/furutachiKurea/gorder/common/handler/factory"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

const (
	confName      = "redis"
	localSupplier = "local"
)

var (
	singleton = factory.NewSingleton(supplier)
)

func Init() {
	conf := viper.GetStringMap(confName)

	for supplyName := range conf {
		Client(supplyName)
	}
}

// LocalClient 获取本地 redis.Client 实例
func LocalClient() *redis.Client {
	return Client(localSupplier)
}

// Client 从单例工厂中获取名为 name 的 redis.Client 实例,
// 这里的目的是将所有的 redis 实例注册到单例工厂
func Client(name string) *redis.Client {
	return singleton.Get(name).(*redis.Client)
}

// supplier 从给定的 key (local, etc.) 中加载对应的配置，并返回 redis.Client 实例
func supplier(key string) any {
	confKey := confName + "." + key
	type Section struct {
		IP           string        `mapstructure:"ip"`
		Port         string        `mapstructure:"port"`
		PoolSize     int           `mapstructure:"pool_size"`
		MaxConn      int           `mapstructure:"max_conn"`
		CoonTimeout  time.Duration `mapstructure:"coon_timeout"`
		ReadTimeout  time.Duration `mapstructure:"read_timeout"`
		WriteTimeout time.Duration `mapstructure:"write_timeout"`
	}

	var c Section
	if err := viper.UnmarshalKey(confKey, &c); err != nil {
		panic(err)
	}

	return redis.NewClient(&redis.Options{
		Network:         "tcp",
		Addr:            fmt.Sprintf("%s:%s", c.IP, c.Port),
		PoolSize:        c.PoolSize,
		MaxActiveConns:  c.MaxConn,
		ConnMaxLifetime: c.CoonTimeout * time.Millisecond,
		ReadTimeout:     c.ReadTimeout * time.Millisecond,
		WriteTimeout:    c.WriteTimeout * time.Millisecond,
	})
}
