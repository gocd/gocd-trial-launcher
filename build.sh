#!/bin/bash

set -e

cd "$(dirname "$0")"

PROGNAME="run-gocd"

rm -f "$PROGNAME"
rm -rf dist

RELEASE="${RELEASE:-X.x.x}"

for arg in $@; do
  case $arg in
    --verbose)
      extra_flags="$extra_flags -v"
      shift
      ;;
    --skip-tests)
      skip=true
      shift
      ;;
    --prod)
      multiplatform=true
      shift
      ;;
    --release)
      RELEASE="$2"
      shift
      ;;
    --release=*)
      RELEASE="${arg#*=}"
      shift
      ;;
    *)
      shift
      ;;
  esac
done

RELEASE="${RELEASE}-${GO_PIPELINE_LABEL:-localbuild}"

function ldflags {
  local _os="${1:-$(go env GOOS)}"
  local _arch="${2:-$(go env GOARCH)}"

  echo "-X main.Version=${RELEASE} -X main.GitCommit=${GIT_COMMIT} -X main.Platform=${_arch}-${_os}"
}

go version

echo "Fetching dependencies"
go get -tags netgo $extra_flags ./...

if [ "true" = "$skip" ]; then
  echo "Skipping tests"
else
   go test $extra_flags ./...
fi

if (which git &> /dev/null); then
  GIT_COMMIT=$(git rev-list --abbrev-commit -1 HEAD)
else
  GIT_COMMIT="unknown"
fi

if [ "true" = "$multiplatform" ]; then
  platforms=(
    darwin/arm64
    linux/amd64
    windows/amd64
  )

  echo "Release: ${RELEASE}, Revision: ${GIT_COMMIT}"

  for plt in "${platforms[@]}"; do
    mkdir -p "dist/${plt}"
    arr=(${plt//\// })
    _os="${arr[0]}"
    _arch="${arr[1]}"
    name="$PROGNAME"
    build_tags="netgo"

    if [ "windows" = "${_os}" ]; then
      name="$name.exe"
      build_tags=""
    fi

    echo "Building $plt..."

    GOOS="${_os}" go get -tags "${build_tags}" $extra_flags ./...
    GOOS="${_os}" GOARCH="${_arch}" go build \
      -tags "${build_tags}" \
      -a \
      -o "dist/${plt}/${name}" \
      -ldflags "$(ldflags "$_os" "$_arch")"
  done
else
  go build \
    -ldflags "$(ldflags)" \
    -o "$PROGNAME"
fi
