package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("DB_PATH", "")
	t.Setenv("CHECK_INTERVAL_SECONDS", "invalid")

	got := Load()
	if got.DBPath != "./data/orpheus.db" {
		t.Fatalf("DBPath = %q, want default", got.DBPath)
	}
	if got.CheckIntervalSeconds != defaultCheckIntervalSeconds {
		t.Fatalf("CheckIntervalSeconds = %d, want %d", got.CheckIntervalSeconds, defaultCheckIntervalSeconds)
	}
}
