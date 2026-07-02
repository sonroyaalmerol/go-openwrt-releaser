package arch

import (
	"testing"
)

func TestResolveAllSubtargets(t *testing.T) {
	for key := range subtargetDB {
		info, err := Resolve(key)
		if err != nil {
			t.Errorf("Resolve(%q): %v", key, err)
			continue
		}
		if info.Supported && info.GOArch == "" {
			t.Errorf("%q: supported but no GOArch", key)
		}
		if info.Supported && info.GOArch != "" && info.GOOS != "linux" {
			t.Errorf("%q: GOOS = %q, want linux", key, info.GOOS)
		}
	}
}

func TestResolveUnsupported(t *testing.T) {
	info, err := Resolve("mpc85xx/p2020")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Supported {
		t.Fatal("mpc85xx should be unsupported")
	}
	if info.UnsupportedReason == "" {
		t.Error("expected unsupported reason")
	}
	if _, err := Resolve("nonexistent/foo"); err == nil {
		t.Error("expected error for unknown target")
	}
}

func TestIPQ806x(t *testing.T) {
	info, err := Resolve("ipq806x/generic")
	if err != nil {
		t.Fatal(err)
	}
	if info.GOArch != "arm" {
		t.Errorf("GOArch = %q, want arm", info.GOArch)
	}
	if info.GOARM != "7" {
		t.Errorf("GOARM = %q, want 7", info.GOARM)
	}
	if info.ArchSuffix != "eabi" {
		t.Errorf("ArchSuffix = %q, want eabi", info.ArchSuffix)
	}
	if info.PKGArch != "arm_cortex-a15_neon-vfpv4" {
		t.Errorf("PKGArch = %q, want arm_cortex-a15_neon-vfpv4", info.PKGArch)
	}
}

func TestX86_64(t *testing.T) {
	info, err := Resolve("x86/64")
	if err != nil {
		t.Fatal(err)
	}
	if info.GOArch != "amd64" {
		t.Errorf("GOArch = %q, want amd64", info.GOArch)
	}
	if info.PKGArch != "x86_64" {
		t.Errorf("PKGArch = %q, want x86_64", info.PKGArch)
	}
}

func TestAarch64(t *testing.T) {
	info, err := Resolve("armsr/armv8")
	if err != nil {
		t.Fatal(err)
	}
	if info.GOArch != "arm64" {
		t.Errorf("GOArch = %q, want arm64", info.GOArch)
	}
	if !info.Is64bit {
		t.Error("should be 64-bit")
	}
}

func TestMipsLE(t *testing.T) {
	info, err := Resolve("ramips/mt7621")
	if err != nil {
		t.Fatal(err)
	}
	if info.GOArch != "mipsle" {
		t.Errorf("GOArch = %q, want mipsle", info.GOArch)
	}
	if info.GOMIPS != "softfloat" {
		t.Errorf("GOMIPS = %q, want softfloat", info.GOMIPS)
	}
}

func TestRISCV64(t *testing.T) {
	info, err := Resolve("starfive/generic")
	if err != nil {
		t.Fatal(err)
	}
	if info.GOArch != "riscv64" {
		t.Errorf("GOArch = %q, want riscv64", info.GOArch)
	}
}

func TestEnvHasGOOS(t *testing.T) {
	info, err := Resolve("ipq806x/generic")
	if err != nil {
		t.Fatal(err)
	}
	env := info.Env()
	found := false
	for _, e := range env {
		if e == "GOOS=linux" {
			found = true
		}
	}
	if !found {
		t.Error("Env() missing GOOS=linux")
	}
}
