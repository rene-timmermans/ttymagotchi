#!/bin/sh
set -eu
root=$1
install -Dm755 bin/ttymagotchi "$root/usr/bin/ttymagotchi"
install -Dm644 LICENSE "$root/usr/share/licenses/ttymagotchi/LICENSE"
install -Dm644 README.md "$root/usr/share/doc/ttymagotchi/README.md"
install -Dm644 packaging/ttymagotchi.desktop "$root/usr/share/applications/ttymagotchi.desktop"
