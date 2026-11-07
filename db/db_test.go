package db

import (
	"strings"
	"testing"
)

func TestGetClientRefusesRealDatabaseInDevelopment(t *testing.T) {
	cases := []struct {
		name, emulator, project string
	}{
		{"nothing configured", "", ""},
		{"no emulator", "", "demo-carrot"},
		{"real project id", "127.0.0.1:8200", "r-j-magenta-carrot-42069"},
		{"default project id", "127.0.0.1:8200", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("APP_ENV", "development")
			t.Setenv("FIRESTORE_EMULATOR_HOST", tc.emulator)
			t.Setenv("FIRESTORE_PROJECT_ID", tc.project)

			if DevSandbox() {
				t.Error("DevSandbox() = true, want false")
			}
			client, _, err := GetClient()
			if err == nil {
				client.Close()
				t.Fatal("GetClient() succeeded, want refusal")
			}
			if !strings.Contains(err.Error(), "refusing") {
				t.Errorf("GetClient() error = %v, want refusal", err)
			}
		})
	}
}

func TestDevSandboxRequiresDevelopment(t *testing.T) {
	t.Setenv("FIRESTORE_EMULATOR_HOST", "127.0.0.1:8200")
	t.Setenv("FIRESTORE_PROJECT_ID", "demo-carrot")

	t.Setenv("APP_ENV", "")
	if DevSandbox() {
		t.Error("DevSandbox() = true without APP_ENV=development")
	}
	t.Setenv("APP_ENV", "development")
	if !DevSandbox() {
		t.Error("DevSandbox() = false with emulator and demo project")
	}
}
