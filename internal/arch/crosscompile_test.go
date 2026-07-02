//go:build crosscompile

package arch_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sonroyaalmerol/go-openwrt-releaser/internal/arch"
)

func TestCrossCompileAllTargets(t *testing.T) {
	if os.Getenv("RUN_CROSS_COMPILE_TEST") == "" {
		t.Skip("set RUN_CROSS_COMPILE_TEST=1 to run cross-compilation tests")
	}
	targets := arch.SupportedTargets()
	t.Logf("cross-compiling %d targets", len(targets))

	fixture := mustFindFixture(t)

	for _, targetName := range targets {
		t.Run(targetName, func(t *testing.T) {
			info, err := arch.Resolve(targetName)
			if err != nil {
				t.Fatalf("Resolve(%s): %v", targetName, err)
			}
			if !info.Supported {
				t.Skipf("unsupported: %s", info.UnsupportedReason)
			}

			outDir := t.TempDir()
			outPath := filepath.Join(outDir, "testbin")

			args := []string{"build", "-o", outPath, "."}
			cmd := exec.Command("go", args...)
			cmd.Dir = fixture
			cmd.Env = append(os.Environ(), info.Env()...)

			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("go build %s (%s): %v\n%s",
					targetName, strings.Join(info.Env(), " "), err, string(out))
			}

			stat, err := os.Stat(outPath)
			if err != nil {
				t.Fatalf("output binary not created: %v", err)
			}
			if stat.Size() < 1024 {
				t.Errorf("binary suspiciously small: %d bytes", stat.Size())
			}

			t.Logf("OK: %s (%d bytes, GOARCH=%s)", targetName, stat.Size(), info.GOArch)
		})
	}
}

func mustFindFixture(t *testing.T) string {
	t.Helper()
	candidates := []string{
		"../../testdata/crosstest",
		"../../../testdata/crosstest",
		"./testdata/crosstest",
	}
	for _, c := range candidates {
		if _, err := os.Stat(filepath.Join(c, "go.mod")); err == nil {
			abs, err := filepath.Abs(c)
			if err != nil {
				t.Fatal(err)
			}
			return abs
		}
	}
	t.Fatal("crosstest fixture not found")
	return ""
}
