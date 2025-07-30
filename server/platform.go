package platform

import "github.com/juanMaAV92/go-utils/env"

func GetBasicServerConfig(serverName string) *BasicConfig {
	return &BasicConfig{
		Port:          env.GetEnv(env.Port),
		GracefullTime: env.GetEnvAsDurationWithDefault(env.GracefulTime, "10s"),
		Environment:   env.GetEnviroment(),
		ServerName:    serverName,
	}
}
