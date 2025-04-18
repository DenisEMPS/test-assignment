package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env      string   `yaml:"env" env-default:"local"`
	Server   Server   `yaml:"server"`
	Postgres Postgres `yaml:"postgres"`
	JWT      JWT      `yaml:"token"`
}

type Server struct {
	Port string `yaml:"port" env-required:"true"`
}

type Postgres struct {
	Username string `yaml:"username" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
	Host     string `yaml:"host" env-required:"true"`
	Port     string `yaml:"port" env-required:"true"`
	DBname   string `yaml:"dbname" env-required:"true"`
	SSLmode  string `yaml:"sslmode" env-default:"disable"`
}

type JWT struct {
	AccessTokenTTL  time.Duration `yaml:"access_ttl" env-required:"true"`
	RefreshTokenTTL time.Duration `yaml:"refresh_ttl" env-required:"true"`
	Secret          string        `env:"SECRET_KEY"`
}

func MustLoad() *Config {
	var cfgPath string

	flag.StringVar(&cfgPath, "config", "", "path to config file")
	flag.Parse()

	if cfgPath == "" {
		cfgPath = os.Getenv("CONFIG_PATH")
	}

	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		panic("config file does not exists")
	}

	var cfg Config

	if err := cleanenv.ReadConfig(cfgPath, &cfg); err != nil {
		panic("failed to read config")
	}

	return &cfg
}
