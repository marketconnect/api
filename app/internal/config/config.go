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
		APIURL string `env:"CARD_CRAFT_AI_API_URL" env-required:"true"`
	}
	WB struct {
		APIKey                 string `env:"WB_API_KEY" env-default:""`
		GetCardListMaxAttempts int    `env:"WB_GET_CARD_LIST_MAX_ATTEMPTS" env-default:"3"`
	}
	HTTP struct {
		Port int `env:"PORT" env-default:"8080"`
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
