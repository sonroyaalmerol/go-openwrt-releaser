# go-openwrt-releaser

Release tool for OpenWrt `.apk` packages built from Go source.

## Approach

Cross-compile the binary with native `go build`, then package it into an
`.apk` using the SDK's prebuilt `apk` host binary. The SDK is downloaded once
only to extract the `apk` tool; `make` and the golang feed are never invoked.

Build time for one target: ~40 seconds cold (SDK download), ~15 seconds warm.

## Pipeline

1. Download the OpenWrt SDK tarball for the configured toolchain target.
2. Extract the prebuilt `apk` host binary plus its shared library dependencies.
3. For each target: run `go build` with `GOOS`/`GOARCH`/`GOARM`/etc.
4. Assemble the install tree (`lib/apk/packages/` metadata, conffiles, file
   lists) and run `apk mkpkg`.
5. Run `apk mkndx` to produce the feed index (`packages.adb`).
6. Optionally sign the index with an ECDSA key.
7. Optionally create a GitHub release via `gh`.

## Supported targets

All 41 OpenWrt board families and their subtargets are mapped in
`internal/arch/arch.go`. Each entry records the Go GOARCH, GOARM/GO386/GOMIPS
defaults, the PKGARCH, and the SDK arch suffix.

| OpenWrt arch | Go GOARCH | Examples |
|---|---|---|
| aarch64 | arm64 | armsr/armv8, ipq807x, mediatek/filogic, bcm27xx/bcm2711 |
| arm | arm (GOARM 5/6/7) | ipq806x, ipq40xx, mvebu/cortexa9, sunxi |
| x86_64 | amd64 | x86/64 |
| i386 | 386 | x86/generic, x86/geode |
| mipsel | mipsle | ramips, bcm47xx, pistachio |
| mips | mips | ath79, lantiq, realtek, bmips |
| mips64 | mips64 | octeon, malta/be64 |
| mips64el | mips64le | malta/le64 |
| riscv64 | riscv64 | sifiveu, starfive, d1 |
| loongarch64 | loong64 | loongarch64/generic |

Unsupported (no Go port): `apm821xx`, `mpc85xx`, `qoriq` (PowerPC),
`ixp4xx` (big-endian ARM). These resolve with `Supported: false` and a reason.

## Config

`openwrt-releaser.yaml`:

```yaml
project_name: myapp
version: 1.0.0

go:
  module: github.com/me/myapp
  main: ./cmd/myapp
  binary: myapp
  ldflags: [-s, -w]

sdk:
  openwrt_version: "25.12.4"
  gcc_version: "14.3.0"
  libc: musl
  toolchain_target: ipq806x/generic

targets:
  - openwrt: ipq806x/generic      # arch auto-detected
  - openwrt: aarch64/generic
  - openwrt: x86/64

packages:
  - name: myapp
    description: "My app"
    license: MIT
    maintainer: me@example.com
    url: https://github.com/me/myapp
    depends: [libc]
    binary: true
    binary_dest: /usr/bin
    files:
      - src: files/myapp.init
        dest: /etc/init.d/myapp
        mode: "0755"
    conffiles:
      - /etc/config/myapp
    scripts:
      post-install: files/postinst

release:
  sign_key_env: APK_SIGN_KEY
  github:
    owner: me
    name: myapp
```

## Usage

```bash
# build and package only
openwrt-releaser --config openwrt-releaser.yaml --root . --skip-release

# full release from a tag
openwrt-releaser --tag v1.0.0
```

## Install

```bash
go install github.com/sonroyaalmerol/go-openwrt-releaser/cmd/openwrt-releaser@latest
```
