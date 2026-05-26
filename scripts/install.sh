#!/usr/bin/env bash
# gm installer for macOS and Linux
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/angei24/go-manager/main/scripts/install.sh | bash
#   ./scripts/install.sh
#   GM_INSTALL_DIR=~/bin ./scripts/install.sh --from-source

set -euo pipefail

GM_REPO="${GM_REPO:-angei24/go-manager}"
GM_BRANCH="${GM_BRANCH:-main}"
GM_VERSION="${GM_VERSION:-}"          # empty = latest release, or "source" to force build
GM_INSTALL_DIR="${GM_INSTALL_DIR:-}"    # default set below
GM_FROM_SOURCE="${GM_FROM_SOURCE:-0}"

usage() {
	cat <<'EOF'
gm installer (macOS / Linux)

Usage:
  install.sh [options]

Options:
  --from-source    Build from git source (requires Go 1.21+)
  --dir PATH       Install directory (default: ~/.local/bin)
  --repo OWNER/REPO  GitHub repository (default: angei24/go-manager)
  --branch NAME    Git branch when building from source (default: main)
  --version TAG    Release tag to download (default: latest)
  -h, --help       Show this help

Environment:
  GM_INSTALL_DIR, GM_REPO, GM_BRANCH, GM_VERSION, GM_FROM_SOURCE

Examples:
  curl -fsSL .../install.sh | bash
  ./scripts/install.sh --from-source
  GM_INSTALL_DIR=$HOME/bin ./scripts/install.sh
EOF
}

log() { printf '==> %s\n' "$*"; }
warn() { printf 'warning: %s\n' "$*" >&2; }
die() { printf 'error: %s\n' "$*" >&2; exit 1; }

parse_args() {
	while [[ $# -gt 0 ]]; do
		case "$1" in
		--from-source) GM_FROM_SOURCE=1; shift ;;
		--dir)
			[[ $# -ge 2 ]] || die "--dir requires a path"
			GM_INSTALL_DIR="$2"
			shift 2
			;;
		--repo)
			[[ $# -ge 2 ]] || die "--repo requires OWNER/REPO"
			GM_REPO="$2"
			shift 2
			;;
		--branch)
			[[ $# -ge 2 ]] || die "--branch requires a name"
			GM_BRANCH="$2"
			shift 2
			;;
		--version)
			[[ $# -ge 2 ]] || die "--version requires a tag"
			GM_VERSION="$2"
			shift 2
			;;
		-h | --help)
			usage
			exit 0
			;;
		*)
			die "unknown option: $1 (try --help)"
			;;
		esac
	done
}

detect_platform() {
	local os arch
	os="$(uname -s | tr '[:upper:]' '[:lower:]')"
	arch="$(uname -m)"
	case "$os" in
	linux) os="linux" ;;
	darwin) os="darwin" ;;
	*) die "unsupported OS: $os (use install.ps1 on Windows)" ;;
	esac
	case "$arch" in
	x86_64 | amd64) arch="amd64" ;;
	aarch64 | arm64) arch="arm64" ;;
	armv7l | armv6l) arch="arm" ;;
	i386 | i686) arch="386" ;;
	*) die "unsupported architecture: $arch" ;;
	esac
	printf '%s %s' "$os" "$arch"
}

default_install_dir() {
	if [[ -n "${GM_INSTALL_DIR:-}" ]]; then
		printf '%s' "$GM_INSTALL_DIR"
		return
	fi
	if [[ -d "$HOME/.local/bin" ]]; then
		printf '%s' "$HOME/.local/bin"
	elif [[ -d "$HOME/bin" ]]; then
		printf '%s' "$HOME/bin"
	else
		printf '%s' "$HOME/.local/bin"
	fi
}

in_path() {
	local dir="$1"
	case ":${PATH}:" in
	*":${dir}:"*) return 0 ;;
	*) return 1 ;;
	esac
}

path_hint() {
	local dir="$1"
	if in_path "$dir"; then
		return
	fi
	cat <<EOF

Add gm to your PATH (pick your shell):

  bash/zsh:  echo 'export PATH="$dir:\$PATH"' >> ~/.bashrc   # or ~/.zshrc
  fish:      fish -c "fish_add_path $dir"

Then restart your shell or run:  export PATH="$dir:\$PATH"
EOF
}

repo_root_from_script() {
	local src="${BASH_SOURCE[0]:-$0}"
	local dir
	dir="$(cd "$(dirname "$src")/.." && pwd)"
	if [[ -f "$dir/go.mod" && -f "$dir/cmd/gm/main.go" ]]; then
		printf '%s' "$dir"
		return 0
	fi
	return 1
}

need_cmd() {
	command -v "$1" >/dev/null 2>&1 || die "'$1' is required but not found in PATH"
}

build_from_dir() {
	local root="$1"
	local dest="$2"
	need_cmd go
	log "Building gm from $root ..."
	(
		cd "$root"
		GO111MODULE=on go build -ldflags="-s -w" -o "$dest" ./cmd/gm
	)
	chmod +x "$dest"
}

build_from_git() {
	local tmp dest
	dest="$1"
	tmp="$(mktemp -d)"
	trap 'rm -rf "$tmp"' RETURN
	need_cmd git
	need_cmd go
	log "Cloning https://github.com/${GM_REPO}.git (branch ${GM_BRANCH}) ..."
	git clone --depth 1 --branch "$GM_BRANCH" "https://github.com/${GM_REPO}.git" "$tmp/repo"
	build_from_dir "$tmp/repo" "$dest"
}

fetch_release_tag() {
	if [[ -n "$GM_VERSION" && "$GM_VERSION" != "latest" ]]; then
		printf '%s' "$GM_VERSION"
		return
	fi
	need_cmd curl
	local tag
	tag="$(curl -fsSL "https://api.github.com/repos/${GM_REPO}/releases/latest" | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1)"
	[[ -n "$tag" ]] || return 1
	printf '%s' "$tag"
}

download_release() {
	local os arch tag dest tmpdir asset url
	read -r os arch <<<"$(detect_platform)"
	dest="$1"
	tag="$(fetch_release_tag)" || return 1
	asset="gm_${tag}_${os}_${arch}.tar.gz"
	url="https://github.com/${GM_REPO}/releases/download/${tag}/${asset}"
	tmpdir="$(mktemp -d)"
	trap 'rm -rf "$tmpdir"' RETURN
	need_cmd curl
	log "Downloading ${url} ..."
	if ! curl -fsSL "$url" -o "$tmpdir/$asset"; then
		return 1
	fi
	tar -xzf "$tmpdir/$asset" -C "$tmpdir"
	if [[ ! -f "$tmpdir/gm" ]]; then
		warn "release archive missing gm binary"
		return 1
	fi
	install -m 0755 "$tmpdir/gm" "$dest"
}

install_binary() {
	local install_dir dest
	install_dir="$(default_install_dir)"
	mkdir -p "$install_dir"
	dest="${install_dir%/}/gm"

	if [[ "$GM_FROM_SOURCE" == "1" || "$GM_VERSION" == "source" ]]; then
		if root="$(repo_root_from_script)"; then
			build_from_dir "$root" "$dest"
		else
			build_from_git "$dest"
		fi
		return
	fi

	if download_release "$dest" 2>/dev/null; then
		log "Installed release binary to $dest"
		return
	fi

	warn "No GitHub release found (or download failed); building from source ..."
	if root="$(repo_root_from_script)"; then
		build_from_dir "$root" "$dest"
	else
		build_from_git "$dest"
	fi
}

verify_install() {
	local dest="$1"
	[[ -x "$dest" ]] || die "install failed: $dest not executable"
	"$dest" --help >/dev/null 2>&1 || die "install failed: gm --help did not run"
	log "Success! gm is ready at $dest"
}

main() {
	parse_args "$@"
	local install_dir dest
	install_dir="$(default_install_dir)"
	GM_INSTALL_DIR="$install_dir"
	log "Platform: $(detect_platform)"
	log "Install dir: $install_dir"
	install_binary
	dest="${install_dir%/}/gm"
	verify_install "$dest"
	path_hint "$install_dir"
}

main "$@"
