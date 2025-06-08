package config

import (
	"os"
	"testing"
)

// TestGetConfig_Singleton verifies that GetConfig always returns the same instance
// and that it loads configuration (implicitly, by checking a required field and a default).
func TestGetConfig_SingletonAndDefaults(t *testing.T) {
	// Store original environment variables to restore them later
	originalApiUrl := os.Getenv("CARD_CRAFT_AI_API_URL")
	originalPort := os.Getenv("PORT")

	// Set a required environment variable to prevent log.Fatal during config loading.
	testApiUrl := "http://dummyurl.com/for/testing"
	os.Setenv("CARD_CRAFT_AI_API_URL", testApiUrl)

	// Unset PORT to test default value loading
	os.Unsetenv("PORT")

	// It's tricky to reset sync.Once for tests. This test assumes GetConfig
	// will be effectively called "for the first time" with these env vars,
	// or that subsequent calls don't change the already loaded config.
	// For robust testing of different env var scenarios, config loading might need refactoring
	// (e.g., pass env map to a loading function instead of global os.Getenv).

	cfg1 := GetConfig()
	if cfg1 == nil {
		t.Fatal("GetConfig() returned nil on first call")
	}

	cfg2 := GetConfig()
	if cfg1 != cfg2 {
		t.Errorf("GetConfig() returned different instances (%p vs %p), expected singleton behavior", cfg1, cfg2)
	}

	if cfg1.CardCraftAi.URL != testApiUrl {
		t.Errorf("Expected CardCraftAi.APIURL to be '%s', got '%s'", testApiUrl, cfg1.CardCraftAi.URL)
	}

	// Check default port (env-default:"8080")
	expectedDefaultPort := 8080
	if cfg1.HTTP.Port != expectedDefaultPort {
		t.Errorf("Expected default HTTP.Port %d, got %d", expectedDefaultPort, cfg1.HTTP.Port)
	}

	// Restore original environment variables
	if originalApiUrl == "" {
		os.Unsetenv("CARD_CRAFT_AI_API_URL")
	} else {
		os.Setenv("CARD_CRAFT_AI_API_URL", originalApiUrl)
	}
	if originalPort == "" {
		os.Unsetenv("PORT")
	} else {
		os.Setenv("PORT", originalPort)
	}
}

// Note: Testing the log.Fatal paths in GetConfig is complex in standard unit tests
// as log.Fatal exits the program. This typically requires specific test harnesses
// or refactoring the code to make errors returnable.
// The current test focuses on the successful path and singleton property.
