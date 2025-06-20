package config

import (
	"log"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	IsDev   bool `env:"IS_DEV" env-default:"false"`
	IsDebug bool `env:"IS_DEBUG" env-default:"false"`

	CardCraftAi struct {
		URL  string `env:"CARD_CRAFT_AI_API_URL" env-required:"true"`
		Port int    `env:"CARD_CRAFT_AI_PORT" env-default:"8080"`
	}
	WB struct {
		GetCardListMaxAttempts int `env:"WB_GET_CARD_LIST_MAX_ATTEMPTS" env-default:"3"`
	}
	TokenCounter struct {
		APIURL string `env:"TOKEN_COUNTER_API_URL" env-required:"true"`
		Port   int    `env:"TOKEN_COUNTER_PORT" env-default:"8080"`
	}
	HTTP struct {
		Port int `env:"PORT" env-default:"8080"`
	}
	FileStorage struct {
		UploadDir  string `env:"FILE_STORAGE_UPLOAD_DIR" env-default:"./uploads"`
		TTLMinutes int    `env:"FILE_STORAGE_TTL_MINUTES" env-default:"10"`
		BaseURL    string `env:"FILE_STORAGE_BASE_URL" env-required:"true"`
	}
	PostgreSQL struct {
		Database string `env:"PG_DATABASE" env-required:"true"`
		Username string `env:"PG_USER" env-required:"true"`
		Password string `env:"PG_PASSWORD" env-required:"true"`
		Host     string `env:"PG_HOST" env-default:"localhost"`
		Port     string `env:"PG_PORT" env-default:"5432"`
	}
	Tinkoff struct {
		SecretKey        string `env:"TINKOFF_SECRET_KEY" env-required:"true"`
		TerminalKey      string `env:"TINKOFF_TERMINAL_KEY" env-required:"true"`
		TelegramBotToken string `env:"TELEGRAM_BOT_TOKEN" env-default:""`
	}
}

// Singleton: Config should only ever be created once.
var instance *Config

// Once is an object that will perform exactly one action.
var once sync.Once

// GetConfig returns pointer to Config.
func GetConfig() *Config {
	// Calls the function if and only if Do is being called for the first time for this instance of Once
	once.Do(func() {
		log.Print("collecting config...")

		// Config initialization
		instance = &Config{}

		// Read environment variables into the instance of the Config
		if err := cleanenv.ReadEnv(instance); err != nil {
			// If something is wrong
			helpText := "Environment variables error:"
			// Returns a description of environment variables with a custom header - helpText
			help, err := cleanenv.GetDescription(instance, &helpText)
			if err != nil {
				log.Fatal(err)
			}
			log.Print(help)
			log.Printf("%+v\n", instance)

			log.Fatal(err)
		}
	})
	return instance
}
