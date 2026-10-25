#!/bin/sh
set -eu

repo="kazuki-kanaya/codex-profiles"
install_dir="${INSTALL_DIR:-$HOME/.local/bin}"

case "$(uname -s)" in
  Darwin) os="darwin" ;;
  Linux) os="linux" ;;
  *) echo "Unsupported OS" >&2; exit 1 ;;
esac

case "$(uname -m)" in
  arm64|aarch64) arch="arm64" ;;
  x86_64|amd64) arch="amd64" ;;
  *) echo "Unsupported architecture" >&2; exit 1 ;;
esac

release_url=$(curl -fsSL -o /dev/null -w '%{url_effective}' \
  "https://github.com/$repo/releases/latest")
tag="${release_url##*/}"
version="${tag#v}"

case "$tag" in
  v[0-9]*) ;;
  *) echo "Could not determine the latest version" >&2; exit 1 ;;
esac

archive="codex-profiles_${version}_${os}_${arch}.tar.gz"
checksums="codex-profiles_${version}_checksums.txt"
download_url="https://github.com/$repo/releases/download/$tag"

tmp_dir=$(mktemp -d)
trap 'rm -rf "$tmp_dir"' 0
trap 'exit 1' HUP INT TERM

echo "Downloading codex-profiles $version ($os/$arch)..."

curl -fsSL "$download_url/$archive" -o "$tmp_dir/$archive"
curl -fsSL "$download_url/$checksums" -o "$tmp_dir/checksums.txt"

expected=$(awk -v file="$archive" '$2 == file {print $1}' \
  "$tmp_dir/checksums.txt")

if command -v sha256sum >/dev/null 2>&1; then
  actual=$(sha256sum "$tmp_dir/$archive" | awk '{print $1}')
elif command -v shasum >/dev/null 2>&1; then
  actual=$(shasum -a 256 "$tmp_dir/$archive" | awk '{print $1}')
else
  echo "sha256sum or shasum is required" >&2
  exit 1
fi

if [ -z "$expected" ] || [ "$actual" != "$expected" ]; then
  echo "Checksum verification failed" >&2
  exit 1
fi

tar -xzf "$tmp_dir/$archive" -C "$tmp_dir" codex-profiles
mkdir -p "$install_dir"
install -m 755 "$tmp_dir/codex-profiles" "$install_dir/codex-profiles"

printf '\n🎉 Installed codex-profiles to %s/codex-profiles\n' "$install_dir"

case ":$PATH:" in
  *":$install_dir:"*) exit 0 ;;
esac

escaped_dir=$(printf '%s' "$install_dir" | sed 's/[\\"$`]/\\&/g')
path_line="export PATH=\"$escaped_dir:\$PATH\""

user_shell="${SHELL:-}"
case "${user_shell##*/}" in
  zsh) shell_config="${ZDOTDIR:-$HOME}/.zshrc" ;;
  bash) shell_config="$HOME/.bashrc" ;;
  *)
    printf '\nAdd this line to your shell configuration:\n\n  %s\n' "$path_line"
    exit 0
    ;;
esac

if ! grep -Fqx "$path_line" "$shell_config" 2>/dev/null; then
  if printf '\n# codex-profiles\n%s\n' "$path_line" >> "$shell_config"; then
    printf '\nPATH configuration added to %s.\n' "$shell_config"
  else
    printf '\nCould not update %s. Add this line manually:\n\n  %s\n' "$shell_config" "$path_line" >&2
    exit 0
  fi
fi

printf '\nRun the following command or open a new terminal:\n\n  source "%s"\n' "$shell_config"
