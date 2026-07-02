package builder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sonroyaalmerol/go-openwrt-releaser/internal/config"
	"github.com/sonroyaalmerol/go-openwrt-releaser/internal/target"
)

type Result struct {
	BinaryPath string
	GOArch     string
}

type Builder struct {
	moduleRoot string
	go_        config.Go
}

func New(moduleRoot string, g config.Go) *Builder {
	return &Builder{moduleRoot: moduleRoot, go_: g}
}

func (b *Builder) Build(plan *target.Plan, outDir string) (*Result, error) {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, err
	}
	binaryName := b.go_.Binary
	if binaryName == "" {
		binaryName = "main"
	}
	outPath := filepath.Join(outDir, binaryName)

	args := []string{"build"}
	for _, t := range b.go_.Tags {
		args = append(args, "-tags", t)
	}
	if len(b.go_.LDFlags) > 0 {
		args = append(args, "-ldflags", joinFlags(b.go_.LDFlags))
	}
	args = append(args, "-o", outPath, b.go_.Main)

	cmd := exec.Command("go", args...)
	cmd.Dir = b.moduleRoot
	cmd.Env = append(os.Environ(), plan.Env()...)
	if b.go_.CGO {
		cmd.Env = append(cmd.Env, "CGO_ENABLED=1")
	}
	for k, v := range b.go_.Env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("go build (%s/%s): %w\n%s", plan.GOOS, plan.GOArch, err, string(out))
	}
	return &Result{BinaryPath: outPath, GOArch: plan.GOArch}, nil
}

func joinFlags(flags []string) string {
	var out strings.Builder
	for i, f := range flags {
		if i > 0 {
			out.WriteString(" ")
		}
		out.WriteString(f)
	}
	return out.String()
}
