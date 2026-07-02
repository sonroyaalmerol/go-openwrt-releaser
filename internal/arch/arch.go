package arch

import (
	"fmt"
	"strings"
)

type Info struct {
	GOArch            string
	GOOS              string
	GOARM             string
	GOAMD64           string
	GO386             string
	GOMIPS            string
	GOMIPS64          string
	GOARM64           string
	PKGArch           string
	ArchSuffix        string
	Supported         bool
	UnsupportedReason string
	OpenWrtARCH       string
	CPUType           string
	Is32bit           bool
	Is64bit           bool
}

func (i Info) Env() []string {
	env := []string{
		"GOOS=" + i.GOOS,
		"GOARCH=" + i.GOArch,
		"CGO_ENABLED=0",
	}
	if i.GOARM != "" {
		env = append(env, "GOARM="+i.GOARM)
	}
	if i.GOAMD64 != "" {
		env = append(env, "GOAMD64="+i.GOAMD64)
	}
	if i.GO386 != "" {
		env = append(env, "GO386="+i.GO386)
	}
	if i.GOMIPS != "" {
		env = append(env, "GOMIPS="+i.GOMIPS)
	}
	if i.GOMIPS64 != "" {
		env = append(env, "GOMIPS64="+i.GOMIPS64)
	}
	if i.GOARM64 != "" {
		env = append(env, "GOARM64="+i.GOARM64)
	}
	return env
}

func Resolve(openwrtTarget string) (*Info, error) {
	parts := strings.SplitN(openwrtTarget, "/", 2)
	board := parts[0]
	subtarget := ""
	if len(parts) == 2 {
		subtarget = parts[1]
	}
	if subtarget == "" {
		subtarget = "generic"
	}

	key := board + "/" + subtarget
	if info, ok := subtargetDB[key]; ok {
		return fill(info), nil
	}
	if info, ok := boardDB[board]; ok {
		return fill(info), nil
	}
	return nil, fmt.Errorf("unknown OpenWrt target %q (board %q); set goarch/pkgarch explicitly in config", openwrtTarget, board)
}

func fill(b baseInfo) *Info {
	i := &Info{
		GOArch:            b.GOArch,
		GOOS:              "linux",
		GOARM:             b.GOARM,
		GOAMD64:           b.GOAMD64,
		GO386:             b.GO386,
		GOMIPS:            b.GOMIPS,
		GOMIPS64:          b.GOMIPS64,
		GOARM64:           b.GOARM64,
		PKGArch:           b.PKGArch,
		ArchSuffix:        b.ArchSuffix,
		Supported:         b.Supported,
		UnsupportedReason: b.UnsupportedReason,
		OpenWrtARCH:       b.OpenWrtARCH,
		CPUType:           b.CPUType,
	}
	switch b.GOArch {
	case "arm", "386", "mips", "mipsle":
		i.Is32bit = true
	case "arm64", "amd64", "mips64", "mips64le", "ppc64", "riscv64", "loong64":
		i.Is64bit = true
	}
	if i.GOARM == "" && b.GOArch == "arm" && i.Supported {
		i.GOARM = "7"
	}
	if i.GO386 == "" && b.GOArch == "386" {
		i.GO386 = "softfloat"
	}
	if i.GOMIPS == "" && (b.GOArch == "mips" || b.GOArch == "mipsle") {
		i.GOMIPS = "softfloat"
	}
	if i.GOMIPS64 == "" && (b.GOArch == "mips64" || b.GOArch == "mips64le") {
		i.GOMIPS64 = "softfloat"
	}
	if i.PKGArch == "" && b.OpenWrtARCH != "" {
		i.PKGArch = b.OpenWrtARCH
		if b.CPUType != "" {
			i.PKGArch += "_" + b.CPUType
		}
	}
	return i
}

type baseInfo struct {
	GOArch            string
	GOARM             string
	GOAMD64           string
	GO386             string
	GOMIPS            string
	GOMIPS64          string
	GOARM64           string
	PKGArch           string
	ArchSuffix        string
	Supported         bool
	UnsupportedReason string
	OpenWrtARCH       string
	CPUType           string
}

func arm(goarm, cpuType, cpuSubtype string) baseInfo {
	pkgarch := "arm_" + cpuType
	if cpuSubtype != "" {
		pkgarch += "_" + cpuSubtype
	}
	return baseInfo{
		GOArch:      "arm",
		GOARM:       goarm,
		ArchSuffix:  "eabi",
		Supported:   true,
		OpenWrtARCH: "arm",
		CPUType:     cpuType,
		PKGArch:     pkgarch,
	}
}

func arm64(cpuType string) baseInfo {
	pkgarch := "aarch64_generic"
	if cpuType != "" {
		pkgarch = "aarch64_" + cpuType
	}
	return baseInfo{
		GOArch:      "arm64",
		Supported:   true,
		OpenWrtARCH: "aarch64",
		CPUType:     cpuType,
		PKGArch:     pkgarch,
	}
}

func mips(cpuType string) baseInfo {
	return baseInfo{
		GOArch:      "mips",
		GOMIPS:      "softfloat",
		Supported:   true,
		OpenWrtARCH: "mips",
		CPUType:     cpuType,
		PKGArch:     "mips_" + cpuType,
	}
}

func mipsle(cpuType string) baseInfo {
	return baseInfo{
		GOArch:      "mipsle",
		GOMIPS:      "softfloat",
		Supported:   true,
		OpenWrtARCH: "mipsel",
		CPUType:     cpuType,
		PKGArch:     "mipsel_" + cpuType,
	}
}

func mips64(cpuType string) baseInfo {
	return baseInfo{
		GOArch:      "mips64",
		GOMIPS64:    "softfloat",
		Supported:   true,
		OpenWrtARCH: "mips64",
		CPUType:     cpuType,
		PKGArch:     "mips64_" + cpuType,
	}
}

func mips64le(cpuType string) baseInfo {
	return baseInfo{
		GOArch:      "mips64le",
		GOMIPS64:    "softfloat",
		Supported:   true,
		OpenWrtARCH: "mips64el",
		CPUType:     cpuType,
		PKGArch:     "mips64el_" + cpuType,
	}
}

func riscv64() baseInfo {
	return baseInfo{
		GOArch:      "riscv64",
		Supported:   true,
		OpenWrtARCH: "riscv64",
		CPUType:     "rv64gc",
		PKGArch:     "riscv64_rv64gc",
	}
}

func loong64() baseInfo {
	return baseInfo{
		GOArch:      "loong64",
		Supported:   true,
		OpenWrtARCH: "loongarch64",
		CPUType:     "generic",
		PKGArch:     "loongarch64_generic",
	}
}

func unsupported(goarch, reason, owArch string) baseInfo {
	return baseInfo{
		GOArch:            goarch,
		Supported:         false,
		UnsupportedReason: reason,
		OpenWrtARCH:       owArch,
	}
}

var boardDB = map[string]baseInfo{
	"ath79":       mips("24kc"),
	"ipq40xx":     arm("7", "cortex-a7", "neon-vfpv4"),
	"ipq806x":     arm("7", "cortex-a15", "neon-vfpv4"),
	"bcm4908":     arm64("cortex-a53"),
	"bcm53xx":     arm("7", "cortex-a9", "vfpv3-d16"),
	"gemini":      unsupported("arm", "FA526 is ARMv4; Go requires ARMv5 (GOARM>=5)", "arm"),
	"ixp4xx":      unsupported("", "big-endian ARM (armeb) is not a Go target", "armeb"),
	"kirkwood":    arm("5", "xscale", ""),
	"mpc85xx":     unsupported("", "32-bit PowerPC (e500) is not a Go target", "powerpc"),
	"apm821xx":    unsupported("", "32-bit PowerPC (e600) is not a Go target", "powerpc"),
	"qoriq":       unsupported("", "64-bit PowerPC (e5500); Go ppc64 target is incompatible with OpenWrt toolchain", "powerpc64"),
	"mxs":         arm("5", "arm926ej-s", ""),
	"octeon":      mips64("octeonplus"),
	"omap":        arm("7", "cortex-a8", "vfpv3"),
	"pistachio":   mipsle("24kc"),
	"ramips":      mipsle("24kc"),
	"bcm47xx":     mipsle("mips32"),
	"realtek":     mips("4kec"),
	"bmips":       mips("mips32"),
	"rockchip":    arm64("cortex-a55"),
	"sifiveu":     riscv64(),
	"starfive":    riscv64(),
	"d1":          riscv64(),
	"microchipsw": arm64("cortex-a53"),
	"loongarch64": loong64(),
	"tegra":       arm("7", "cortex-a9", "vfpv3-d16"),
	"zynq":        arm("7", "cortex-a9", "neon"),
	"mediatek":    arm64("cortex-a53"),
	"x86":         {},
}

var subtargetDB = map[string]baseInfo{
	"armsr/armv7":          arm("7", "cortex-a15", "neon-vfpv4"),
	"armsr/armv8":          arm64(""),
	"imx/cortexa7":         arm("7", "cortex-a7", "neon-vfpv4"),
	"imx/cortexa9":         arm("7", "cortex-a9", "neon"),
	"imx/cortexa53":        arm64("cortex-a53"),
	"mvebu/cortexa9":       arm("7", "cortex-a9", "vfpv3-d16"),
	"mvebu/cortexa53":      arm64("cortex-a53"),
	"mvebu/cortexa72":      arm64("cortex-a72"),
	"layerscape/armv7":     arm("7", "cortex-a7", "neon-vfpv4"),
	"layerscape/armv8_64b": arm64(""),
	"bcm27xx/bcm2708":      arm("6", "arm1176jzf-s", "vfp"),
	"bcm27xx/bcm2709":      arm("7", "cortex-a7", "neon-vfpv4"),
	"bcm27xx/bcm2710":      arm64("cortex-a53"),
	"bcm27xx/bcm2711":      arm64("cortex-a72"),
	"bcm27xx/bcm2712":      arm64("cortex-a76"),
	"sunxi/cortexa7":       arm("7", "cortex-a7", "neon-vfpv4"),
	"sunxi/cortexa8":       arm("7", "cortex-a8", "vfpv3"),
	"sunxi/cortexa53":      arm64("cortex-a53"),
	"sunxi/arm926ejs":      arm("5", "arm926ej-s", ""),
	"mediatek/mt7623":      arm("7", "cortex-a7", "neon-vfpv4"),
	"mediatek/mt7629":      arm("7", "cortex-a7", "neon-vfpv4"),
	"mediatek/mt7622":      arm64("cortex-a53"),
	"mediatek/filogic":     arm64("cortex-a53"),
	"qualcommax/ipq807x":   arm64("cortex-a53"),
	"qualcommax/ipq60xx":   arm64("cortex-a53"),
	"qualcommax/ipq50xx":   arm64("cortex-a53"),
	"rockchip/armv8":       arm64("cortex-a55"),
	"stm32/stm32mp1":       arm("7", "cortex-a7", "neon-vfpv4"),
	"x86/64":               {GOArch: "amd64", GOAMD64: "v1", Supported: true, OpenWrtARCH: "x86_64", PKGArch: "x86_64"},
	"x86/generic":          {GOArch: "386", GO386: "softfloat", Supported: true, OpenWrtARCH: "i386", PKGArch: "i386_pentium4"},
	"x86/legacy":           {GOArch: "386", GO386: "softfloat", Supported: true, OpenWrtARCH: "i386", PKGArch: "i386_pentium-mmx"},
	"x86/geode":            {GOArch: "386", GO386: "softfloat", Supported: true, OpenWrtARCH: "i386", PKGArch: "i386_geode"},
	"lantiq/xrx200":        mips("24kc"),
	"lantiq/xrx200_legacy": mips("24kc"),
	"lantiq/xway":          mips("24kc"),
	"lantiq/xway_legacy":   mips("24kc"),
	"lantiq/falcon":        mips("24kc"),
	"lantiq/ase":           mips("mips32"),
	"malta/be":             mips("24kc"),
	"malta/le":             mipsle("24kc"),
	"malta/be64":           mips64("24kc"),
	"malta/le64":           mips64le("24kc"),
	"ramips/mt7620":        mipsle("24kc"),
	"ramips/mt7621":        mipsle("24kc"),
	"ramips/mt76x8":        mipsle("24kc"),
	"ramips/rt305x":        mipsle("24kc"),
	"ramips/rt3883":        mipsle("24kc"),
	"ramips/rt288x":        mipsle("24kc"),
	"siflower/sf19a2890":   mipsle("24kc"),
	"siflower/sf21":        mipsle("24kc"),
	"bcm47xx/generic":      mipsle("mips32"),
	"bcm47xx/legacy":       mipsle("mips32"),
	"bcm47xx/mips74k":      mipsle("24kc"),
	"bmips/bcm6318":        mips("mips32"),
	"bmips/bcm6328":        mips("mips32"),
	"bmips/bcm6358":        mips("mips32"),
	"bmips/bcm6362":        mips("mips32"),
	"bmips/bcm6368":        mips("mips32"),
	"bmips/bcm63268":       mips("mips32"),
	"realtek/rtl838x":      mips("4kec"),
	"realtek/rtl839x":      mips("4kec"),
	"realtek/rtl930x":      mips("4kec"),
	"realtek/rtl930x_nand": mips("4kec"),
	"realtek/rtl931x":      mips("4kec"),
	"realtek/rtl931x_nand": mips("4kec"),
	"ath79/generic":        mips("24kc"),
	"ath79/mikrotik":       mips("24kc"),
	"ath79/nand":           mips("24kc"),
	"ath79/tiny":           mips("24kc"),
	"d1/generic":           riscv64(),
	"sifiveu/generic":      riscv64(),
	"starfive/generic":     riscv64(),
	"microchipsw/lan969x":  arm64("cortex-a53"),
	"ipq40xx/generic":      arm("7", "cortex-a7", "neon-vfpv4"),
	"ipq40xx/chromium":     arm("7", "cortex-a7", "neon-vfpv4"),
	"ipq40xx/mikrotik":     arm("7", "cortex-a7", "neon-vfpv4"),
	"ipq806x/generic":      arm("7", "cortex-a15", "neon-vfpv4"),
	"ipq806x/chromium":     arm("7", "cortex-a15", "neon-vfpv4"),
	"bcm4908/generic":      arm64("cortex-a53"),
	"bcm53xx/generic":      arm("7", "cortex-a9", "vfpv3-d16"),
	"tegra/generic":        arm("7", "cortex-a9", "vfpv3-d16"),
	"zynq/generic":         arm("7", "cortex-a9", "neon"),
	"omap/generic":         arm("7", "cortex-a8", "vfpv3"),
	"pistachio/generic":    mipsle("24kc"),
	"loongarch64/generic":  loong64(),
	"mpc85xx/p1010":        unsupported("", "32-bit PowerPC (e500) is not a Go target", "powerpc"),
	"mpc85xx/p1020":        unsupported("", "32-bit PowerPC (e500) is not a Go target", "powerpc"),
	"mpc85xx/p2020":        unsupported("", "32-bit PowerPC (e500) is not a Go target", "powerpc"),
	"apm821xx/nand":        unsupported("", "32-bit PowerPC (e600) is not a Go target", "powerpc"),
	"apm821xx/sata":        unsupported("", "32-bit PowerPC (e600) is not a Go target", "powerpc"),
	"qoriq/generic":        unsupported("", "64-bit PowerPC (e5500); Go ppc64 target is incompatible with OpenWrt toolchain", "powerpc64"),
	"gemini/generic":       unsupported("arm", "FA526 is ARMv4; Go requires ARMv5 (GOARM>=5)", "arm"),
	"ixp4xx/generic":       unsupported("", "big-endian ARM (armeb) is not a Go target", "armeb"),
	"kirkwood/generic":     arm("5", "xscale", ""),
	"mxs/generic":          arm("5", "arm926ej-s", ""),
	"at91/sam9x":           arm("5", "arm926ej-s", ""),
	"at91/sama5":           arm("7", "cortex-a5", "vfpv4"),
	"at91/sama7":           arm("7", "cortex-a7", "vfpv4"),
}

func SupportedBoards() []string {
	seen := map[string]bool{}
	for k := range boardDB {
		board := k
		if before, _, ok := strings.Cut(k, "/"); ok {
			board = before
		}
		seen[board] = true
	}
	for k := range subtargetDB {
		board := k[:strings.Index(k, "/")]
		seen[board] = true
	}
	var out []string
	for b := range seen {
		out = append(out, b)
	}
	return out
}
