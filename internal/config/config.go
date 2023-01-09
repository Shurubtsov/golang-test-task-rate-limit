package config

import (
	"log"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	BoundDuration int    `env:"BOUND"`
	BlockDuration int    `env:"BLOCK"`
	RequestsLimit uint16 `env:"LIMIT"`
}

var once sync.Once

var instance *Config

func GetConfig() *Config {
	once.Do(func() {
		log.Println("Getting configuration")
		instance = &Config{}
		if err := cleanenv.ReadEnv(instance); err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			log.Fatal(help, err)
		}

		log.Println("Configure: ", instance)
	})

	return instance
}
