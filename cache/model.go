package cache

import (
	"github.com/juanMaAV92/go-utils/env"
	"github.com/juanMaAV92/go-utils/log"
)

type Cache struct {
	serviceCache string
	instance     Redis
	logger       log.Logger
}

type CacheConfig struct {
	Host       string
	Port       string
	ServerName string
}

func GetCacheConfig(serverName string) CacheConfig {
	return CacheConfig{
		Host:       env.GetEnv(env.CacheHost),
		Port:       env.GetEnv(env.CachePort),
		ServerName: serverName,
	}
}
