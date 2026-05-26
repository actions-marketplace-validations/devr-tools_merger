#!/usr/bin/env zsh
set -euo pipefail

toolchain_version="go1.25.10"
toolchain_bin="${HOME}/go/bin/${toolchain_version}"
toolchain_root="${HOME}/sdk/${toolchain_version}"

if ! command -v go >/dev/null 2>&1; then
  echo "go is required to install ${toolchain_version}" >&2
  exit 1
fi

if [[ ! -x "${toolchain_bin}" ]]; then
  echo "installing ${toolchain_version} launcher" >&2
  GOTELEMETRY=off go install "golang.org/dl/${toolchain_version}@latest"
fi

if [[ ! -x "${toolchain_root}/bin/go" ]]; then
  echo "downloading ${toolchain_version} toolchain" >&2
  GOTELEMETRY=off "${toolchain_bin}" download >/dev/null
fi

cat <<EOF
export GOROOT="${toolchain_root}"
export PATH="${toolchain_root}/bin:\$PATH"
export GO="${toolchain_root}/bin/go"
EOF
