// Package config loads the settings from config/local.yaml.
//
// JAVA EQUIVALENT: application.yml + a @ConfigurationProperties class.
package config

import (
	"flag"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv" // reads YAML files (and env vars) into a struct
)

// HTTPServer holds the web server settings (the http_server: block in YAML).
type HTTPServer struct {
	// `yaml:"address"`        -> fill this field from the "address" key
	// `env-required:"true"`   -> cleanenv fails if the value is missing
	Addr string `yaml:"address" env-required:"true"`
}

// env-default:"production"

// Config mirrors the whole YAML file.
type Config struct {
	// `env:"ENV"` = the ENV environment variable can override the YAML value.
	Env         string `yaml:"env" env:"ENV" env-required:"true"`
	StoragePath string `yaml:"storage_path" env-required:"true"`

	// EMBEDDING: a field with a type but NO NAME. Its fields are "promoted",
	// so you can write cfg.Addr instead of cfg.HTTPServer.Addr.
	// (A bit like inheritance in Java, but it is really composition.)
	HTTPServer `yaml:"http_server"`
}

// MustLoad finds the config file, reads it, and returns the settings.
//
// The "Must" prefix is a Go convention: this function does not return an
// error, it STOPS THE PROGRAM (log.Fatal) if something is wrong. That is OK
// here because the app can't run without its config.
func MustLoad() *Config {
	var configPath string

	// 1st choice: the CONFIG_PATH environment variable.
	configPath = os.Getenv("CONFIG_PATH")

	if configPath == "" {
		// 2nd choice: a command-line flag:  go run ./cmd/students-api -config config/local.yaml
		// flag.String returns a POINTER; its value is filled in by flag.Parse().
		flags := flag.String("config", "", "path to the configuration file")
		flag.Parse()

		configPath = *flags // * = read the value the pointer points to

		if configPath == "" {
			log.Fatal("Config path is not set")
		}
	}

	// os.Stat asks for file info; os.IsNotExist checks for "file not found".
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	// &cfg = pass the address, so cleanenv can fill in the struct.
	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		log.Fatalf("can not read config file: %s", err.Error())
	}

	return &cfg
}
