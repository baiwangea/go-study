package conf

// Config holds the application configuration.
type Config struct {
	App AppConfig `yaml:"app"`
}

// AppConfig holds application specific settings.
type AppConfig struct {
	Port int `yaml:"port"`
}

// DefaultConfig returns a default configuration.
func DefaultConfig() *Config {
	return &Config{
		App: AppConfig{
			Port: 1323,
		},
	}
}
