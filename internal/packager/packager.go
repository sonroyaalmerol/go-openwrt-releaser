package packager

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sonroyaalmerol/go-openwrt-releaser/internal/config"
	"github.com/sonroyaalmerol/go-openwrt-releaser/internal/target"
)

type Packager struct {
	apkBin  string
	version string
	project string
}

func New(apkBin, version, project string) *Packager {
	return &Packager{apkBin: apkBin, version: version, project: project}
}

type PackResult struct {
	ApkPath string
	PKGArch string
}

func (p *Packager) Package(pkg config.Package, plan *target.Plan, buildOut string, outDir string) (*PackResult, error) {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, err
	}

	workDir, err := os.MkdirTemp("", "owrt-pkg-")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(workDir) }()

	idir := filepath.Join(workDir, "files")
	metaDir := filepath.Join(idir, "lib", "apk", "packages")
	if err := os.MkdirAll(metaDir, 0o755); err != nil {
		return nil, err
	}

	if pkg.Binary {
		binDest := pkg.BinaryDest
		if binDest == "" {
			binDest = "/usr/bin"
		}
		if err := installBinary(buildOut, pkg.Name, binDest, idir); err != nil {
			return nil, err
		}
	}

	for _, fe := range pkg.Files {
		if err := installFile(fe, idir); err != nil {
			return nil, fmt.Errorf("install %s -> %s: %w", fe.Src, fe.Dest, err)
		}
	}

	pkgBaseName := pkg.Name
	if pkg.ABIVersion != "" {
		pkgBaseName += "_" + pkg.ABIVersion
	}

	if len(pkg.Conffiles) > 0 || hasConfFiles(pkg) {
		conffilePath := filepath.Join(metaDir, pkgBaseName+".conffiles")
		sortedConf := append([]string{}, pkg.Conffiles...)
		for _, fe := range pkg.Files {
			if fe.Conf {
				sortedConf = append(sortedConf, fe.Dest)
			}
		}
		sort.Strings(sortedConf)
		data := strings.Join(sortedConf, "\n") + "\n"
		if err := os.WriteFile(conffilePath, []byte(data), 0o644); err != nil {
			return nil, err
		}
		staticPath := filepath.Join(metaDir, pkgBaseName+".conffiles_static")
		var sb strings.Builder
		for _, c := range sortedConf {
			rel := strings.TrimPrefix(c, "/")
			full := filepath.Join(idir, rel)
			sum, err := fileSHA256(full)
			if err != nil {
				return nil, fmt.Errorf("hash conffile %s: %w", c, err)
			}
			sb.WriteString(c + " " + sum + "\n")
		}
		if err := os.WriteFile(staticPath, []byte(sb.String()), 0o644); err != nil {
			return nil, err
		}
	}

	if err := writeList(idir, metaDir, pkgBaseName); err != nil {
		return nil, err
	}

	archComponent := plan.PKGArch
	if pkg.ArchAll {
		archComponent = "all"
	}
	apkName := fmt.Sprintf("%s-%s_%s.apk", pkgBaseName, p.version, archComponent)
	outPath := filepath.Join(outDir, apkName)

	args := []string{"mkpkg",
		"--info", "name:" + pkgBaseName,
		"--info", "version:" + p.version,
		"--info", "description:" + escapeDesc(pkg.Description),
	}
	if pkg.License != "" {
		args = append(args, "--info", "license:"+pkg.License)
	}
	if pkg.Maintainer != "" {
		args = append(args, "--info", "maintainer:"+pkg.Maintainer)
	}
	if pkg.URL != "" {
		args = append(args, "--info", "url:"+pkg.URL)
	}
	if pkg.ABIVersion != "" {
		args = append(args, "--info", "tags:openwrt:abiversion="+pkg.ABIVersion)
	}
	if pkg.ArchAll {
		args = append(args, "--info", "arch:noarch")
	} else {
		args = append(args, "--info", "arch:"+plan.PKGArch)
	}
	args = append(args, "--info", "origin:"+p.project)
	if len(pkg.Provides) > 0 {
		args = append(args, "--info", "provides:"+strings.Join(pkg.Provides, " "))
	}
	if len(pkg.Depends) > 0 {
		args = append(args, "--info", "depends:"+strings.Join(pkg.Depends, " "))
	}
	for scriptType, scriptPath := range pkg.Scripts {
		args = append(args, "--script", scriptType+":"+scriptPath)
	}
	args = append(args, "--files", idir, "--output", outPath)

	cmd := exec.Command(p.apkBin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("apk mkpkg: %w", err)
	}
	return &PackResult{ApkPath: outPath, PKGArch: plan.PKGArch}, nil
}

func installBinary(buildOut, name, dest, idir string) error {
	src := filepath.Join(buildOut, name)
	if _, err := os.Stat(src); os.IsNotExist(err) {
		candidates, globErr := filepath.Glob(filepath.Join(buildOut, "*"))
		if globErr != nil || len(candidates) == 0 {
			return fmt.Errorf("binary %s not found in %s", name, buildOut)
		}
		src = candidates[0]
	}
	relDest := strings.TrimPrefix(dest, "/")
	destDir := filepath.Join(idir, relDest)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}
	return copyFile(src, filepath.Join(destDir, filepath.Base(src)), 0o755)
}

func installFile(fe config.FileEntry, idir string) error {
	info, err := os.Stat(fe.Src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return filepath.Walk(fe.Src, func(path string, fi os.FileInfo, _ error) error {
			if fi == nil || fi.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(fe.Src, path)
			if err != nil {
				return err
			}
			dst := filepath.Join(idir, strings.TrimPrefix(fe.Dest, "/"), rel)
			return copyFile(path, dst, 0o644)
		})
	}
	mode := os.FileMode(0o644)
	if fe.Mode != "" {
		var m uint32
		if _, err := fmt.Sscanf(fe.Mode, "%o", &m); err != nil {
			return fmt.Errorf("invalid mode %q: %w", fe.Mode, err)
		}
		mode = os.FileMode(m)
	}
	dst := filepath.Join(idir, strings.TrimPrefix(fe.Dest, "/"))
	return copyFile(fe.Src, dst, mode)
}

func copyFile(src, dst string, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func writeList(idir, metaDir, pkgBaseName string) error {
	var files []string
	err := filepath.Walk(idir, func(path string, fi os.FileInfo, _ error) error {
		if fi == nil || fi.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(idir, path)
		if err != nil {
			return err
		}
		files = append(files, "/"+rel)
		return nil
	})
	if err != nil {
		return err
	}
	sort.Strings(files)
	listPath := filepath.Join(metaDir, pkgBaseName+".list")
	return os.WriteFile(listPath, []byte(strings.Join(files, "\n")+"\n"), 0o644)
}

func hasConfFiles(pkg config.Package) bool {
	for _, fe := range pkg.Files {
		if fe.Conf {
			return true
		}
	}
	return false
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func escapeDesc(s string) string {
	return strings.ReplaceAll(s, "\n", " ")
}
