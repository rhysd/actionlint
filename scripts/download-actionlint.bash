#!/bin/bash

set -e -o pipefail

# Default value is updated manually on release
version="${1:-1.6.11}"

if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Given version '${version}' does not match to regex" '^[0-9]+\.[0-9]+\.[0-9]+$' >&2
    exit 1
fi

echo "Start downloading actionlint v${version}"

case "$OSTYPE" in
    linux-*)
        os=linux
        ext=tar.gz
    ;;
    darwin*)
        os=darwin
        ext=tar.gz
    ;;
    freebsd*)
        os=freebsd
        ext=tar.gz
    ;;
    msys|cygwin|win32)
        os=windows
        ext=zip
    ;;
    *)
        echo "OS '${OSTYPE}' is not supported. Note: If you're using Windows, please ensure bash is used to run this script" >&2
        exit 1
    ;;
esac


machine="$(uname -m)"
case "$machine" in
    x86_64) arch=amd64 ;;
    i?86) arch=386 ;;
    aarch64|arm64) arch=arm64 ;;
    arm*) arch=armv6 ;;
    *)
        echo "Could not determine arch from machine hardware name '${machine}'"
        exit 1
    ;;
esac

echo "Detected OS=${os} ext=${ext} arch=${arch}"

# https://github.com/rhysd/actionlint/releases/download/v1.0.0/actionlint_1.0.0_linux_386.tar.gz
file="actionlint_${version}_${os}_${arch}.${ext}"
url="https://github.com/rhysd/actionlint/releases/download/v${version}/${file}"

echo "Downloading ${url} with curl"

if [[ "$os" == "windows" ]]; then
    tempdir="$(mktemp -d actionlint.XXXXXXXXXXXXXXXX)"
    curl -L -o "$tempdir/tmp.zip" "${url}"
    unzip "$tempdir/tmp.zip" actionlint.exe -d .
    rm -r "$tempdir"
    exe="$(pwd)/actionlint.exe"
else
    curl -L "${url}" | tar xvz actionlint
    exe="$(pwd)/actionlint"
fi

echo "Downloaded and unarchived executable: ${exe}"

echo "Done: $("${exe}" -version)"

if [ -n "$GITHUB_ACTION" ]; then
    # On GitHub Actions, set executable path to output
    echo "::set-output name=executable::${exe}"
fi
