package config

import (
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"log"
)
import _ "github.com/joho/godotenv"

type Value struct {
	EtcdHost           string `env:"ETCD_HOST" envDefault:"localhost"`
	EtcdPort           string `env:"ETCD_PORT" envDefault:"2379"`
	ProxyHost          string `env:"PROXY_HOST" envDefault:"localhost"`
	ProxyPort          string `env:"PROXY_PORT" envDefault:"8080"`
	SelfOnCompletePort string `env:"SELF_ON_COMPLETE_PORT" envDefault:"4040"`
}

func Get() Value {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
	var conf Value
	if err := env.Parse(&conf); err != nil {
		return Value{}
	}
	return conf
}
