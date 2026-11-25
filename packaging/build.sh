#!/bin/sh
# Run with native packaging tools installed: build.sh debian|fedora|arch INPUT OUTPUT
set -eu
family=$1
input=$(realpath "$2")
output=$(realpath "$3")
version=$(cat VERSION)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
archive="ttymagotchi-$version-linux-amd64.tar.gz"
cp "$input/$archive" "$work/"
cd "$work"
case "$family" in
 debian)
  tar xzf "$archive"
  cd "ttymagotchi-$version"
  sh packaging/stage.sh "$work/root"
  mkdir -p "$work/root/DEBIAN"
  sed "s/@VERSION@/$version/g" packaging/debian/control.in > "$work/root/DEBIAN/control"
  dpkg-deb --root-owner-group --build "$work/root" "$output/ttymagotchi_${version}-1_amd64.deb"
  ;;
 fedora)
  tar xzf "$archive"
  mkdir -p "$work/rpm/SOURCES"
  cp "$archive" "$work/rpm/SOURCES/"
  rpmbuild -bb --define "_topdir $work/rpm" --define 'dist %{nil}' --define "tty_version $version" "ttymagotchi-$version/packaging/fedora/ttymagotchi.spec"
  cp "$work"/rpm/RPMS/x86_64/*.rpm "$output/"
  ;;
 arch)
  cp "$input/PKGBUILD" .
  if [ "$(id -u)" = 0 ]; then
   id builder >/dev/null 2>&1 || useradd -m builder
   chmod 755 "$work"
   chown -R builder "$work"
   su builder -c "cd '$work' && makepkg --noconfirm"
  else
   makepkg --noconfirm
  fi
  cp ./*.pkg.tar.zst "$output/"
  ;;
 *) echo 'Expected debian, fedora or arch' >&2; exit 2 ;;
esac
