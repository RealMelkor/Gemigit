package config

import "github.com/kkyr/fig"

var Cfg Config

type Config struct {
	Gemigit struct {
		Https                 bool   `validate:"required"`
		Port                  int    `validate:"required"`
		Name                  string `validate:"required"`
		Domain                string `validate:"required"`
		AllowRegistration     bool   `validate:"required"`
		Database              string `validate:"required"`
		MaxAttemptsForIP      int    `validate:"required"`
		MaxAttemptsForAccount int    `validate:"required"`
		AuthTimeout           int    `validate:"required"`
	}
}

func LoadConfig() error {
	return fig.Load(&Cfg,
		fig.File("config.yaml"),
		fig.Dirs(".", "/etc/gemigit", "/usr/local/etc/gemigit"),
	)
}
