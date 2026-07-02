package indexer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Indexer struct {
	apkBin string
}

func New(apkBin string) *Indexer {
	return &Indexer{apkBin: apkBin}
}

func (i *Indexer) Index(apkDir, signKeyPath string) (string, error) {
	if err := os.MkdirAll(apkDir, 0o755); err != nil {
		return "", err
	}
	indexPath := filepath.Join(apkDir, "packages.adb")
	args := []string{"mkndx", "--allow-untrusted", "--output", indexPath}
	if signKeyPath != "" {
		args = []string{"mkndx", "--allow-untrusted", "--sign-key", signKeyPath, "--output", indexPath}
	}
	entries, err := filepath.Glob(filepath.Join(apkDir, "*.apk"))
	if err != nil {
		return "", err
	}
	if len(entries) == 0 {
		return "", fmt.Errorf("no .apk files found in %s", apkDir)
	}
	args = append(args, entries...)

	cmd := exec.Command(i.apkBin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("apk mkndx: %w", err)
	}
	return indexPath, nil
}
