## Local interactive edition

Based on upstream v0.2.0. Source: `/home/deck/Documenten/Codex/Termagotchi`.

- New pets also receive a random name from the 100 most common Dutch first given names. Source: [Nederlandse Voornamenbank, Meertens Instituut KNAW (2022)](https://nvb.meertens.knaw.nl/downloads/Top_eerste_voornamen_NL_2022.zip), combining female and male carrier counts. Every name in the pool has equal selection probability. Existing saved names stay unchanged; a new game/restart draws a new name.
- New pets randomly become a cat, dog, rabbit, guinea pig or penguin (Tux). All start in the same mystery egg; their species is revealed only after hatching and saved across restarts. Existing saves receive a species once; their pet stats are preserved.
- Each species has ten appearances: Egg, Hatchling, Baby, Toddler, Youngster, Child, Preteen, Teen, Adolescent, Adult. They arrive at ages 0, 1, 2, 3, 4, 5, 7, 9, 12 and 14 days. Status shows the current stage out of ten.
- Pets and care scenes use larger, detailed terminal drawings (roughly four times the former figure area), with blinking, tails, ears or feet moving on Status. Food and washing effects use each species’ face and body positions.
- Diaper care is required in stages 1–3 (Egg, Hatchling, Baby); from stage 4 (Toddler) it is replaced by Brushing. Only the active age-specific bar appears and counts for Hygiene, Happiness, health damage and recovery. Diapers change only at four or five randomly timed bowel events per normal game day, dropping 35 points each (100 → 65 → 30 → 0); Brushing loses a point every 21 periods (100 → 0 in seven days at 50% speed). The transition starts Brushing at 100% with no old diaper penalty. The daily event count, randomized schedule, progress and next event persist across restarts and offline time; changing a diaper does not delay bowel events. The rhythm scales with speed. Both care actions animate the pet and persist separately.
- Each of the eight Play actions has a distinct three-round minigame: Play Ball uses a timing marker; Run Around dodges six lane obstacles; Sing Songs follows notes with 1–3; Draw Pictures traces an outline; Solve Puzzle repeats a hidden arrow sequence; Dance Party follows visible arrows; Read Books identifies the first letter of a localized word; Hide & Seek tracks a box through swaps. Scores determine rewards.
- Press Enter to begin. Each minigame explains its own keys: Space, arrows/WASD or 1–3. A round has a time limit; Esc cancels unfinished games without changing stats.
- Happiness, health and weight loss scale with the number of successful rounds. Completing a game spends its listed energy even on a zero score. The list shows maximum rewards.
- Esc and Ctrl+S always return directly to Status, including from activities and the restart dialog. An unfinished game is cancelled without rewards; pending food/care is applied once. Enter returns from a result to its activity menu. Ctrl+C exits and saves normally.
- Feeding shows a three-second animation with approaching food, chewing and a result showing stat changes. Esc finishes the meal and returns to Status; each meal is applied exactly once.
- The pet itself animates in every life stage: feeding opens and closes its mouth as food approaches its face; sleeping closes its eyes and moves its breathing body with rising Zzz; washing moves a sponge over its body in a bubbly bath; Brushing moves the brush through the coat. Esc finishes care once and returns to Status.
- Six Happiness expressions run from delighted through content and neutral to worried, sad and tearful. The same expression is used during care and on result screens. Neglected hygiene, hunger or exhaustion gradually lowers Happiness too.
- Every status bar displays changes of 1% using partial blocks.
- Ctrl+Y opens Hygiene. Cleanliness and the age-specific care need decay independently. At 20% and below each causes 1–5 health damage per decay tick, increasing toward zero. The overall Hygiene bar shows the lowest active need. Washing restores cleanliness; brushing restores coat care. These stop neglect damage but do not directly heal existing damage.
- Excess nutrition damages health by half the excess (minimum 5). Feeding at hunger 0 therefore always causes damage. Hunger at 90 or higher causes 2 health damage per decay tick. Health 0 immediately means death, which ordinary care cannot reverse.
- Sustained good care restores +1 health every five decay periods (24 minutes at normal 50% speed). Happiness, Energy, Cleanliness and the active Diaper/Brushing bar must all stay at least 40%, with Hunger at most 60%. Any lapse or overfeeding resets recovery progress. This can heal an ill pet below 40 health, caps at 100, scales with speed and also works offline. Sleep and minigames retain their immediate health rewards.
- Ctrl+O opens Options with English, Français, Español and Deutsch. Choose a language with Enter, or use Left/Right on the language rows to cycle. Menus, instructions, care results, dialogs, species/stage names, Help and existing event-log entries switch immediately. Pet names stay intact, and the language saves immediately; older saves default to English. Translation catalogs live in `internal/app/locales/`.
- Pets refuse to sleep above 60% Energy, shaking their head with a localized red “No!” animation. Refusal grants no energy, health or happiness and does not change the last-sleep timestamp. At exactly 60% or below sleep works normally.
- Ctrl+O opens Options. Left/Right or +/- adjusts speed by 1%, Page Up/Down by 10%; Enter selects a preset. Speed 0% means one week, 50% (default) eight hours, and 100% five minutes for hunger 0 → 100. Intermediate values use a smooth geometric scale. Energy, hygiene and neglect scale with this speed; animations remain short. Offline time counts at the saved speed.
- Settings save immediately. Hygiene values, care timestamps, recovery progress and fractional decay progress persist. Older saves start with clean hygiene and 50% speed; existing pet stats are preserved.

Build and install for this user:

```sh
go test -race ./...
go vet ./...
go build -o bin/termagotchi ./cmd/termagotchi
install -Dm755 bin/termagotchi /home/deck/.local/bin/termagotchi
```

The end-to-end terminal check uses an isolated save: `python3 scripts/smoke_test.py bin/termagotchi adult en` (also accepts egg/hatchling/baby and fr/es/de).

Run `termagotchi` in a new terminal. The user binary takes precedence over the AUR binary; `/usr/bin/termagotchi` still runs the original version. Avoid running both against the same save at once.

---

# Termagotchi 🐾

A terminal-based Tamagotchi simulation written in Go, featuring a simple and hacky TUI interface.

## Features

- 🎮 **Interactive TUI**: Beautiful terminal user interface with keyboard navigation
- 🐾 **Digital Pet Care**: Feed, play, and put your tamagotchi to sleep
- 📊 **Real-time Stats**: Monitor hunger, happiness, health, and energy levels
- 🔄 **Life Stages**: Watch your tamagotchi evolve from egg to adult
- 📝 **Event History**: Track all interactions and milestones
- 💾 **Auto-save**: Progress is automatically saved to your config directory
- ⏰ **Time-based Mechanics**: Stats change over time, requiring regular care
- 🔄 **Restart Feature**: Reset to a new tamagotchi with confirmation modal

## Installation

### Homebrew (recommended, after release)

```sh
brew tap ezeoleaf/tap
brew install termagotchi
```

### Download a Release

- Go to [Releases](https://github.com/ezeoleaf/termagotchi/releases) and download the binary for your OS.
- Unpack and move it to a directory in your `$PATH` (e.g., `/usr/local/bin`).

### Build from Source

#### Prerequisites

- Go 1.24.0 or later

```bash
git clone <repository-url>
cd termagotchi
go mod tidy
go build -o termagotchi cmd/termagotchi/main.go
./termagotchi
```

## Usage

### Controls

- **Ctrl+S**: Status - View tamagotchi stats
- **Ctrl+F**: Feed - Give food to tamagotchi
- **Ctrl+P**: Play - Play games with tamagotchi
- **Ctrl+L**: Sleep - Put tamagotchi to sleep
- **Ctrl+E**: Events - View game history
- **Ctrl+H**: Help - Show help page
- **Ctrl+R**: Restart - Reset tamagotchi to new egg
- **Ctrl+C**: Quit - Exit the game

### Navigation

- Use arrow keys to navigate lists
- Press Enter to select items
- Use Ctrl+key shortcuts for quick access

## Game Mechanics

### Stats

- **Hunger**: 0 = Full, 100 = Starving
- **Happiness**: 0 = Very Sad, 100 = Very Happy
- **Health**: 0 = Sick, 100 = Healthy
- **Energy**: 0 = Tired, 100 = Energetic

### Life Stages

1. **Egg** (0-1 days)
2. **Baby** (1-3 days)
3. **Child** (3-7 days)
4. **Teen** (7-14 days)
5. **Adult** (14+ days)

### Food Types

- 🍎 Apple: Good nutrition, low weight gain
- 🍕 Pizza: High nutrition and happiness
- 🥗 Salad: Healthy option
- 🍔 Burger: High nutrition but heavy
- 🍦 Ice Cream: High happiness boost
- 🥕 Carrot: Balanced nutrition
- 🍫 Chocolate: Happiness boost
- 🥩 Steak: Maximum nutrition

### Games

- 🎾 Play Ball: Classic fun
- 🏃‍♂️ Run Around: Good exercise
- 🎵 Sing Songs: Low energy, high happiness
- 🎨 Draw Pictures: Creative fun
- 🧩 Solve Puzzle: Mental stimulation
- 🎭 Dance Party: High energy fun
- 📚 Read Books: Educational
- 🎪 Play Hide & Seek: Interactive fun

### Sleep Options

- 😴 Short Nap (30 min): Quick energy boost
- 😪 Medium Sleep (2 hours): Balanced rest
- 😴 Long Sleep (6 hours): Good recovery
- 😴 Full Night (8 hours): Complete restoration

## Tips for Success

1. **Feed Regularly**: Keep hunger below 80 to maintain happiness
2. **Play Often**: Increase happiness and health through games
3. **Rest When Needed**: Put to sleep when energy is low
4. **Monitor Health**: Low health can lead to death
5. **Balance Stats**: Keep all stats in good ranges
6. **Restart When Needed**: Use Ctrl+R to start fresh if your tamagotchi dies

## Configuration

The game automatically saves your progress to:

- **macOS**: `~/Library/Application Support/termagotchi/config.yml`
- **Linux**: `~/.config/termagotchi/config.yml`
- **Windows**: `%APPDATA%\termagotchi\config.yml`

## Screenshots

Home

<img width="782" height="450" alt="image" src="https://github.com/user-attachments/assets/39cf86c2-a318-4140-8e07-ba252d38f3c3" />

Feed

<img width="785" height="448" alt="image" src="https://github.com/user-attachments/assets/58d7897f-c55a-42e0-8a5f-072ebce347e4" />

Play

<img width="782" height="450" alt="image" src="https://github.com/user-attachments/assets/edffe7be-4c00-4e03-8c4c-ca6e2ce6e2af" />


Help

<img width="1025" height="647" alt="image" src="https://github.com/user-attachments/assets/ccad0600-d168-4d6f-bb90-10fcd366d78f" />

## Development

### Project Structure

```
termagotchi/
├── cmd/
│   └── termagotchi/
│       └── main.go
├── internal/
│   ├── app/
│   │   ├── app.go
│   │   ├── structs.go
│   │   ├── pages.go
│   │   ├── status.go
│   │   ├── feed.go
│   │   ├── play.go
│   │   ├── sleep.go
│   │   ├── events.go
│   │   └── help.go
│   └── config/
│       └── config.go
├── go.mod
├── go.sum
└── README.md
```

### Dependencies

- `github.com/gdamore/tcell/v2`: Terminal UI framework
- `github.com/rivo/tview`: TUI components
- `gopkg.in/yaml.v3`: Configuration file handling

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Acknowledgments

- Inspired by the classic Tamagotchi digital pet
- Built with the amazing `tview` TUI framework
- Thanks to the Go community for excellent libraries
