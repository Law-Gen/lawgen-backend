package config

import "github.com/spf13/viper"

type Config struct {
	Server struct {
		Port int
	}
	Mongo struct {
		URI    string
		DBName string
	}
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigName("config")
	v.AddConfigPath("./configs")
	v.AutomaticEnv()

	_ = v.ReadInConfig()

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Mongo.DBName == "" {
		cfg.Mongo.DBName = "lawgen"
	}
	return cfg, nil
}