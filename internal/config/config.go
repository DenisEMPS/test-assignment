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
	Port string `yaml:"port" env:"PORT" env-required:"true"`
}

type Postgres struct {
	Username string `yaml:"username" env:"POSTGRES_USER" env-required:"true"`
	Password string `yaml:"password" env:"POSTGRES_PASSWORD" env-required:"true"`
	Host     string `yaml:"host" env:"POSTGRES_HOST" env-required:"true"`
	Port     string `yaml:"port" env:"POSTGRES_PORT" env-required:"true"`
	DBname   string `yaml:"dbname" env:"POSTGRES_DB" env-required:"true"`
	SSLmode  string `yaml:"sslmode" env-default:"disable"`
}

type JWT struct {
	AccessTokenTTL  time.Duration `yaml:"access_ttl" env:"ACCESS_TOKEN_TTL" env-required:"true"`
	RefreshTokenTTL time.Duration `yaml:"refresh_ttl" env:"REFRESH_TOKEN_TTL" env-required:"true"`
	Secret          string        `yaml:"secret_key" env:"SECRET_KEY" env-required:"true"`
}

func MustLoad() *Config {
	var path string

	flag.StringVar(&path, "config", "", "path to config file")
	flag.Parse()

	if path == "" {
		path = os.Getenv("CONFIG_PATH")
	}

	return MustLoadByPath(path)
}

func MustLoadByPath(path string) *Config {
	var cfg Config

	if path == "" {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			panic("failed to read env vars" + err.Error())
		}

		return &cfg
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exists")
	}

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("failed to read config" + err.Error())
	}

	return &cfg
}
