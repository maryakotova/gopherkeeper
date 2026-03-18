package config

type Config struct {
	RunAddress  string
	DatabaseURI string
	JWTKey      string
}

func NewConfig() *Config {
	flags := ParseFlags()

	return &Config{
		RunAddress:  flags.RunAddress,
		DatabaseURI: flags.DatabaseURI,
		JWTKey:      flags.JWTKey,
	}
}
