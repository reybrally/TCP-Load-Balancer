package config

type Config struct {
	Server   ServerConfig    `mapstructure:"server"`
	Backends []BackendConfig `mapstructure:"backends"`
	App      AppConfig       `mapstructure:"app"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

type BackendConfig struct {
	Address string `mapstructure:"address"`
	Port    int    `mapstructure:"port"`
	Weight  int    `mapstructure:"weight"`
}

type AppConfig struct {
	Environment string `mapstructure:"environment"`
	LogLevel    string `mapstructure:"log_level"`
}
