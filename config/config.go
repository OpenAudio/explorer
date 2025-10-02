package config

type Config struct {
	Environment string
	PgURL       string
	NodeURL     string
}

func NewConfig() *Config {
	return &Config{
		Environment: "",
		PgURL:       "",
		NodeURL:     "",
	}
}
