package target

import (
	"testing"

	"github.com/sonroyaalmerol/go-openwrt-releaser/internal/config"
)

func TestResolveX86_64(t *testing.T) {
	p, err := Resolve(config.Target{OpenWrt: "x86/64"}, config.SDK{})
	if err != nil {
		t.Fatal(err)
	}
	if p.Board != "x86" || p.Subtarget != "64" {
		t.Errorf("Board=%q Subtarget=%q", p.Board, p.Subtarget)
	}
	if p.GOArch != "amd64" {
		t.Errorf("GOArch = %q, want amd64", p.GOArch)
	}
	if p.PKGArch != "x86_64" {
		t.Errorf("PKGArch = %q, want x86_64", p.PKGArch)
	}
}

func TestResolveOverrideGOArch(t *testing.T) {
	p, err := Resolve(config.Target{
		OpenWrt: "x86/64",
		GOArch:  "386",
	}, config.SDK{})
	if err != nil {
		t.Fatal(err)
	}
	if p.GOArch != "386" {
		t.Errorf("GOArch = %q, want 386", p.GOArch)
	}
}

func TestResolveOverridePKGArch(t *testing.T) {
	p, err := Resolve(config.Target{
		OpenWrt: "x86/64",
		PKGArch: "custom_arch",
	}, config.SDK{})
	if err != nil {
		t.Fatal(err)
	}
	if p.PKGArch != "custom_arch" {
		t.Errorf("PKGArch = %q, want custom_arch", p.PKGArch)
	}
}

func TestResolveAutoPKGArchIgnored(t *testing.T) {
	p, err := Resolve(config.Target{
		OpenWrt: "x86/64",
		PKGArch: "auto",
	}, config.SDK{})
	if err != nil {
		t.Fatal(err)
	}
	if p.PKGArch != "x86_64" {
		t.Errorf("PKGArch = %q, want x86_64 (auto should not override)", p.PKGArch)
	}
}

func TestResolveUnknownTarget(t *testing.T) {
	_, err := Resolve(config.Target{OpenWrt: "nonexistent/board"}, config.SDK{})
	if err == nil {
		t.Fatal("expected error for unknown target")
	}
}

func TestResolveUnsupportedTarget(t *testing.T) {
	_, err := Resolve(config.Target{OpenWrt: "mpc85xx/p2020"}, config.SDK{})
	if err == nil {
		t.Fatal("expected error for unsupported target")
	}
}

func TestResolveGOARMOverride(t *testing.T) {
	p, err := Resolve(config.Target{
		OpenWrt: "ipq806x/generic",
		GOARM:   "6",
	}, config.SDK{})
	if err != nil {
		t.Fatal(err)
	}
	if p.GOARM != "6" {
		t.Errorf("GOARM = %q, want 6", p.GOARM)
	}
}

func TestEnvContainsGoVars(t *testing.T) {
	p, err := Resolve(config.Target{OpenWrt: "ipq806x/generic"}, config.SDK{})
	if err != nil {
		t.Fatal(err)
	}
	env := p.Env()
	expect := map[string]string{
		"GOOS=linux":    "linux",
		"GOARCH=arm":    "arm",
		"GOARM=7":       "7",
		"CGO_ENABLED=0": "0",
	}
	for _, e := range env {
		if _, ok := expect[e]; ok {
			delete(expect, e)
		}
	}
	if len(expect) > 0 {
		t.Errorf("Env() missing expected entries: %v", expect)
	}
}

func TestSplitTarget(t *testing.T) {
	board, sub := splitTarget("ipq806x/generic")
	if board != "ipq806x" || sub != "generic" {
		t.Errorf("got %q/%q, want ipq806x/generic", board, sub)
	}
}

func TestSplitTargetNoSubtarget(t *testing.T) {
	board, sub := splitTarget("ath79")
	if board != "ath79" {
		t.Errorf("board = %q, want ath79", board)
	}
	if sub != "generic" {
		t.Errorf("subtarget = %q, want generic", sub)
	}
}

func TestOutputDir(t *testing.T) {
	p, err := Resolve(config.Target{OpenWrt: "armsr/armv8"}, config.SDK{})
	if err != nil {
		t.Fatal(err)
	}
	if p.OutputDir() != "armsr/armv8" {
		t.Errorf("OutputDir = %q", p.OutputDir())
	}
}
