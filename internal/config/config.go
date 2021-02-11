package config

import (
	"sync"

	"github.com/caarlos0/env/v6"
)

// Config is a struct.
type Config struct {
	Port                         int    `env:"PORT" envDefault:"55555"`
	MongoDbURI                   string `env:"MONGO_DB_URI" envDefault:"mongodb://mongodb:27017"`
	MongoDb                      string `env:"MONGO_DB" envDefault:"servicedb"`
	MongoPriceCollection         string `env:"MONGO_PRICE_COLLECTION" envDefault:"pricelist"`
	MongoClientConnectionTimeout int    `env:"MONGO_CIENT_CONNECTION_TEMEOUT" envDefault:"10"` /// seconds
}

var (
	once sync.Once // nolint:gochecknoglobals
	cfg  *Config   // nolint:gochecknoglobals
)

// GetConfig возвращает конфигурацию заданную env-переменными.
func GetConfig() (*Config, error) {
	var err error

	once.Do(func() {
		cfg = &Config{}
		err = env.Parse(cfg)
	})

	if err != nil {
		return nil, err
	}

	return cfg, nil
}
