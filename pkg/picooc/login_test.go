package picooc

import (
	"os"
	"testing"
)

func TestOnlineGetAllWeights(t *testing.T) {
	username := os.Getenv("PICOOC_USERNAME")
	password := os.Getenv("PICOOC_PASSWORD")
	if username == "" || password == "" {
		t.Skip("set PICOOC_USERNAME and PICOOC_PASSWORD")
	}

	client := NewClient()
	if err := client.Login(username, password); err != nil {
		t.Fatalf("login: %v", err)
	}

	t.Logf("login user id: %s", client.userID)
	weights, err := client.GetAllWeights()
	if err != nil {
		t.Fatalf("get weights: %v", err)
	}
	if len(weights) == 0 {
		t.Fatal("expected at least one weight")
	}
	t.Logf("loaded %d weights, latest: %s %.1fkg", len(weights), weights[0].Date.Format("2006-01-02"), weights[0].Weight)
}
