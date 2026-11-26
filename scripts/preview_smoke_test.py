#!/usr/bin/env python3
"""Check that the live art preview works without reading or changing saves."""
import fcntl, os, pty, select, struct, sys, tempfile, termios, time
from pathlib import Path
columns, rows = (map(int, sys.argv[2:4]) if len(sys.argv) >= 4 else (80, 24))
binary = str(Path(sys.argv[1] if len(sys.argv) > 1 else 'bin/ttymagotchi').resolve())
with tempfile.TemporaryDirectory(prefix='ttymagotchi-preview-check-') as td:
    root = Path(td)
    for name in ('termagotchi', 'ttymagotchi'):
        (root / name).mkdir()
        (root / name / 'config.yml').write_text('invalid: [yaml -- must never be read\n')
    before = {str(p.relative_to(root)): p.read_bytes() for p in root.rglob('*') if p.is_file()}
    pid, master = pty.fork()
    if pid == 0:
        os.environ.update(TERM='xterm-256color', XDG_CONFIG_HOME=td)
        os.execv(binary, [binary, '--preview'])
    fcntl.ioctl(master, termios.TIOCSWINSZ, struct.pack('HHHH', rows, columns, 0, 0))
    output = bytearray()
    exited = False

    def pump(seconds):
        end = time.monotonic() + seconds
        while time.monotonic() < end:
            if select.select([master], [], [], 0.05)[0]:
                try:
                    output.extend(os.read(master, 65536))
                except OSError:
                    break
    try:
        pump(1)
        assert b'Voorbeeldmodus' in output, output[:400]
        for key in [b'2', b'3', b'4', b'5', b'1', b'c', b'c', b'\x1b[6~', b'\x1b[5~', b'\x1b[D' * 5, b'\x1b[C' * 3, b'\x1b[B', b'\x1b[A', b'\t' * 5, b'f', b' ', b' ']:
            os.write(master, key)
            pump(0.15)
        os.write(master, b'\t' * 5 + b'\x1b[D' * 2)  # Idle, crying.
        pump(2.0)
        assert '💧'.encode() in output, 'No animated tear in terminal output'
        os.write(master, b't')
        pump(0.2)
        os.write(master, b't')
        pump(0.2)
        os.write(master, b'\x1b')
        pump(1)
        for _ in range(20):
            child, status = os.waitpid(pid, os.WNOHANG)
            if child:
                exited = True
                assert os.waitstatus_to_exitcode(status) == 0, status
                break
            pump(0.1)
        assert exited, 'preview did not exit'
        assert b'DATA RACE' not in output
        after = {str(p.relative_to(root)): p.read_bytes() for p in root.rglob('*') if p.is_file()}
        assert before == after, 'preview modified a save'
        print(f'Preview PTY passed at {columns}x{rows}: animated display, controls, exit, malformed saves untouched.')
    finally:
        if not exited:
            os.kill(pid, 9)
            os.waitpid(pid, 0)
        os.close(master)
