package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type (
	Config struct {
		Env         string            `yaml:"env" env-required:"true" env-default:"local"`
		HTTPServer  HTTPServer        `yaml:"http"`
		AuthService AuthServiceConfig `yaml:"auth-service"`
		DataService DataServiceConfig `yaml:"data-service"`
	}

	HTTPServer struct {
		Host               string        `yaml:"host" env-default:"localhost"`
		Port               int           `yaml:"port" env-default:"8888"`
		WriteTimeout       time.Duration `yaml:"writeTimeout" env-default:"10s"`
		ReadTimeout        time.Duration `yaml:"readTimeout" env-default:"10s"`
		MaxHeaderMegabytes int           `yaml:"maxHeaderMegabytes" env-default:"1"`
	}

	AuthServiceConfig struct {
		Http HTTPConfig `yaml:"http"`
	}

	DataServiceConfig struct {
		Http HTTPConfig `yaml:"http"`
	}

	HTTPConfig struct {
		Schema string `yaml:"schema"`
		Host   string `yaml:"host"`
		Port   int    `yaml:"port"`
	}
)

func InitConfig(configPath string) *Config {
	if configPath == "" {
		log.Fatal("configPath is not set")
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("cannot find config file.. \"%s\"", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config file: \"%s\"", err.Error())
	}

	return &cfg
}
