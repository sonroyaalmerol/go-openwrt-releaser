package builder

import (
	"testing"

	"github.com/sonroyaalmerol/go-openwrt-releaser/internal/config"
)

func TestJoinFlags(t *testing.T) {
	got := joinFlags([]string{"-s", "-w", "-X main.version=1.0"})
	want := "-s -w -X main.version=1.0"
	if got != want {
		t.Errorf("joinFlags = %q, want %q", got, want)
	}
}

func TestJoinFlagsEmpty(t *testing.T) {
	if joinFlags(nil) != "" {
		t.Error("expected empty string for nil")
	}
}

func TestJoinFlagsSingle(t *testing.T) {
	got := joinFlags([]string{"-s"})
	if got != "-s" {
		t.Errorf("joinFlags = %q, want -s", got)
	}
}

func TestNewBuilder(t *testing.T) {
	g := config.Go{
		Module: "example.com/app",
		Main:   "./cmd/app",
		Binary: "app",
	}
	b := New("/root", g)
	if b.moduleRoot != "/root" {
		t.Errorf("moduleRoot = %q", b.moduleRoot)
	}
	if b.go_.Binary != "app" {
		t.Errorf("Binary = %q", b.go_.Binary)
	}
}
