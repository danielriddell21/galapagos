package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultRacingValid(t *testing.T) {
	if err := DefaultRacing().Validate(); err != nil {
		t.Fatalf("default config invalid: %v", err)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Racing)
		wantErr bool
	}{
		{"ok", func(c *Racing) {}, false},
		{"tiny population", func(c *Racing) { c.Population = 1 }, true},
		{"bad elite", func(c *Racing) { c.EliteFraction = 1 }, true},
		{"bad rays", func(c *Racing) { c.Rays = 0 }, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := DefaultRacing()
			tt.mutate(&c)
			if err := c.Validate(); (err != nil) != tt.wantErr {
				t.Fatalf("Validate() err=%v, wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadRacingOverridesDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "c.yaml")
	if err := os.WriteFile(path, []byte("population: 50\nseed: 7\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := LoadRacing(path)
	if err != nil {
		t.Fatalf("LoadRacing: %v", err)
	}
	if c.Population != 50 || c.Seed != 7 {
		t.Fatalf("overrides not applied: %+v", c)
	}
	// Unspecified fields keep defaults.
	if c.HiddenSize != DefaultRacing().HiddenSize {
		t.Fatalf("default hidden_size lost: %d", c.HiddenSize)
	}
}
