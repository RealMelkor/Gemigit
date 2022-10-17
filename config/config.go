package config

import "github.com/kkyr/fig"

var Cfg Config

type Config struct {
	Title		string `validate:"required"`
	Database	string `validate:"required"`
	Certificate	string `validate:"required"`
	Key		string `validate:"required"`
	Git struct {
		Https	bool	`validate:"required"`
		Domain	string	`validate:"required"`
		Port	int	`validate:"required"`
	}
	Ldap struct {
		Enabled		bool	//`validate:"required"`
		Url		string	`validate:"required"`
		Attribute	string	`validate:"required"`
		Binding		string	`validate:"required"`
	}
	Users struct {
		Registration	bool   `validate:"required"`
		Session		int    `validate:"required"`
	}
	Protection struct {
		Ip	int    `validate:"required"`
		Account	int    `validate:"required"`
		Reset	int    `validate:"required"`
	}
}

func LoadConfig() error {
	err := fig.Load(
		&Cfg,
		fig.File("config.yaml"),
		fig.Dirs(".", "/etc/gemigit", "/usr/local/etc/gemigit"),
	)
	if err == nil && Cfg.Ldap.Enabled {
		Cfg.Users.Registration = false
	}
	return err
}
