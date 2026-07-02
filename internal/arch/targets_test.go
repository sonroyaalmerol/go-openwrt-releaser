package arch

import (
	"fmt"
	"testing"
)

func TestAllSupportedTargetsResolve(t *testing.T) {
	targets := SupportedTargets()
	if len(targets) == 0 {
		t.Fatal("no supported targets returned")
	}
	t.Logf("testing %d supported targets", len(targets))
	for _, target := range targets {
		t.Run(target, func(t *testing.T) {
			info, err := Resolve(target)
			if err != nil {
				t.Fatalf("Resolve: %v", err)
			}
			if !info.Supported {
				t.Fatal("SupportedTargets returned unsupported target")
			}
			if info.GOArch == "" {
				t.Fatal("GOArch is empty")
			}
			if info.GOOS != "linux" {
				t.Errorf("GOOS = %q, want linux", info.GOOS)
			}
			if info.PKGArch == "" {
				t.Error("PKGArch is empty")
			}
			if info.GOArch == "arm" && info.GOARM == "" {
				t.Error("GOARM is empty for arm target")
			}
		})
	}
}

func TestAllUnsupportedTargetsHaveReason(t *testing.T) {
	targets := UnsupportedTargets()
	for _, target := range targets {
		t.Run(target, func(t *testing.T) {
			info, err := Resolve(target)
			if err != nil {
				t.Fatalf("Resolve: %v", err)
			}
			if info.Supported {
				t.Fatal("expected unsupported")
			}
			if info.UnsupportedReason == "" {
				t.Error("UnsupportedReason is empty")
			}
		})
	}
}

func TestNoDuplicateTargets(t *testing.T) {
	targets := SupportedTargets()
	seen := map[string]bool{}
	for _, target := range targets {
		if seen[target] {
			t.Errorf("duplicate target %q in SupportedTargets", target)
		}
		seen[target] = true
	}
}

func TestSupportedTargetsAreSorted(t *testing.T) {
	targets := SupportedTargets()
	for i := 1; i < len(targets); i++ {
		if targets[i-1] > targets[i] {
			t.Errorf("not sorted: %q > %q at index %d", targets[i-1], targets[i], i)
			break
		}
	}
}

func TestRoundTripAllEnvVars(t *testing.T) {
	for _, target := range SupportedTargets() {
		info, err := Resolve(target)
		if err != nil {
			t.Fatalf("Resolve(%s): %v", target, err)
		}
		env := info.Env()
		hasGOOS := false
		hasGOARCH := false
		hasCGO := false
		for _, e := range env {
			if e == "GOOS=linux" {
				hasGOOS = true
			}
			if e == fmt.Sprintf("GOARCH=%s", info.GOArch) {
				hasGOARCH = true
			}
			if e == "CGO_ENABLED=0" {
				hasCGO = true
			}
		}
		if !hasGOOS || !hasGOARCH || !hasCGO {
			t.Errorf("target %s: env missing required vars: %v", target, env)
		}
	}
}
