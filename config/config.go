package config

import "github.com/kkyr/fig"

var Cfg Config

type Config struct {
	Title		string `validate:"required"`
	Database struct {
		Type		string `validate:"required"`
		Url		string `validate:"required"`
	}
	Gemini struct {
		Certificate	string `validate:"required"`
		Key		string `validate:"required"`
		Address		string `validate:"required"`
		Port		string `validate:"required"`
		Templates	string `validate:"required"`
	}
	Git struct {
		Https	bool
		Domain	string	`validate:"required"`
		Remote struct {
			Enabled bool
			Url	string
			Address string
			Key	string
		}
		Address string  `validate:"required"`
		Port	int	`validate:"required"`
		Key	string
		Public	bool
	}
	Ldap struct {
		Enabled		bool
		Url		string
		Attribute	string
		Binding		string
	}
	Users struct {
		Registration	bool
	}
	Protection struct {
		Ip		int    `validate:"required"`
		Account		int    `validate:"required"`
		Registration	int    `validate:"required"`
		Reset		int    `validate:"required"`
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
