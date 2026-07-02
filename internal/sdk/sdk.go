package sdk

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sonroyaalmerol/go-openwrt-releaser/internal/arch"
	"github.com/sonroyaalmerol/go-openwrt-releaser/internal/config"
)

const defaultDownloadBase = "https://downloads.openwrt.org/releases"

type Manager struct {
	cacheDir   string
	sdk        config.SDK
	httpClient *http.Client
}

func New(cacheDir string, sdk config.SDK) *Manager {
	return &Manager{
		cacheDir:   cacheDir,
		sdk:        sdk,
		httpClient: &http.Client{},
	}
}

func (m *Manager) APKBinary() (string, error) {
	apkPath := filepath.Join(m.cacheDir, "host", "bin", "apk")
	if info, err := os.Stat(apkPath); err == nil && !info.IsDir() {
		if err := verifyAPK(apkPath); err == nil {
			return apkPath, nil
		}
	}
	if err := m.extractAPKBinary(); err != nil {
		return "", err
	}
	return apkPath, nil
}

func verifyAPK(path string) error {
	cmd := exec.Command(path, "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("apk --version failed: %w", err)
	}
	if !strings.Contains(string(out), "apk-tools") {
		return fmt.Errorf("unexpected apk version output: %s", string(out))
	}
	return nil
}

func (m *Manager) sdkURL(board, subtarget, goArch string) string {
	suffix := archSuffix(goArch)
	libc := m.sdk.Libc
	if libc == "" {
		libc = "musl"
	}
	if subtarget == "" {
		subtarget = "generic"
	}
	base := fmt.Sprintf("%s/%s/targets/%s/%s/openwrt-sdk-%s-%s-%s_gcc-%s_%s",
		defaultDownloadBase, m.sdk.OpenWrtVersion, board, subtarget,
		m.sdk.OpenWrtVersion, board, subtarget, m.sdk.GCCVersion, libc)
	if suffix != "" {
		base += "_" + suffix
	}
	return base + ".Linux-x86_64.tar.zst"
}

func (m *Manager) extractAPKBinary() error {
	board := m.toolchainBoard()
	subtarget := m.toolchainSubtarget()
	goArch := toolchainGoArch(board)
	url := m.sdkURL(board, subtarget, goArch)
	shaURL := fmt.Sprintf("%s/%s/targets/%s/%s/sha256sums",
		defaultDownloadBase, m.sdk.OpenWrtVersion, board, subtarget)

	if err := os.MkdirAll(m.cacheDir, 0o755); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}

	tarPath := filepath.Join(m.cacheDir, "sdk.tar.zst")
	if _, err := os.Stat(tarPath); os.IsNotExist(err) {
		fmt.Printf("  downloading SDK from %s\n", url)
		if err := download(url, tarPath, m.httpClient); err != nil {
			return fmt.Errorf("download SDK: %w", err)
		}
	}

	fmt.Printf("  verifying checksum\n")
	if err := verifySHA256(tarPath, url, shaURL, m.httpClient); err != nil {
		return err
	}

	fmt.Printf("  extracting apk toolchain from SDK\n")
	if err := extractAPKToolchain(tarPath, m.cacheDir); err != nil {
		return fmt.Errorf("extract apk toolchain: %w", err)
	}
	return verifyAPK(filepath.Join(m.cacheDir, "host", "bin", "apk"))
}

func (m *Manager) toolchainBoard() string {
	tt := m.sdk.ToolchainTarget
	for i := 0; i < len(tt); i++ {
		if tt[i] == '/' {
			return tt[:i]
		}
	}
	return tt
}

func (m *Manager) toolchainSubtarget() string {
	tt := m.sdk.ToolchainTarget
	for i := 0; i < len(tt); i++ {
		if tt[i] == '/' {
			return tt[i+1:]
		}
	}
	return "generic"
}

func download(url, dest string, client *http.Client) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d for %s", resp.StatusCode, url)
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

func verifySHA256(tarPath, sdkURL, shaURL string, client *http.Client) error {
	resp, err := client.Get(shaURL)
	if err != nil {
		return fmt.Errorf("download sha256sums: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d for %s", resp.StatusCode, shaURL)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	base := sdkURL[strings.LastIndex(sdkURL, "/")+1:]
	var expected string
	for line := range strings.SplitSeq(string(body), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		name := strings.TrimPrefix(fields[1], "*")
		if name == base {
			expected = fields[0]
			break
		}
	}
	if expected == "" {
		return fmt.Errorf("checksum for %s not found in %s", base, shaURL)
	}
	data, err := os.ReadFile(tarPath)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(data)
	actual := hex.EncodeToString(sum[:])
	if actual != expected {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expected, actual)
	}
	return nil
}

func extractAPKToolchain(tarPath, cacheDir string) error {
	hostDir := filepath.Join(cacheDir, "host")
	if err := os.MkdirAll(hostDir, 0o755); err != nil {
		return err
	}
	cmd := exec.Command("tar", "--zstd", "-xf", tarPath,
		"-C", hostDir,
		"--strip-components", "3",
		"--wildcards",
		"*/staging_dir/host/bin/apk",
		"*/staging_dir/host/bin/.apk.bin",
		"*/staging_dir/host/lib/ld-linux-x86-64.so.2",
		"*/staging_dir/host/lib/libc.so.6",
		"*/staging_dir/host/lib/libpthread.so.0",
		"*/staging_dir/host/lib/librt.so.1",
		"*/staging_dir/host/lib/libm.so.6",
		"*/staging_dir/host/lib/libgcc_s.so.1",
		"*/staging_dir/host/lib/runas.so",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tar: %w: %s", err, string(out))
	}
	if _, err := os.Stat(filepath.Join(hostDir, "bin", "apk")); err != nil {
		return fmt.Errorf("apk wrapper not extracted: %w", err)
	}
	return nil
}

func toolchainGoArch(board string) string {
	if info, err := arch.Resolve(board + "/generic"); err == nil {
		return info.GOArch
	}
	return "amd64"
}

func archSuffix(goArch string) string {
	if goArch == "arm" {
		return "eabi"
	}
	return ""
}
