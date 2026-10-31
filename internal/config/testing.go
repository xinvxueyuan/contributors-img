package config

import (
	"contrib.rocks/pkg/env"
)

func NewTestConfig() *Config {
	return &Config{
		Port:            "3333",
		Env:             env.EnvDevelopment,
		GitHubAuthToken: "test",
	}
}
