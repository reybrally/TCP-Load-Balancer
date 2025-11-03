package config

import (
	"fmt"
	cfg "github.com/reybrally/TCP-Load-Balancer/internal/config"

	"github.com/spf13/viper"
)

func Load(configPath string) (*cfg.Config, error) {
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.SetConfigName("config")

	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("app.environment", "development")

	viper.AutomaticEnv()
	viper.SetEnvPrefix("LB")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config cfg.Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}
