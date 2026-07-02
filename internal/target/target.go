package target

import (
	"fmt"

	"github.com/sonroyaalmerol/go-openwrt-releaser/internal/arch"
	"github.com/sonroyaalmerol/go-openwrt-releaser/internal/config"
)

type Plan struct {
	OpenWrt           string
	Board             string
	Subtarget         string
	GOOS              string
	PKGArch           string
	GOArch            string
	GOARM             string
	GOAMD64           string
	GO386             string
	GOMIPS            string
	GOMIPS64          string
	GOARM64           string
	ArchSuffix        string
	ApkArch           string
	Is32bit           bool
	Is64bit           bool
	Supported         bool
	UnsupportedReason string
}

func (p *Plan) Env() []string {
	info := arch.Info{
		GOOS:     "linux",
		GOArch:   p.GOArch,
		GOARM:    p.GOARM,
		GOAMD64:  p.GOAMD64,
		GO386:    p.GO386,
		GOMIPS:   p.GOMIPS,
		GOMIPS64: p.GOMIPS64,
		GOARM64:  p.GOARM64,
	}
	return info.Env()
}

func (p *Plan) OutputDir() string {
	return p.OpenWrt
}

func Resolve(t config.Target, sdk config.SDK) (*Plan, error) {
	info, err := arch.Resolve(t.OpenWrt)
	if err != nil {
		return nil, err
	}

	board, subtarget := splitTarget(t.OpenWrt)
	p := &Plan{
		OpenWrt:           t.OpenWrt,
		Board:             board,
		Subtarget:         subtarget,
		GOOS:              info.GOOS,
		PKGArch:           info.PKGArch,
		GOArch:            info.GOArch,
		GOARM:             info.GOARM,
		GOAMD64:           info.GOAMD64,
		GO386:             info.GO386,
		GOMIPS:            info.GOMIPS,
		GOMIPS64:          info.GOMIPS64,
		GOARM64:           info.GOARM64,
		ArchSuffix:        info.ArchSuffix,
		ApkArch:           info.OpenWrtARCH,
		Is32bit:           info.Is32bit,
		Is64bit:           info.Is64bit,
		Supported:         info.Supported,
		UnsupportedReason: info.UnsupportedReason,
	}

	if t.GOArch != "" {
		p.GOArch = t.GOArch
	}
	if t.PKGArch != "" && t.PKGArch != "auto" {
		p.PKGArch = t.PKGArch
	}
	if t.GOARM != "" {
		p.GOARM = t.GOARM
	}
	if t.GOAMD64 != "" {
		p.GOAMD64 = t.GOAMD64
	}
	if t.GO386 != "" {
		p.GO386 = t.GO386
	}
	if t.GOMIPS != "" {
		p.GOMIPS = t.GOMIPS
	}
	if t.GOMIPS64 != "" {
		p.GOMIPS64 = t.GOMIPS64
	}
	if t.GOARM64 != "" {
		p.GOARM64 = t.GOARM64
	}

	if !info.Supported {
		return p, fmt.Errorf("target %s is not supported by Go: %s", t.OpenWrt, info.UnsupportedReason)
	}

	return p, nil
}

func splitTarget(openwrt string) (string, string) {
	for i := 0; i < len(openwrt); i++ {
		if openwrt[i] == '/' {
			return openwrt[:i], openwrt[i+1:]
		}
	}
	return openwrt, "generic"
}
