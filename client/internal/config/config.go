package config

type Config struct {
	RunAddress  string
	DatabaseURI string
	TokenPath   string
}

func NewConfig() *Config {
	flags := ParseFlags()

	return &Config{
		RunAddress:  flags.RunAddress,
		DatabaseURI: flags.DatabaseURI,
		TokenPath:   flags.TokenPath,
	}
}
