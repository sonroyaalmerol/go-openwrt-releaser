package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "openwrt-releaser.yaml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

const minimalConfig = `
project_name: testapp
version: 1.0.0
go:
  module: example.com/testapp
targets:
  - openwrt: x86/64
packages:
  - name: testapp
    description: "test"
`

func TestLoadMinimal(t *testing.T) {
	f, err := Load(writeTempConfig(t, minimalConfig))
	if err != nil {
		t.Fatal(err)
	}
	if f.ProjectName != "testapp" {
		t.Errorf("ProjectName = %q", f.ProjectName)
	}
	if f.Go.Main != "." {
		t.Errorf("expected Main default '.', got %q", f.Go.Main)
	}
	if f.Go.Binary != "testapp" {
		t.Errorf("expected Binary default 'testapp', got %q", f.Go.Binary)
	}
}

func TestDefaults(t *testing.T) {
	f, err := Load(writeTempConfig(t, minimalConfig))
	if err != nil {
		t.Fatal(err)
	}
	if f.SDK.OpenWrtVersion != "25.12.4" {
		t.Errorf("OpenWrtVersion default = %q, want 25.12.4", f.SDK.OpenWrtVersion)
	}
	if f.SDK.GCCVersion != "14.3.0" {
		t.Errorf("GCCVersion default = %q, want 14.3.0", f.SDK.GCCVersion)
	}
	if f.SDK.Libc != "musl" {
		t.Errorf("Libc default = %q, want musl", f.SDK.Libc)
	}
	if f.SDK.ToolchainTarget != "x86/64" {
		t.Errorf("ToolchainTarget default = %q, want x86/64", f.SDK.ToolchainTarget)
	}
}

func TestMissingProjectName(t *testing.T) {
	_, err := Load(writeTempConfig(t, `
go:
  module: example.com/x
targets:
  - openwrt: x86/64
packages:
  - name: x
`))
	if err == nil {
		t.Fatal("expected error for missing project_name")
	}
}

func TestMissingGoModule(t *testing.T) {
	_, err := Load(writeTempConfig(t, `
project_name: x
targets:
  - openwrt: x86/64
packages:
  - name: x
`))
	if err == nil {
		t.Fatal("expected error for missing go.module")
	}
}

func TestNoTargets(t *testing.T) {
	_, err := Load(writeTempConfig(t, `
project_name: x
go:
  module: example.com/x
packages:
  - name: x
`))
	if err == nil {
		t.Fatal("expected error for no targets")
	}
}

func TestNoPackages(t *testing.T) {
	_, err := Load(writeTempConfig(t, `
project_name: x
go:
  module: example.com/x
targets:
  - openwrt: x86/64
`))
	if err == nil {
		t.Fatal("expected error for no packages")
	}
}

func TestInvalidGoArch(t *testing.T) {
	_, err := Load(writeTempConfig(t, `
project_name: x
go:
  module: example.com/x
targets:
  - openwrt: x86/64
    goarch: "INVALID!"
packages:
  - name: x
`))
	if err == nil {
		t.Fatal("expected error for invalid goarch")
	}
}

func TestFileNotExist(t *testing.T) {
	_, err := Load("/nonexistent/path/to/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestInvalidYAML(t *testing.T) {
	_, err := Load(writeTempConfig(t, `
project_name: x
  bad: indentation
`))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestPath(t *testing.T) {
	p := writeTempConfig(t, minimalConfig)
	f, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if f.Path() != p {
		t.Errorf("Path() = %q, want %q", f.Path(), p)
	}
}
