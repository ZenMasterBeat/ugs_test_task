package buildings

import (
	"ugc_test_task/src/pg"
)

type Config struct {
	pgConfig pg.Config
}

func NewConfig(pgConfig pg.Config) Config {
	return Config{
		pgConfig: pgConfig,
	}
}

func (c Config) Validate() error {
	return c.pgConfig.Validate()
}