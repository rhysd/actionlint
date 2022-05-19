#!/bin/bash

set -e -o pipefail

# Default value is updated manually on release
version="1.6.13"

function usage() {
    echo 'USAGE:' >&2
    echo '  bash download-actionlint.bash [DIR] [VERSION]' >&2
    echo >&2
    echo 'This script downloads actionlint binary from the following release page. curl' >&2
    echo 'command is required as dependency' >&2
    echo 'https://github.com/rhysd/actionlint/releases' >&2
    echo >&2
    echo 'DIR:' >&2
    echo '  Directory to put the downloaded binary (e.g. /path/to/dir). When this value is' >&2
    echo '  omitted, the binary will be put in the current directory.' >&2
    echo >&2
    echo 'VERSION:' >&2
    echo '   Version of actionlint to download (e.g. 1.6.9). When this value is omitted,' >&2
    echo '   the latest version will be selected.' >&2
    echo >&2
    echo 'EXAMPLE:' >&2
    echo '  - Download the latest binary to the current directory' >&2
    echo >&2
    echo '      $ bash download-actionlint.bash' >&2
    echo >&2
    echo '  - Download the latest binary to /usr/bin' >&2
    echo >&2
    echo '      $ bash download-actionlint.bash /usr/bin' >&2
    echo >&2
    echo '  - Download version 1.6.9 to the current directory' >&2
    echo >&2
    echo '      $ bash download-actionlint.bash 1.6.9' >&2
    echo >&2
    echo '  - Download version 1.6.9 to /usr/bin' >&2
    echo >&2
    echo '      $ bash download-actionlint.bash /usr/bin 1.6.9' >&2
}

if [[ "$1" == "-h" || "$1" == "--help" ]]; then
    usage
    exit
fi

target_dir="$(pwd)"
if [ -n "$1" ]; then
    if [[ "$1" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        version="$1"
    elif [ -d "$1" ]; then
        target_dir="${1%/}"
    else
        echo "Directory '$1' does not exist" >&2
        echo >&2
        usage
        exit 1
    fi
fi

if [ -n "$2" ]; then
    if [[ "$2" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        version="$2"
    else
        echo "Given version '${version}' does not match to regex" '^[0-9]+\.[0-9]+\.[0-9]+$' >&2
        echo >&2
        usage
        exit 1
    fi
fi

echo "Start downloading actionlint v${version} to ${target_dir}"

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
        echo "Could not determine arch from machine hardware name '${machine}'" >&2
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
    unzip "$tempdir/tmp.zip" actionlint.exe -d "$target_dir"
    rm -r "$tempdir"
    exe="$target_dir/actionlint.exe"
else
    curl -L "${url}" | tar xvz -C "$target_dir" actionlint
    exe="$target_dir/actionlint"
fi

echo "Downloaded and unarchived executable: ${exe}"

echo "Done: $("${exe}" -version)"

if [ -n "$GITHUB_ACTION" ]; then
    # On GitHub Actions, set executable path to output
    echo "::set-output name=executable::${exe}"
fi
