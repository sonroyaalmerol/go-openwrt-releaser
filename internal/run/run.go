package run

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sonroyaalmerol/go-openwrt-releaser/internal/builder"
	"github.com/sonroyaalmerol/go-openwrt-releaser/internal/config"
	"github.com/sonroyaalmerol/go-openwrt-releaser/internal/indexer"
	"github.com/sonroyaalmerol/go-openwrt-releaser/internal/packager"
	"github.com/sonroyaalmerol/go-openwrt-releaser/internal/releaser"
	"github.com/sonroyaalmerol/go-openwrt-releaser/internal/sdk"
	"github.com/sonroyaalmerol/go-openwrt-releaser/internal/target"
)

type Options struct {
	ConfigPath  string
	ModuleRoot  string
	DistDir     string
	Tag         string
	SkipRelease bool
	SkipIndex   bool
}

func Run(opts Options) error {
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return err
	}

	version := cfg.Version
	if opts.Tag != "" {
		version = stripV(opts.Tag)
	}
	if version == "" {
		return fmt.Errorf("version required (config version or --tag)")
	}

	distDir := opts.DistDir
	if distDir == "" {
		distDir = filepath.Join(opts.ModuleRoot, "dist")
	}
	if err := os.MkdirAll(distDir, 0o755); err != nil {
		return err
	}
	cacheDir := filepath.Join(opts.ModuleRoot, ".owrt-sdk-cache")

	fmt.Printf("→ resolving SDK apk tool\n")
	mgr := sdk.New(cacheDir, cfg.SDK)
	apkBin, err := mgr.APKBinary()
	if err != nil {
		return fmt.Errorf("sdk setup: %w", err)
	}

	type uniqueArch struct {
		plan    *target.Plan
		wantBin bool
	}
	seen := map[string]*uniqueArch{}
	var uniquePlans []*uniqueArch

	for _, t := range cfg.Targets {
		if t.Skip {
			continue
		}
		plan, err := target.Resolve(t, cfg.SDK)
		if err != nil {
			return fmt.Errorf("target %s: %w", t.OpenWrt, err)
		}
		key := plan.PKGArch
		if _, ok := seen[key]; ok {
			fmt.Printf("→ target %s shares PKGArch %s, skipping\n", t.OpenWrt, key)
			continue
		}
		wn := false
		for _, pkg := range cfg.Packages {
			if pkg.Binary {
				wn = true
				break
			}
		}
		ua := &uniqueArch{plan: plan, wantBin: wn}
		seen[key] = ua
		uniquePlans = append(uniquePlans, ua)
	}
	fmt.Printf("→ %d unique architectures\n", len(uniquePlans))

	cfg.Go.Version = version
	b := builder.New(opts.ModuleRoot, cfg.Go)
	var allApks []string

	archAllDone := false

	for _, ua := range uniquePlans {
		plan := ua.plan
		pkgDir := filepath.Join(distDir, plan.PKGArch)
		buildOut := filepath.Join(pkgDir, "build")

		if ua.wantBin {
			start := time.Now()
			fmt.Printf("\n→ %s (%s)\n", plan.PKGArch, plan.GOArch)
			if _, err := b.Build(plan, buildOut); err != nil {
				return err
			}
			fmt.Printf("  built in %s\n", time.Since(start).Truncate(time.Millisecond))
		}

		pkgr := packager.New(apkBin, version, cfg.ProjectName)
		for _, pkg := range cfg.Packages {
			if pkg.ArchAll && archAllDone {
				continue
			}
			if pkg.Binary && !ua.wantBin {
				continue
			}
			start := time.Now()
			fmt.Printf("  packaging %s\n", pkg.Name)
			pr, err := pkgr.Package(pkg, plan, buildOut, pkgDir)
			if err != nil {
				return err
			}
			fmt.Printf("    %s in %s\n", filepath.Base(pr.ApkPath), time.Since(start).Truncate(time.Millisecond))
			allApks = append(allApks, pr.ApkPath)
			if pkg.ArchAll {
				archAllDone = true
			}
		}
	}

	if !opts.SkipIndex && len(allApks) > 0 {
		fmt.Printf("\n=== indexing feed ===\n")
		feedDir := filepath.Join(distDir, "feed")
		if err := os.MkdirAll(feedDir, 0o755); err != nil {
			return err
		}
		for _, apk := range allApks {
			dst := filepath.Join(feedDir, filepath.Base(apk))
			if err := copyFile(apk, dst); err != nil {
				return err
			}
		}
		var signKey string
		if cfg.Release.SignKeyEnv != "" {
			candidate := filepath.Join(feedDir, "private-key.pem")
			if _, err := os.Stat(candidate); err == nil {
				signKey = candidate
			}
		}
		idx := indexer.New(apkBin)
		if _, err := idx.Index(feedDir, signKey); err != nil {
			return err
		}
		fmt.Printf("  feed index written\n")
	}

	if opts.SkipRelease {
		fmt.Printf("\n→ skipping release (per flag)\n")
		return nil
	}
	if cfg.Release.GitHub == nil {
		fmt.Printf("\n→ no release.github config; skipping publish\n")
		return nil
	}

	fmt.Printf("\n=== releasing ===\n")
	rel := releaser.New(&releaser.GitHub{Owner: cfg.Release.GitHub.Owner, Name: cfg.Release.GitHub.Name}, opts.ModuleRoot)
	if opts.Tag == "" {
		fmt.Printf("  no tag; skipping github release upload\n")
		return nil
	}
	if err := rel.CreateRelease(opts.Tag, version, releaseNotes(version), allApks); err != nil {
		return err
	}
	return nil
}

func stripV(s string) string {
	for len(s) > 0 && (s[0] == 'v' || s[0] == 'V') {
		s = s[1:]
	}
	return s
}

func releaseNotes(version string) string {
	return fmt.Sprintf("OpenWrt packages for %s.", version)
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}
