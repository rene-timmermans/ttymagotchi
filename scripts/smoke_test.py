#!/usr/bin/env python3
"""Exercise terminal navigation and persistence using an isolated save directory."""
import datetime
import fcntl
import os
from pathlib import Path
import pty
import select
import struct
import sys
import tempfile
import termios
import time

binary = str(Path(sys.argv[1] if len(sys.argv) > 1 else "bin/termagotchi").resolve())
with tempfile.TemporaryDirectory(prefix="termagotchi-smoke-") as config_dir:
    save_dir = Path(config_dir) / "termagotchi"
    save_dir.mkdir()
    now = datetime.datetime.now(datetime.timezone.utc).isoformat()
    save = save_dir / "config.yml"
    save.write_text(f"""app:
  speed_percent: 50
  last_login: {now}
tamagotchi:
  name: Smoke
  hunger: 50
  happiness: 50
  health: 80
  energy: 10
  cleanliness: 10
  grooming: 10
  stage: baby
  created: {now}
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
        assert text.encode() in output, f"Missing terminal output: {text}"

    try:
        pump(0.8)
        send(b"\x19")  # Ctrl+Y: Hygiene
        saw("HYGIENE")
        send(b"\x1b[B\r")
        saw("Washing")
        pump(0.5)
        send(b"\x1b")  # Skip bath
        pump(0.6)
        saw("Status:")
        send(b"\x19")
        send(b"\x1b[B\r")  # Select the next item: brush fur
        saw("Brushing")
        pump(4.2)  # Complete this animation naturally
        saw("brushed!")
        send(b"\r")
        send(b"\x0c")  # Ctrl+L: Sleep
        send(b"\x1b[B\x1b[B\r")
        saw("Sleeping")
        saw("Zzz")
        send(b"\x1b")
        pump(0.6)
        saw("Status:")
        send(b"\x0f")  # Ctrl+O: Options
        saw("Speed:")
        send(b"+")
        assert "speed_percent: 51" in save.read_text(), "Speed was not saved immediately"
        send(b"-")
        assert "speed_percent: 50" in save.read_text(), "Speed decrease was not saved"
        send(b"\x06")  # Ctrl+F: Feed
        send(b"\x1b[B\x1b[B\x1b[B\r")
        saw("Dinner")
        pump(3.2)
        saw("Yum!")
        send(b"\r")
        send(b"\x1b[B" * 7 + b"\r")  # Steak exceeds remaining hunger
        pump(3.2)
        saw("Overfed!")
        send(b"\x1b")
        pump(0.6)
        saw("Status:")
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
        for expected in ("speed_percent: 50", "cleanliness: 100", "grooming: 100", "hunger: 0", "health: 70"):
            assert expected in saved, f"Missing saved state: {expected}\n{saved}"
        print("PTY smoke test passed: wash, brushing, sleep, speed, feeding, overfeeding, quit and save.")
    finally:
        if not exited:
            try:
                os.kill(pid, 9)
                os.waitpid(pid, 0)
            except ProcessLookupError:
                pass
        os.close(master)
