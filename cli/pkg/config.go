package pkg

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	APIURL string `mapstructure:"api_url"`
	Token  string `mapstructure:"token"`
}

func Load() Config {
	cfg := Config{
		APIURL: os.Getenv("KUBEROUTE_API_URL"),
		Token:  os.Getenv("KUBEROUTE_TOKEN"),
	}
	if cfg.APIURL == "" {
		cfg.APIURL = viper.GetString("api_url")
	}
	if cfg.APIURL == "" {
		cfg.APIURL = "http://localhost:8080"
	}
	if cfg.Token == "" {
		cfg.Token = viper.GetString("token")
	}
	return cfg
}

func Save(cfg Config) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".kuberoute")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	path := filepath.Join(dir, "config.yaml")
	content := fmt.Sprintf("api_url: %q\ntoken: %q\n", cfg.APIURL, cfg.Token)
	return os.WriteFile(path, []byte(content), 0o600)
}

func RequireToken(cfg Config) error {
	if cfg.Token == "" {
		return fmt.Errorf("not logged in; run kr login")
	}
	return nil
}
