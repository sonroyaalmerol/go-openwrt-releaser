package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sonroyaalmerol/go-openwrt-releaser/internal/run"
)

func main() {
	configPath := flag.String("config", "openwrt-releaser.yaml", "path to config file")
	moduleRoot := flag.String("root", ".", "go module root")
	distDir := flag.String("dist", "", "output directory (default <root>/dist)")
	tag := flag.String("tag", "", "release tag (e.g. v1.2.3); enables github release")
	skipRelease := flag.Bool("skip-release", false, "build and package only; do not publish")
	skipIndex := flag.Bool("skip-index", false, "skip feed index generation")
	flag.Parse()

	if err := run.Run(run.Options{
		ConfigPath:  *configPath,
		ModuleRoot:  *moduleRoot,
		DistDir:     *distDir,
		Tag:         *tag,
		SkipRelease: *skipRelease,
		SkipIndex:   *skipIndex,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
