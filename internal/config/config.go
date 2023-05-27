package config

import (
	"flag"
	"log"
	"os"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Hook struct {
		Username string `yaml:"name" env:"T_NAME"`
		Token    string `yaml:"token" env:"T_TOKEN"`
		ChatID   string `yaml:"id" env:"T_ID"`
	} `yaml:"hook"`
	PostgreSQL struct {
		Host     string `yaml:"host" env:"DB_HOST" env-required:"true"`
		Username string `yaml:"username" env:"DB_USER" env-required:"true"`
		Password string `yaml:"password" env:"DB_PASS" env-required:"true"`
		Database string `yaml:"database" env:"DB_DBNAME" env-required:"true"`
		Port     string `yaml:"port" env:"DB_PORT" env-required:"true"`
	} `yaml:"postgresql"`
	HTTP struct {
		IP   string `yaml:"ip" env:"IP"`
		Port int    `yaml:"port" env:"PORT"`
		CORS struct {
			AllowedMethods []string `yaml:"allowed_methods"`
			// TODO add origins
			AllowedOrigins []string `yaml:"allowed_origins"`
			AllowedHeaders []string `yaml:"allowed_headers"`
		} `yaml:"cors"`
	} `yaml:"http"`
	GRPC struct {
		IP   string `yaml:"ip" env:"IP"`
		Port int    `yaml:"port" env:"PORT"`
	} `yaml:"grpc"`

	AppConfig struct {
		LogLevel string `yaml:"log-level" env:"LOG_LEVEL" env-default:"trace"`
	} `yaml:"app"`
}

const (
	EnvConfigPathName  = "CONFIG-PATH"
	FlagConfigPathName = "config"
)

var configPath string
var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		flag.StringVar(&configPath, FlagConfigPathName, ".configs/config.local.yaml", "this is app config file")
		flag.Parse()

		log.Print("config init")

		if configPath == "" {
			configPath = os.Getenv(EnvConfigPathName)
		}

		if configPath == "" {
			log.Fatal("config path is required")
		}

		instance = &Config{}

		if err := cleanenv.ReadConfig(configPath, instance); err != nil {
			helpText := "Service"
			help, _ := cleanenv.GetDescription(instance, &helpText)
			log.Print(help)
			log.Fatal(err)
		}
	})
	return instance
}
