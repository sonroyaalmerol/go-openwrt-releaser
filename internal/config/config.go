package config

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

type File struct {
	ProjectName string    `yaml:"project_name"`
	Version     string    `yaml:"version"`
	Go          Go        `yaml:"go"`
	SDK         SDK       `yaml:"sdk"`
	Targets     []Target  `yaml:"targets"`
	Packages    []Package `yaml:"packages"`
	Release     Release   `yaml:"release"`

	sourcePath string
}

type Go struct {
	Module  string            `yaml:"module"`
	Main    string            `yaml:"main"`
	Binary  string            `yaml:"binary"`
	LDFlags []string          `yaml:"ldflags"`
	Vars    []string          `yaml:"vars"`
	Tags    []string          `yaml:"tags"`
	Env     map[string]string `yaml:"env"`
	CGO     bool              `yaml:"cgo"`
}

type SDK struct {
	OpenWrtVersion  string `yaml:"openwrt_version"`
	GCCVersion      string `yaml:"gcc_version"`
	Libc            string `yaml:"libc"`
	ToolchainTarget string `yaml:"toolchain_target"`
}

type Target struct {
	OpenWrt  string `yaml:"openwrt"`
	PKGArch  string `yaml:"pkgarch"`
	GOArch   string `yaml:"goarch"`
	GOARM    string `yaml:"goarm"`
	GOAMD64  string `yaml:"goamd64"`
	GO386    string `yaml:"go386"`
	GOMIPS   string `yaml:"gomips"`
	GOMIPS64 string `yaml:"gomips64"`
	GOARM64  string `yaml:"goarm64"`
	Skip     bool   `yaml:"skip"`
}

type Package struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	License     string            `yaml:"license"`
	Maintainer  string            `yaml:"maintainer"`
	URL         string            `yaml:"url"`
	Depends     []string          `yaml:"depends"`
	Provides    []string          `yaml:"provides"`
	ArchAll     bool              `yaml:"arch_all"`
	Binary      bool              `yaml:"binary"`
	BinaryDest  string            `yaml:"binary_dest"`
	Files       []FileEntry       `yaml:"files"`
	Conffiles   []string          `yaml:"conffiles"`
	Scripts     map[string]string `yaml:"scripts"`
	ABIVersion  string            `yaml:"abi_version"`
}

type FileEntry struct {
	Src  string `yaml:"src"`
	Dest string `yaml:"dest"`
	Mode string `yaml:"mode"`
	Conf bool   `yaml:"conf"`
}

type Release struct {
	GitHub     *GitHub `yaml:"github"`
	SignKeyEnv string  `yaml:"sign_key_env"`
	Pages      bool    `yaml:"pages"`
}

type GitHub struct {
	Owner string `yaml:"owner"`
	Name  string `yaml:"name"`
}

func Load(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var f File
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	f.sourcePath = path
	if err := f.validate(); err != nil {
		return nil, err
	}
	return &f, nil
}

var goArchRe = regexp.MustCompile(`^[a-z0-9_]+$`)

func (f *File) validate() error {
	if f.ProjectName == "" {
		return fmt.Errorf("project_name is required")
	}
	if f.Go.Module == "" {
		return fmt.Errorf("go.module is required")
	}
	if f.Go.Main == "" {
		f.Go.Main = "."
	}
	if f.Go.Binary == "" {
		f.Go.Binary = f.ProjectName
	}
	if f.SDK.OpenWrtVersion == "" {
		f.SDK.OpenWrtVersion = "25.12.4"
	}
	if f.SDK.GCCVersion == "" {
		f.SDK.GCCVersion = "14.3.0"
	}
	if f.SDK.Libc == "" {
		f.SDK.Libc = "musl"
	}
	if f.SDK.ToolchainTarget == "" && len(f.Targets) > 0 {
		f.SDK.ToolchainTarget = f.Targets[0].OpenWrt
	}
	if len(f.Targets) == 0 {
		return fmt.Errorf("at least one target is required")
	}
	for i := range f.Targets {
		t := &f.Targets[i]
		if t.PKGArch == "" && t.OpenWrt != "" {
			t.PKGArch = "auto"
		}
		if t.GOArch != "" && !goArchRe.MatchString(t.GOArch) {
			return fmt.Errorf("target %d: invalid goarch %q", i, t.GOArch)
		}
	}
	if len(f.Packages) == 0 {
		return fmt.Errorf("at least one package is required")
	}
	return nil
}

func (f *File) Path() string { return f.sourcePath }
