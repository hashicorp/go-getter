#!/usr/bin/env bash

# We use it in build.yml for building release artifacts with CRT in the Go Build step. 

set -euo pipefail

# We don't want to get stuck in some kind of interactive pager
export GIT_PAGER=cat

# Get the build date from the latest commit since it can be used across all
# builds
function build_date() {
  # It's tricky to do an RFC3339 format in a cross platform way, so we hardcode UTC
  : "${DATE_FORMAT:="%Y-%m-%dT%H:%M:%SZ"}"
  git show --no-show-signature -s --format=%cd --date=format:"$DATE_FORMAT" HEAD
}

# Get the revision, which is the latest commit SHA
function build_revision() {
  git rev-parse HEAD
}

# Determine our repository by looking at our origin URL
function repo() {
  basename -s .git "$(git config --get remote.origin.url)"
}

# Determine the root directory of the repository
function repo_root() {
  git rev-parse --show-toplevel
}

# Build 
function build() {
  local revision
  local build_date
  local ldflags
  local msg

  # Get or set our basic build metadata
  revision=$(build_revision)
  build_date=$(build_date) #
  : "${BIN_PATH:="dist/"}" #if not run by actions-go-build (enos local) then set this explicitly
  : "${GO_TAGS:=""}"
  : "${KEEP_SYMBOLS:=""}"

  # Build our ldflags
  msg="--> Building go-getter revision $revision, built $build_date"

  # Strip the symbol and dwarf information by default
  if [ -n "$KEEP_SYMBOLS" ]; then
    ldflags=""
  else
    ldflags="-s -w "
  fi

  # if building locally with enos - don't need to set version/prerelease/metadata as the default from version_base.go will be used
  ldflags="${ldflags} -X github.com/hashicorp/go-getter/version.GitCommit=$revision -X github.com/hashicorp/go-getter/version.BuildDate=$build_date"

  if [[ ${BASE_VERSION+x} ]]; then
    msg="${msg}, base version ${BASE_VERSION}"
    ldflags="${ldflags} -X github.com/hashicorp/go-getter/version.Version=$BASE_VERSION"
  fi

  if [[ ${PRERELEASE_VERSION+x} ]]; then
    msg="${msg}, prerelease ${PRERELEASE_VERSION}"
    ldflags="${ldflags} -X github.com/hashicorp/go-getter/version.VersionPrerelease=$PRERELEASE_VERSION"
  fi

  if [[ ${METADATA_VERSION+x} ]]; then
    msg="${msg}, metadata ${METADATA_VERSION}"
    ldflags="${ldflags} -X github.com/hashicorp/go-getter/version.VersionMetadata=$METADATA_VERSION"
  fi

  # Build go-getter
  echo "$msg"
  go build -o "$BIN_PATH" -tags "$GO_TAGS" -ldflags "$ldflags" -trimpath -buildvcs=false
}

# Run the CRT Builder
function main() {
  case $1 in
  build)
    build
  ;;

  date)
    build_date
  ;;

  revision)
    build_revision
  ;;
  *)
    echo "unknown sub-command" >&2
    exit 1
  ;;
  esac
}

main "$@"