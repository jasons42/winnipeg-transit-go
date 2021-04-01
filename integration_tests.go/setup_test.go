package integration_tests

import (
	"os"
	"testing"

	transit "github.com/jasons42/winnipeg-transit-go"
)

func client(t *testing.T) *transit.Client {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		t.Fatal("integration tests require setting API_KEY")
	}
	c := transit.NewClient(apiKey)
	return c
}
