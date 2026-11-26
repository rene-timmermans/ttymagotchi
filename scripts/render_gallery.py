#!/usr/bin/env python3
"""Render review PNGs from the game's own ASCII renderers (Go + Pillow).

No player save is opened. Install Pillow only for this optional development tool.
"""
import argparse
from collections import Counter
from functools import lru_cache
import hashlib
import json
import math
from pathlib import Path
import subprocess
import tempfile

from PIL import Image, ImageDraw, ImageFont

ROOT = Path(__file__).resolve().parent.parent
BG = (20, 27, 34)
PANEL = (27, 36, 45)
TEXT = (234, 241, 246)
MUTED = (154, 174, 190)
ACCENT = (114, 211, 178)
SPECIES = {"cat": "Cat", "dog": "Dog", "rabbit": "Rabbit", "guinea_pig": "Guinea pig", "penguin": "Penguin / Tux", "interface": "Interface", "upstream": "Upstream legacy"}
FONT_PATHS = ["/usr/share/fonts/TTF/DejaVuSansMono.ttf", "/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf", "/usr/share/fonts/noto/NotoSansMono-Regular.ttf"]
EMOJI_PATHS = ["/usr/share/fonts/twemoji/twemoji.ttf", "/usr/share/fonts/truetype/noto/NotoColorEmoji.ttf"]


@lru_cache(maxsize=16)
def font(size):
    for path in FONT_PATHS:
        if Path(path).exists():
            return ImageFont.truetype(path, size)
    raise SystemExit("Install a DejaVu Sans Mono or Noto Sans Mono font to render the gallery.")


def emoji_font():
    for path in EMOJI_PATHS:
        if not Path(path).exists():
            continue
        for size in (61, 109, 128, 72):
            try:
                return ImageFont.truetype(path, size)
            except OSError:
                continue
    return None


PALETTE_COLORS = [BG, PANEL, TEXT, MUTED, ACCENT, (90, 116, 139), (237, 190, 105), (238, 124, 133)]
PALETTE_COLORS += [(r, g, b) for r in range(0, 256, 51) for g in range(0, 256, 51) for b in range(0, 256, 51)]
PALETTE_COLORS += [(i, i, i) for i in range(0, 256, 8)]
PALETTE = Image.new("P", (1, 1))
PALETTE.putpalette([component for color in PALETTE_COLORS for component in color])
EMOJI = emoji_font()
EMOJI_CACHE = {}


def indexed(image):
    return image.quantize(palette=PALETTE, dither=Image.Dither.NONE)


def text_lines(card, draw, origin, size):
    face = font(size)
    advance = face.getlength("M")
    line_height = size + 2
    x0, y0 = origin
    for row, runs in enumerate(card["lines"]):
        column = 0
        for run in runs:
            text, cells = run["text"], run["cells"]
            x, y = x0 + round(column * advance), y0 + row * line_height
            if cells == 2 and EMOJI and any(ord(c) >= 0x1F000 or c in "⭕" for c in text):
                key = (text, round(cells * advance), line_height)
                if key not in EMOJI_CACHE:
                    stamp = Image.new("RGBA", (180, 160))
                    ImageDraw.Draw(stamp).text((0, 0), text, font=EMOJI, embedded_color=True)
                    bbox = stamp.getbbox()
                    if bbox:
                        stamp = stamp.crop(bbox)
                        stamp.thumbnail((key[1], key[2]), Image.Resampling.LANCZOS)
                    EMOJI_CACHE[key] = stamp
                stamp = EMOJI_CACHE[key]
                draw._image.paste(stamp, (x, y), stamp)
            else:
                if text.rstrip(" "):
                    draw.text((x, y), text.rstrip(" "), font=face, fill=TEXT)
            column += cells


def draw_card(card, width, height, size):
    image = Image.new("RGB", (width, height), PANEL)
    draw = ImageDraw.Draw(image)
    heading = f"{SPECIES.get(card['species'], card['species'])} | {card['stage']}"
    draw.text((10, 7), heading, font=font(11), fill=ACCENT)
    label = card["label"]
    # Emoji labels are also printed inside the source-rendered interface cards.
    label = "".join(c for c in label if ord(c) < 0x1F000 and c not in "\ufe0f\u200d♂").strip()
    if card.get("frames"):
        label += " | f" + ",".join(str(i) for i in card["frames"])
    available = max(10, int((width - 20) / font(10).getlength("M")))
    if len(label) > available:
        label = label[:available - 3] + "..."
    draw.text((10, 23), label, font=font(10), fill=MUTED)
    text_lines(card, draw, (10, 42), size)
    return indexed(image)


def dimensions(cards, size, minimum_width=0):
    columns = max((sum(run["cells"] for run in row) for c in cards for row in c["lines"]), default=1)
    rows = max((len(c["lines"]) for c in cards), default=1)
    return max(minimum_width, math.ceil(columns * font(size).getlength("M")) + 24), rows * (size + 2) + 56


def sheet(path, title, description, groups, target_width, size):
    layouts = []
    height = 110
    for heading, cards in groups:
        if not cards:
            continue
        cw, ch = dimensions(cards, size)
        columns = max(1, (target_width - 32) // (cw + 8))
        rows = math.ceil(len(cards) / columns)
        layouts.append((heading, cards, cw, ch, columns, height))
        height += 46 + rows * (ch + 8) + 16
    if target_width >= 32768 or height >= 32768:
        raise SystemExit(f"Atlas exceeds the 32767-pixel dimension budget: {target_width}x{height}")
    canvas = Image.new("P", (target_width, height), 0)
    canvas.putpalette(PALETTE.getpalette())
    draw = ImageDraw.Draw(canvas)
    draw.text((20, 17), title, font=font(24), fill=4)
    draw.text((20, 55), description, font=font(12), fill=3)
    draw.text((20, 78), "Real source artwork • neutral mood unless labeled • identical frames merged • fN = N × 250 ms (idle: 400 ms)", font=font(11), fill=3)
    for heading, cards, cw, ch, columns, y0 in layouts:
        draw.text((20, y0 + 8), f"{heading} — {len(cards)} cards", font=font(17), fill=4)
        for index, card in enumerate(cards):
            x = 16 + (index % columns) * (cw + 8)
            y = y0 + 42 + (index // columns) * (ch + 8)
            canvas.paste(draw_card(card, cw, ch, size), (x, y))
    path.parent.mkdir(parents=True, exist_ok=True)
    canvas.save(path, optimize=True)
    print(f"{path.relative_to(ROOT)}: {target_width}x{height}, {path.stat().st_size:,} bytes", flush=True)
    return {"file": path.name, "width": target_width, "height": height, "bytes": path.stat().st_size}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path, default=ROOT / "docs/gallery")
    args = parser.parse_args()
    with tempfile.TemporaryDirectory(prefix="ttymagotchi-gallery-") as temporary:
        source = Path(temporary) / "gallery.json"
        subprocess.run(["go", "run", "-tags", "artgallery", "./cmd/artgallery", "-out", str(source)], cwd=ROOT, check=True)
        data = json.loads(source.read_text())
    cards = data["cards"]
    groups = [(section, [c for c in cards if c["section"] == section]) for section in ["Growth", "Compact", "Expressions", "Idle", "Care", "Food", "Legacy", "Games", "Dialogs"]]
    output = args.output.resolve()
    # Indexed PNGs keep the full atlas memory bounded despite its many frames.
    artifacts = [sheet(output / "all-art.png", "TTYMAGOTCHI — COMPLETE ASCII & ANIMATION ATLAS", "5 pets × 10 growth forms • every expression • every distinct frame • all food icons • minigames • refusals • legacy artwork", groups, 18000, 9)]
    for species in data["species"]:
        subset = [(section, [c for c in selected if c["species"] == species]) for section, selected in groups]
        artifacts.append(sheet(output / f"{species}.png", f"TTYMAGOTCHI — {SPECIES[species].upper()}", "Complete species sheet • ten growth stages • all distinct frames and expressions", subset, 6000, 11))
    samples = []
    for label, stage in [("Eating", "adult"), ("Sleeping", "adult"), ("Washing", "adult"), ("Diaper", "baby"), ("Brushing", "adult"), ("Refusal (sleep / full stomach)", "adult")]:
        selected = [c for c in cards if c["section"] == "Care" and c["species"] == "cat" and c["stage"] == stage and c["label"] == label]
        samples += selected[:1] + selected[-1:]
    overview = [
        ("Every growth form — species stays hidden in the shared egg", [c for c in cards if c["section"] == "Growth"]),
        ("Expressions — delighted, content, neutral, worried, sad, tearful and dead", [c for c in cards if c["section"] == "Expressions" and c["stage"] == "adult" and c["species"] == "cat"]),
        ("Care examples — first and last distinct frame (complete sequences in the full atlas)", samples),
        ("Eight different minigames", [c for c in cards if c["section"] == "Games" and c["stage"] == "playing"]),
        ("No! — sleep refusal and stop / feed-anyway decisions", [c for c in cards if c["section"] == "Dialogs"]),
    ]
    artifacts.append(sheet(output / "overview.png", "TTYMAGOTCHI — ART REVIEW OVERVIEW", "Quick review of all 50 growth forms, six moods + death, care animations, eight games and refusal dialogs", overview, 4800, 12))
    adults = [c for c in cards if c["section"] == "Growth" and c["stage"] == "adult"]
    artifacts.append(sheet(output / "adult-portraits.png", "TTYMAGOTCHI — DETAILED ADULT PORTRAITS", "Five species in the same 58 × 34 terminal-cell frame • shared ground line • full-size terminal artwork", [("Cat, dog, rabbit, guinea pig and penguin", adults)], 3200, 17))
    compact = [c for c in cards if c["section"] == "Compact" and c["stage"] == "adult"]
    artifacts.append(sheet(output / "compact-portraits.png", "TTYMAGOTCHI — COMPACT ADULT PORTRAITS", "Small-terminal versions • C toggles compact/detail in the live preview", [("All five compact adult forms", compact)], 2400, 18))
    review = []
    for species in data["species"]:
        expressions = [c for c in cards if c["section"] == "Expressions" and c["species"] == species and c["stage"] == "adult"]
        review.append((f"{SPECIES[species]} — Happiness 100 / 70 / 50 / 30 / 10 / 0, then death", expressions))
    artifacts.append(sheet(output / "character-review.png", "TTYMAGOTCHI — FIVE PETS & EVERY EXPRESSION", "Animal portraits with individual anatomy • six Happiness levels and death • expressions also apply during care", review, 4800, 18))
    growth = [(f"{SPECIES[s]} — ten individually drawn growth forms", [c for c in cards if c["section"] == "Growth" and c["species"] == s]) for s in data["species"]]
    artifacts.append(sheet(output / "growth-review.png", "TTYMAGOTCHI — GROWTH, POSTURE & ANATOMY", "Lying newborns, short young legs and mature animal silhouettes • the egg stays a mystery", growth, 6200, 16))
    # A real animation sample, assembled from the same exported game frames.
    motion_rows = [("Young & happy", "baby", "Happy idle"), ("Adult & happy", "adult", "Happy idle"), ("Crying", "adult", "Crying idle"), ("Tired / yawning", "adult", "Sleepy idle")]
    idle = [c for c in cards if c["section"] == "Idle"]
    cw, ch = dimensions(idle, 15, 280)
    width = 32 + len(data["species"]) * (cw + 8)
    height = 92 + len(motion_rows) * (ch + 40)
    movie = []
    for tick in range(24):
        canvas = Image.new("P", (width, height), 0)
        canvas.putpalette(PALETTE.getpalette())
        draw = ImageDraw.Draw(canvas)
        draw.text((20, 16), "TTYMAGOTCHI — LIVE ANIMAL BEHAVIOURS", font=font(21), fill=4)
        draw.text((20, 48), "Actual game frames • calm pauses, blinking, tails, ears, waddles, tears and yawns", font=font(12), fill=3)
        for row_index, (heading, stage, label) in enumerate(motion_rows):
            y = 85 + row_index * (ch + 40)
            draw.text((20, y), heading, font=font(16), fill=4)
            for col, species in enumerate(data["species"]):
                source = next(c for c in idle if c["species"] == species and c["stage"] == stage and c["label"].startswith(label) and tick in c.get("frames", []))
                card = dict(source, frames=[])
                canvas.paste(draw_card(card, cw, ch, 15), (16 + col * (cw + 8), y + 28))
        movie.append(canvas)
    motion_path = output / "behaviours.gif"
    movie[0].save(motion_path, save_all=True, append_images=movie[1:], duration=400, loop=0, disposal=2, optimize=False)
    artifacts.append({"file": motion_path.name, "width": width, "height": height, "frames": Image.open(motion_path).n_frames, "frame_samples": 24, "duration_ms": 9600, "bytes": motion_path.stat().st_size})
    print(f"{motion_path.relative_to(ROOT)}: {width}x{height}, 24 frames, {motion_path.stat().st_size:,} bytes", flush=True)
    digest = hashlib.sha256()
    for path in sorted((ROOT / "internal/app").glob("*.go")):
        if not path.name.endswith("_test.go"):
            digest.update(path.name.encode())
            digest.update(path.read_bytes())
    manifest = {"renderer_source_sha256": digest.hexdigest(), "cards": len(cards), "frames_covered": sum(len(c.get("frames", [])) for c in cards), "sections": dict(Counter(c["section"] for c in cards)), "artifacts": artifacts}
    (output / "manifest.json").write_text(json.dumps(manifest, indent=2) + "\n")


if __name__ == "__main__":
    main()
