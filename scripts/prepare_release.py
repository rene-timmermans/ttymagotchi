#!/usr/bin/env python3
"""Build the portable Linux release and render native packaging metadata."""
import hashlib
from pathlib import Path
import shutil
import subprocess
import tarfile
import tempfile
import os

root = Path(__file__).resolve().parent.parent
os.chdir(root)
version = (root / "VERSION").read_text().strip()
out = root / "dist"
out.mkdir(exist_ok=True)
env = dict(os.environ, CGO_ENABLED="0", GOOS="linux", GOARCH="amd64")
subprocess.run(["go", "build", "-trimpath", "-ldflags", f"-s -w -X main.version={version}", "-o", "bin/ttymagotchi", "./cmd/ttymagotchi"], env=env, check=True)
archive = out / f"ttymagotchi-{version}-linux-amd64.tar.gz"
with tempfile.TemporaryDirectory() as td:
    stage = Path(td) / f"ttymagotchi-{version}"
    (stage / "bin").mkdir(parents=True)
    shutil.copy2(root / "bin/ttymagotchi", stage / "bin/ttymagotchi")
    for name in ["LICENSE", "README.md", "VERSION"]:
        shutil.copy2(root / name, stage / name)
    shutil.copytree(root / "packaging", stage / "packaging")
    epoch = int(subprocess.check_output(["git", "log", "-1", "--format=%ct"]).strip())
    subprocess.run(["tar", "--sort=name", f"--mtime=@{epoch}", "--owner=0", "--group=0", "--numeric-owner", "-czf", str(archive), "-C", td, stage.name], check=True)
checksum = hashlib.sha256(archive.read_bytes()).hexdigest()
(out / "PKGBUILD").write_text((root / "packaging/PKGBUILD.in").read_text().replace("@VERSION@", version).replace("@BINARY_SHA256@", checksum))
print(archive)
