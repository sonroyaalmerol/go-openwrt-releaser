package releaser

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Releaser struct {
	gh      *GitHub
	workDir string
}

type GitHub struct {
	Owner string
	Name  string
}

func New(gh *GitHub, workDir string) *Releaser {
	return &Releaser{gh: gh, workDir: workDir}
}

func (r *Releaser) CreateRelease(tag, title, notes string, files []string) error {
	if r.gh == nil || r.gh.Owner == "" {
		fmt.Println("  no github config; skipping release upload")
		return nil
	}
	repo := r.gh.Owner + "/" + r.gh.Name
	args := []string{"release", "create", tag}
	if title != "" {
		args = append(args, "--title", title)
	}
	if notes != "" {
		args = append(args, "--notes", notes)
	}
	args = append(args, "--repo", repo)
	args = append(args, files...)

	cmd := exec.Command("gh", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = r.workDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh release create: %w", err)
	}
	return nil
}

func (r *Releaser) PrepareFeed(feedDir, signKeyEnv string) error {
	if err := os.MkdirAll(feedDir, 0o755); err != nil {
		return err
	}
	if signKeyEnv != "" {
		key := os.Getenv(signKeyEnv)
		if key == "" {
			fmt.Printf("  warning: env %s not set; feed will be unsigned\n", signKeyEnv)
		} else {
			keyPath := filepath.Join(feedDir, "private-key.pem")
			if err := os.WriteFile(keyPath, []byte(key), 0o600); err != nil {
				return err
			}
			cmd := exec.Command("openssl", "ec", "-in", keyPath, "-pubout",
				"-out", filepath.Join(feedDir, "feed-public.pem"))
			if out, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("openssl extract pubkey: %w: %s", err, string(out))
			}
		}
	}
	return nil
}

func (r *Releaser) DeployPages(srcDir string) error {
	if _, err := os.Stat(srcDir); err != nil {
		return fmt.Errorf("feed dir missing: %w", err)
	}
	fmt.Println("  note: GitHub Pages deployment requires the upload-pages-artifact action;")
	fmt.Println("        the feed artifact is ready at", srcDir)
	return nil
}

func CollectApks(dir string) ([]string, error) {
	entries, err := filepath.Glob(filepath.Join(dir, "**/*.apk"))
	if err != nil {
		return nil, err
	}
	top, err := filepath.Glob(filepath.Join(dir, "*.apk"))
	if err != nil {
		return nil, err
	}
	all := append(top, entries...)
	if len(all) == 0 {
		return nil, fmt.Errorf("no apk files found under %s", dir)
	}
	out := make([]string, 0, len(all))
	for _, p := range all {
		clean := filepath.Clean(p)
		if !strings.HasSuffix(clean, ".apk") {
			continue
		}
		out = append(out, clean)
	}
	return out, nil
}
