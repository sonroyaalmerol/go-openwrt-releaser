package packager

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/sonroyaalmerol/go-openwrt-releaser/internal/config"
	"github.com/sonroyaalmerol/go-openwrt-releaser/internal/target"
)

func TestInstallFileSingle(t *testing.T) {
	srcDir := t.TempDir()
	src := filepath.Join(srcDir, "init.sh")
	if err := os.WriteFile(src, []byte("#!/bin/sh\necho hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	idir := t.TempDir()

	err := installFile(config.FileEntry{
		Src:  src,
		Dest: "/etc/init.d/myapp",
		Mode: "0755",
	}, idir)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(idir, "etc", "init.d", "myapp"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "#!/bin/sh\necho hi\n" {
		t.Errorf("content mismatch: %q", string(data))
	}
}

func TestInstallFileDirRecursive(t *testing.T) {
	srcDir := t.TempDir()
	sub := filepath.Join(srcDir, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "b.txt"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	idir := t.TempDir()

	err := installFile(config.FileEntry{
		Src:  srcDir,
		Dest: "/usr/share/myapp",
	}, idir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(idir, "usr/share/myapp/a.txt")); err != nil {
		t.Errorf("a.txt not copied: %v", err)
	}
	if _, err := os.Stat(filepath.Join(idir, "usr/share/myapp/sub/b.txt")); err != nil {
		t.Errorf("sub/b.txt not copied: %v", err)
	}
}

func TestInstallFileInvalidMode(t *testing.T) {
	src := filepath.Join(t.TempDir(), "f.txt")
	if err := os.WriteFile(src, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := installFile(config.FileEntry{
		Src:  src,
		Dest: "/f",
		Mode: "not-a-mode",
	}, t.TempDir())
	if err == nil {
		t.Fatal("expected error for invalid mode")
	}
}

func TestInstallFileMissingSrc(t *testing.T) {
	err := installFile(config.FileEntry{
		Src:  "/nonexistent/file",
		Dest: "/f",
	}, t.TempDir())
	if err == nil {
		t.Fatal("expected error for missing src")
	}
}

func TestCopyFile(t *testing.T) {
	src := filepath.Join(t.TempDir(), "src.txt")
	if err := os.WriteFile(src, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	dstDir := t.TempDir()
	dst := filepath.Join(dstDir, "nested", "out.txt")
	if err := copyFile(src, dst, 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Errorf("content = %q", string(data))
	}
}

func TestInstallBinary(t *testing.T) {
	buildOut := t.TempDir()
	binPath := filepath.Join(buildOut, "myapp")
	if err := os.WriteFile(binPath, []byte("BINARY"), 0o755); err != nil {
		t.Fatal(err)
	}
	idir := t.TempDir()
	if err := installBinary(buildOut, "myapp", "/usr/bin", idir); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(idir, "usr/bin/myapp")
	if _, err := os.Stat(dst); err != nil {
		t.Errorf("binary not installed: %v", err)
	}
}

func TestInstallBinaryFallback(t *testing.T) {
	buildOut := t.TempDir()
	binPath := filepath.Join(buildOut, "something")
	if err := os.WriteFile(binPath, []byte("BIN"), 0o755); err != nil {
		t.Fatal(err)
	}
	idir := t.TempDir()
	if err := installBinary(buildOut, "nonexistent-name", "/usr/bin", idir); err != nil {
		t.Fatal(err)
	}
	matches, _ := filepath.Glob(filepath.Join(idir, "usr/bin/*"))
	if len(matches) != 1 {
		t.Fatalf("expected 1 binary, got %d", len(matches))
	}
}

func TestInstallBinaryNotFound(t *testing.T) {
	buildOut := t.TempDir()
	err := installBinary(buildOut, "x", "/usr/bin", t.TempDir())
	if err == nil {
		t.Fatal("expected error for missing binary")
	}
}

func TestWriteList(t *testing.T) {
	idir := t.TempDir()
	for _, p := range []string{"usr/bin/app", "etc/config/app", "lib/apk/packages/app.list"} {
		full := filepath.Join(idir, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	metaDir := filepath.Join(idir, "lib/apk/packages")
	if err := writeList(idir, metaDir, "app"); err != nil {
		t.Fatal(err)
	}
	listData, err := os.ReadFile(filepath.Join(metaDir, "app.list"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(listData)), "\n")
	expect := []string{
		"/etc/config/app",
		"/lib/apk/packages/app.list",
		"/usr/bin/app",
	}
	sort.Strings(lines)
	if len(lines) != len(expect) {
		t.Fatalf("got %d lines, want %d: %v", len(lines), len(expect), lines)
	}
	for i := range expect {
		if lines[i] != expect[i] {
			t.Errorf("line %d = %q, want %q", i, lines[i], expect[i])
		}
	}
}

func TestFileSHA256(t *testing.T) {
	path := filepath.Join(t.TempDir(), "f")
	if err := os.WriteFile(path, []byte("hello world"), 0o644); err != nil {
		t.Fatal(err)
	}
	sum, err := fileSHA256(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	if sum != want {
		t.Errorf("sha256 = %q, want %q", sum, want)
	}
}

func TestEscapeDesc(t *testing.T) {
	got := escapeDesc("line1\nline2\nline3")
	if got != "line1 line2 line3" {
		t.Errorf("escapeDesc = %q", got)
	}
}

func TestHasConfFiles(t *testing.T) {
	pkg := config.Package{
		Files: []config.FileEntry{
			{Src: "a", Dest: "/a"},
			{Src: "b", Dest: "/b", Conf: true},
		},
	}
	if !hasConfFiles(pkg) {
		t.Error("expected true")
	}
	pkg2 := config.Package{
		Files: []config.FileEntry{{Src: "a", Dest: "/a"}},
	}
	if hasConfFiles(pkg2) {
		t.Error("expected false")
	}
}

func TestPackageFull(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skip on windows")
	}
	apkBin := findAPKBinary(t)
	if apkBin == "" {
		t.Skip("apk binary not available")
	}

	root := t.TempDir()
	binOut := filepath.Join(root, "build")
	if err := os.MkdirAll(binOut, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(binOut, "myapp"), []byte("FAKE_BINARY"), 0o755); err != nil {
		t.Fatal(err)
	}

	initFile := filepath.Join(root, "init.sh")
	if err := os.WriteFile(initFile, []byte("#!/bin/sh\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	plan := &target.Plan{
		OpenWrt: "x86/64",
		PKGArch: "x86_64",
		GOArch:  "amd64",
		GOOS:    "linux",
	}
	pkg := config.Package{
		Name:        "myapp",
		Description: "test package",
		License:     "MIT",
		Maintainer:  "test@example.com",
		URL:         "https://example.com",
		Binary:      true,
		BinaryDest:  "/usr/bin",
		Depends:     []string{"libc"},
		Files: []config.FileEntry{
			{Src: initFile, Dest: "/etc/init.d/myapp", Mode: "0755"},
		},
		Conffiles: []string{"/etc/config/myapp"},
	}

	outDir := filepath.Join(root, "out")
	pkgr := New(apkBin, "1.0.0", "myapp")
	result, err := pkgr.Package(pkg, plan, binOut, outDir)
	if err != nil {
		t.Fatal(err)
	}
	if result.ApkPath == "" {
		t.Fatal("empty ApkPath")
	}
	if _, err := os.Stat(result.ApkPath); err != nil {
		t.Fatalf("apk not created: %v", err)
	}
	if result.PKGArch != "x86_64" {
		t.Errorf("PKGArch = %q", result.PKGArch)
	}
}

func findAPKBinary(t *testing.T) string {
	t.Helper()
	for _, p := range []string{
		"/usr/bin/apk",
		"/usr/local/bin/apk",
		"../multicast-relay/.owrt-sdk-cache/host/bin/apk",
	} {
		abs, err := filepath.Abs(p)
		if err != nil {
			continue
		}
		if info, err := os.Stat(abs); err == nil && !info.IsDir() {
			return abs
		}
	}
	return ""
}
