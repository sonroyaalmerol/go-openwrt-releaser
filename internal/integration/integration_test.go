//go:build integration

package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/sonroyaalmerol/go-openwrt-releaser/internal/arch"
	"github.com/sonroyaalmerol/go-openwrt-releaser/internal/builder"
	"github.com/sonroyaalmerol/go-openwrt-releaser/internal/config"
	"github.com/sonroyaalmerol/go-openwrt-releaser/internal/packager"
	"github.com/sonroyaalmerol/go-openwrt-releaser/internal/sdk"
	"github.com/sonroyaalmerol/go-openwrt-releaser/internal/target"
)

func TestBuildAndPackageAllTargets(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION") == "" {
		t.Skip("set RUN_INTEGRATION=1 to run integration tests")
	}

	fixture := mustFindFixture(t)
	apkBin := mustGetAPKBinary(t)

	targets := arch.SupportedTargets()
	t.Logf("building and packaging %d targets", len(targets))

	conf := config.Go{
		Module:  "github.com/sonroyaalmerol/go-openwrt-releaser/testdata/crosstest",
		Main:    ".",
		Binary:  "crosstest",
		LDFlags: []string{"-s", "-w"},
	}
	b := builder.New(fixture, conf)

	for _, targetName := range targets {
		t.Run(targetName, func(t *testing.T) {
			plan, err := target.Resolve(config.Target{OpenWrt: targetName}, config.SDK{})
			if err != nil {
				t.Fatalf("Resolve(%s): %v", targetName, err)
			}
			if !plan.Supported {
				t.Skipf("unsupported: %s", plan.UnsupportedReason)
			}

			work := t.TempDir()
			buildOut := filepath.Join(work, "build")

			if _, err := b.Build(plan, buildOut); err != nil {
				t.Fatalf("build %s: %v", targetName, err)
			}

			pkg := config.Package{
				Name:        "crosstest",
				Description: "integration test package",
				License:     "MIT",
				Maintainer:  "ci@example.com",
				URL:         "https://example.com",
				Depends:     []string{"libc"},
				Binary:      true,
				BinaryDest:  "/usr/bin",
			}
			outDir := filepath.Join(work, "out")
			pkgr := packager.New(apkBin, "1.0.0", "crosstest")
			result, err := pkgr.Package(pkg, plan, buildOut, outDir)
			if err != nil {
				t.Fatalf("package %s: %v", targetName, err)
			}

			if _, err := os.Stat(result.ApkPath); err != nil {
				t.Fatalf("apk not created: %v", err)
			}

			validateAPK(t, result.ApkPath, apkBin, plan.PKGArch)
		})
	}
}

func validateAPK(t *testing.T, apkPath, apkBin, expectPKGArch string) {
	t.Helper()
	cmd := exec.Command(apkBin, "adbdump", apkPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("apk adbdump: %v\n%s", err, string(out))
	}
	dump := string(out)

	if !strings.Contains(dump, "name: crosstest") {
		t.Errorf("apk metadata missing name: crosstest")
	}
	if !strings.Contains(dump, "version: 1.0.0") {
		t.Errorf("apk metadata missing version: 1.0.0")
	}
	if !strings.Contains(dump, "arch: "+expectPKGArch) {
		t.Errorf("apk metadata missing or wrong arch\nwant: %s\ndump:\n%s", expectPKGArch, dump)
	}
	if !strings.Contains(dump, "name: usr/bin") {
		t.Errorf("apk does not contain usr/bin directory\n%s", dump)
	}
	if !strings.Contains(dump, "installed-size:") {
		t.Errorf("apk missing installed-size\n%s", dump)
	}
	if !strings.Contains(dump, "data block") {
		t.Errorf("apk has no data blocks (binary not embedded)\n%s", dump)
	}
}

func mustFindFixture(t *testing.T) string {
	t.Helper()
	for _, rel := range []string{
		"../../testdata/crosstest",
		"../../../testdata/crosstest",
	} {
		abs, err := filepath.Abs(rel)
		if err != nil {
			continue
		}
		if st, err := os.Stat(filepath.Join(abs, "go.mod")); err == nil && !st.IsDir() {
			return abs
		}
	}
	t.Fatal("testdata/crosstest fixture not found")
	return ""
}

func mustGetAPKBinary(t *testing.T) string {
	t.Helper()
	if runtime.GOOS != "linux" || runtime.GOARCH != "amd64" {
		t.Skip("apk host tool requires linux/amd64")
	}
	cacheDir, err := filepath.Abs(".integration-sdk-cache")
	if err != nil {
		t.Fatal(err)
	}
	mgr := sdk.New(cacheDir, config.SDK{
		OpenWrtVersion:  "25.12.4",
		GCCVersion:      "14.3.0",
		Libc:            "musl",
		ToolchainTarget: "x86/64",
	})
	apkBin, err := mgr.APKBinary()
	if err != nil {
		t.Fatalf("get apk binary: %v", err)
	}
	t.Logf("apk binary: %s", apkBin)
	return apkBin
}
