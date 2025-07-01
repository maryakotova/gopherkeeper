package config

type Config struct {
	RunAddress  string
	DatabaseURI string
}

func NewConfig() *Config {
	flags := ParseFlags()

	return &Config{
		RunAddress:  flags.RunAddress,
		DatabaseURI: flags.DatabaseURI,
	}
}
