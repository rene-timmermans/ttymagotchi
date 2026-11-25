package app

import (
	"github.com/rivo/tview"
)

func (a *App) generateHelpList(listHelp *tview.List) {
	defer preserveSelection(listHelp)()
	listHelp.Clear()
	for _, line := range helpText {
		listHelp.AddItem(a.tr(line), "", 0, nil)
	}
}

func (a *App) helpPage() (title string, content tview.Primitive) {
	listHelp := a.viewsList["help"]
	if listHelp == nil {
		listHelp = getList()
		a.viewsList["help"] = listHelp
	}

	a.generateHelpList(listHelp)

	title = helpSection
	return title, tview.NewFlex().
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(listHelp, 0, 1, true), 0, 1, true)
}

var helpText = []string{
	"=== TERMAGOTCHI HELP ===",
	"CARE & HEALTH",
	"Hunger: 0 = Full, 100 = Starving",
	"Happiness: 0 = Very Sad, 100 = Very Happy",
	"Health: 0 = Dead, 100 = Healthy",
	"Energy: 0 = Tired, 100 = Energetic",
	"Overfeeding damages health; hunger at 90+ also causes damage.",
	"Diaper: stages 1–3. Brushing: stage 4 onward (~7 days at 50%).",
	"Hygiene <= 20 damages health; closer to 0 means more damage.",
	"Good care restores +1 health every 24 min at 50% speed.",
	"Keep Happiness, Energy and active Hygiene needs >= 40%.",
	"Keep Hunger <= 60%; a lapse resets the recovery timer.",
	"Above 60% energy, your pet refuses to sleep.",
	"GROWTH",
	"Egg → Hatchling → Baby → Toddler → Youngster → Child",
	"→ Preteen → Teen → Adolescent → Adult (10 forms)",
	"All species start in the same mystery egg.",
	"MINIGAMES & FEEDING",
	"Each Play activity has its own minigame; score scales rewards.",
	"Enter starts; Esc cancels unfinished games and opens Status.",
	"Use Space, arrow keys, WASD or 1–3 as shown in each game.",
	"Esc finishes food/care once and returns directly to Status.",
	"SPEED & LANGUAGE",
	"Ctrl+O: choose language and speed. Settings save immediately.",
	"Speed: 0% = 1 week, 50% = 8 hours, 100% = 5 minutes.",
	"This is the time for hunger 0 → 100; offline time counts.",
	"KEYBOARD CONTROLS",
	"Esc / Ctrl+S: Status - Return to your tamagotchi",
	"Ctrl+F: Feed - Give food to tamagotchi",
	"Ctrl+P: Play - Play games with tamagotchi",
	"Ctrl+L: Sleep - Put tamagotchi to sleep",
	"Ctrl+Y: Hygiene - Wash, Diaper (1–3) or Brushing (4+)",
	"Ctrl+O: Options - Adjust speed with Left/Right or +/-",
	"Ctrl+E: Events - View game history",
	"Ctrl+H: Help - Show this help page",
	"Ctrl+R: Restart - Reset tamagotchi to new egg",
	"Ctrl+C: Quit - Exit the game",
	"Use arrow keys and Enter to navigate menus.",
	"Your tamagotchi progress is automatically saved.",
	"Data is stored in your config directory.",
}
