package app

import (
	"os"
	"testing"
)

func TestNewApp(t *testing.T) {
	originalApiUrl := os.Getenv("CARD_CRAFT_AI_API_URL")
	err := os.Setenv("CARD_CRAFT_AI_API_URL", "http://localhost:8000/test-newapp")
	if err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}
	defer func() {
		if originalApiUrl == "" {
			os.Unsetenv("CARD_CRAFT_AI_API_URL")
		} else {
			os.Setenv("CARD_CRAFT_AI_API_URL", originalApiUrl)
		}
	}()

	// Reset config singleton for this test if other tests have run.
	// This is tricky. For this test, we assume config will load with the above URL.
	// If config.GetConfig() was already called, it might have a different URL.
	// This highlights a challenge with global singletons in tests.

	appInstance := NewApp()
	if appInstance == nil {
		t.Fatal("NewApp() returned nil")
	}
	if appInstance.mux == nil {
		t.Error("NewApp().mux is nil")
	}
	if appInstance.cfg == nil {
		t.Error("NewApp().cfg is nil")
	}
	if appInstance.httpClient == nil {
		t.Error("NewApp().httpClient is nil")
	}
	if appInstance.cardCraftAiAPIURL != appInstance.cfg.CardCraftAi.URL {
		t.Errorf("Expected appInstance.cardCraftAiAPIURL (%s) to match cfg.CardCraftAi.APIURL (%s)",
			appInstance.cardCraftAiAPIURL, appInstance.cfg.CardCraftAi.URL)
	}
}
