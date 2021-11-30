package main

import "github.com/kkyr/fig"

var cfg Config

type Config struct {
	Gemigit struct {
		Https             bool   `validate:"required"`
		Port              int    `validate:"required"`
		Name              string `validate:"required"`
		Domain            string `validate:"required"`
		AllowRegistration bool   `validate:"required"`
		Database          string `validate:"required"`
	}
}

func loadConfig() error {
	return fig.Load(&cfg,
		fig.File("config.yaml"),
		fig.Dirs(".", "/etc/gemigit", "/usr/local/etc/gemigit"),
	)
}
