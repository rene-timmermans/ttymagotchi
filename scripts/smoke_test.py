#!/usr/bin/env python3
"""Exercise terminal navigation and persistence using an isolated save directory."""
import datetime
import fcntl
import os
import json
import re
from pathlib import Path
import pty
import select
import struct
import sys
import tempfile
import termios
import time

binary = str(Path(sys.argv[1] if len(sys.argv) > 1 else "bin/ttymagotchi").resolve())
pet_stage = sys.argv[2] if len(sys.argv) > 2 else "adult"
language = sys.argv[3] if len(sys.argv) > 3 else "en"
catalog = json.loads((Path(__file__).resolve().parent.parent / "internal/app/locales" / (language + ".json")).read_text())
young = pet_stage in ("egg", "hatchling", "baby")
care_field = "diaper" if young else "grooming"
with tempfile.TemporaryDirectory(prefix="ttymagotchi-smoke-") as config_dir:
    save_dir = Path(config_dir) / "ttymagotchi"
    save_dir.mkdir()
    current_time = datetime.datetime.now(datetime.timezone.utc)
    now = current_time.isoformat()
    birth = (current_time - datetime.timedelta(days={"egg": 0, "hatchling": 1, "baby": 2}.get(pet_stage, 14))).isoformat()
    save = save_dir / "config.yml"
    save.write_text(f"""app:
  language: {language}
  speed_percent: 50
  last_login: {now}
tamagotchi:
  name: Smoke
  hunger: 50
  happiness: 50
  health: 80
  energy: 10
  cleanliness: 10
  diaper: 10
  grooming: 10
  stage: {pet_stage}
  created: {birth}
  is_alive: true
""")
    pid, master = pty.fork()
    if pid == 0:
        os.environ.update(TERM="xterm-256color", XDG_CONFIG_HOME=config_dir)
        os.execv(binary, [binary])
    fcntl.ioctl(master, termios.TIOCSWINSZ, struct.pack("HHHH", 24, 80, 0, 0))
    output = bytearray()
    exited = False

    def pump(seconds):
        end = time.monotonic() + seconds
        while time.monotonic() < end:
            if select.select([master], [], [], min(0.05, max(0, end - time.monotonic())))[0]:
                try:
                    output.extend(os.read(master, 65536))
                except OSError:
                    break

    def send(data):
        os.write(master, data)
        pump(0.35)

    def saw(text):
        ansi = r"\x1b\[[0-?]*[ -/]*[@-~]|\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)|\x1b[()][A-Za-z0-9]"
        plain = re.sub(ansi, "", output.decode("utf-8", errors="replace"))
        assert "".join(text.split()) in "".join(plain.split()), f"Missing terminal output: {text}"

    def saw_key(key):
        saw(catalog[key].split("%")[0])

    try:
        pump(0.8)
        send(b"\x19")  # Ctrl+Y: Hygiene
        saw_key("=== HYGIENE ===")
        send(b"\x1b[B\r")
        saw_key("Washing away the dirt...")
        pump(0.5)
        send(b"\x1b")  # Skip bath
        pump(0.6)
        saw_key("Status: %s")
        send(b"\x19")
        send(b"\x1b[B\r")  # Select the next item: brush fur
        saw_key("Diaper" if young else "Brushing")
        pump(4.2)  # Complete this animation naturally
        saw_key("Fresh diaper! Diaper %d → 100" if young else "Freshly brushed! Brushing %d → 100")
        send(b"\r")
        send(b"\x0c")  # Ctrl+L: Sleep
        send(b"\x1b[B\x1b[B\r")
        saw_key("Sleeping peacefully...")
        saw("Zzz")
        send(b"\x1b")
        pump(0.6)
        saw_key("Status: %s")
        send(b"\x0f")  # Ctrl+O: Options
        saw_key("Speed: %d%%   ← slower / faster →   (+ / -)")
        send(b"\x1b[B" * (["en", "fr", "es", "de", "nl"].index(language) + 1) + b"\r")
        assert f"language: {language}" in save.read_text(), "Language choice not saved"
        send(b"+")
        assert "speed_percent: 51" in save.read_text(), "Speed was not saved immediately"
        send(b"-")
        assert "speed_percent: 50" in save.read_text(), "Speed decrease was not saved"
        send(b"\x10")  # Ctrl+P: Play
        send(b"\x1b[B" * 3 + b"\r")  # Run Around
        saw_key("Dodge six obstacles. Move left/right with arrows or A/D.")
        send(b"\r")
        saw_key("Dodge! ← / → or A / D")
        send(b"\x1b")
        pump(0.6)
        send(b"\x06")  # Ctrl+F: Feed
        send(b"\x1b[B\x1b[B\x1b[B\r")
        saw_key("Dinner is coming!")
        pump(3.2)
        saw_key("Yum! ")
        send(b"\r")
        send(b"\x1b[B" * 7 + b"\r")  # Steak exceeds remaining hunger
        pump(3.2)
        saw_key("Overfed! Your pet feels sick.\n")
        send(b"\x1b")
        pump(0.6)
        saw_key("Status: %s")
        send(b"\x06")
        send(b"\x1b[A" * 7 + b"\r")  # Apple while full: first stop.
        saw_key("I'm full! Feeding anyway costs %d health.")
        pump(3.2)
        saw_key("Stop feeding")
        send(b"\r")  # Default is stop; no penalty.
        pump(0.6)
        send(b"\x06")
        send(b"\r")  # Same apple, now explicitly continue.
        pump(3.2)
        saw_key("Feed anyway")
        send(b"\x1b[C\r")
        pump(3.2)
        saw_key("Overfed! Your pet feels sick.\n")
        send(b"\x1b")
        pump(0.6)
        send(b"\x0c")
        send(b"\r")  # Energy is now 80: refuse further sleep.
        saw_key("No!")
        saw_key("Not sleepy! Energy must be 60% or lower.")
        pump(0.4)
        send(b"\x1b")
        pump(0.6)
        send(b"\x03")
        for _ in range(50):
            child, status = os.waitpid(pid, os.WNOHANG)
            if child:
                exited = True
                break
            pump(0.1)
        assert exited, "Application did not quit"
        assert os.waitstatus_to_exitcode(status) == 0, f"Exit status: {status}"
        assert b"DATA RACE" not in output, "Go race detector reported a race"
        saved = save.read_text()
        for expected in ("speed_percent: 50", "cleanliness: 100", f"{care_field}: 100", "hunger: 0", "health: 60", "energy: 80", f"language: {language}"):
            assert expected in saved, f"Missing saved state: {expected}\n{saved}"
        print(f"PTY smoke test passed ({pet_stage}, {language}): wash, {care_field}, sleep, speed, feeding, refusal, stop/continue, overfeeding, quit and save.")
    finally:
        if not exited:
            try:
                os.kill(pid, 9)
                os.waitpid(pid, 0)
            except ProcessLookupError:
                pass
        os.close(master)
