package app

import (
	"github.com/rivo/tview"
)

func (a *App) generateHelpList(listHelp *tview.List) {
	listHelp.Clear()

	listHelp.AddItem("=== TERMAGOTCHI HELP ===", "", 0, nil)
	listHelp.AddItem("", "", 0, nil) // Empty line

	listHelp.AddItem("🎮 GAME OVERVIEW", "", 0, nil)
	listHelp.AddItem("Termagotchi is a terminal-based Tamagotchi simulation.", "", 0, nil)
	listHelp.AddItem("Take care of your digital pet by feeding, playing, and sleeping.", "", 0, nil)
	listHelp.AddItem("", "", 0, nil) // Empty line

	listHelp.AddItem("📊 STATS EXPLANATION", "", 0, nil)
	listHelp.AddItem("Hunger: 0 = Full, 100 = Starving", "", 0, nil)
	listHelp.AddItem("Happiness: 0 = Very Sad, 100 = Very Happy", "", 0, nil)
	listHelp.AddItem("Health: 0 = Dead, 100 = Healthy", "", 0, nil)
	listHelp.AddItem("Energy: 0 = Tired, 100 = Energetic", "", 0, nil)
	listHelp.AddItem("", "", 0, nil) // Empty line

	listHelp.AddItem("🔄 STAGES OF LIFE", "", 0, nil)
	listHelp.AddItem("Egg → Hatchling → Baby → Toddler → Youngster → Child", "", 0, nil)
	listHelp.AddItem("→ Preteen → Teen → Adolescent → Adult (10 forms)", "", 0, nil)
	listHelp.AddItem("Random cat, dog, rabbit, guinea pig or penguin; saved per pet.", "", 0, nil)
	listHelp.AddItem("All species start in the same mystery egg.", "", 0, nil)
	listHelp.AddItem("Diaper: stages 1–3. Brushing: stage 4 onward (~7 days at 50%).", "", 0, nil)
	listHelp.AddItem("Your tamagotchi evolves based on age.", "", 0, nil)
	listHelp.AddItem("", "", 0, nil) // Empty line

	listHelp.AddItem("⚡ GAME MECHANICS", "", 0, nil)
	listHelp.AddItem("• Stats change automatically over time", "", 0, nil)
	listHelp.AddItem("• Keep hunger low and happiness high", "", 0, nil)
	listHelp.AddItem("• Low health can lead to death", "", 0, nil)
	listHelp.AddItem("• Energy is needed for playing", "", 0, nil)
	listHelp.AddItem("", "", 0, nil) // Empty line

	listHelp.AddItem("MINIGAMES & FEEDING", "", 0, nil)
	listHelp.AddItem("Timing: Space in the green zone. Memory: repeat the arrows.", "", 0, nil)
	listHelp.AddItem("Follow the arrows: arrow keys or WASD. Scores scale rewards.", "", 0, nil)
	listHelp.AddItem("Enter starts; Esc cancels unfinished games and opens Status.", "", 0, nil)
	listHelp.AddItem("Esc finishes food/care once and returns directly to Status.", "", 0, nil)
	listHelp.AddItem("", "", 0, nil)
	listHelp.AddItem("CARE & SPEED", "", 0, nil)
	listHelp.AddItem("Overfeeding damages health; hunger at 90+ also causes damage.", "", 0, nil)
	listHelp.AddItem("Good care restores +1 health every 24 min at 50% speed.", "", 0, nil)
	listHelp.AddItem("Keep Happiness, Energy and active Hygiene needs >= 40%.", "", 0, nil)
	listHelp.AddItem("Keep Hunger <= 60%; a lapse resets the recovery timer.", "", 0, nil)
	listHelp.AddItem("Hygiene <= 20 damages health; closer to 0 means more damage.", "", 0, nil)
	listHelp.AddItem("Baths and Diaper/Brushing restore age-specific hygiene needs.", "", 0, nil)
	listHelp.AddItem("Sleep and hygiene animate; Enter returns from results.", "", 0, nil)
	listHelp.AddItem("Speed: 0% = 1 week, 50% = 8 hours, 100% = 5 minutes.", "", 0, nil)
	listHelp.AddItem("This is the time for hunger 0 → 100; offline time counts.", "", 0, nil)
	listHelp.AddItem("", "", 0, nil)
	listHelp.AddItem("🎯 TIPS FOR SUCCESS", "", 0, nil)
	listHelp.AddItem("• Feed regularly to prevent hunger", "", 0, nil)
	listHelp.AddItem("• Play games to increase happiness", "", 0, nil)
	listHelp.AddItem("• Put to sleep when energy is low", "", 0, nil)
	listHelp.AddItem("• Balance all stats for best health", "", 0, nil)
	listHelp.AddItem("", "", 0, nil) // Empty line

	listHelp.AddItem("⌨️  KEYBOARD CONTROLS", "", 0, nil)
	listHelp.AddItem("Esc / Ctrl+S: Status - Return to your tamagotchi", "", 0, nil)
	listHelp.AddItem("Ctrl+F: Feed - Give food to tamagotchi", "", 0, nil)
	listHelp.AddItem("Ctrl+P: Play - Play games with tamagotchi", "", 0, nil)
	listHelp.AddItem("Ctrl+L: Sleep - Put tamagotchi to sleep", "", 0, nil)
	listHelp.AddItem("Ctrl+Y: Hygiene - Wash, Diaper (1–3) or Brushing (4+)", "", 0, nil)
	listHelp.AddItem("Ctrl+O: Options - Adjust speed with Left/Right or +/-", "", 0, nil)
	listHelp.AddItem("Ctrl+E: Events - View game history", "", 0, nil)
	listHelp.AddItem("Ctrl+H: Help - Show this help page", "", 0, nil)
	listHelp.AddItem("Ctrl+R: Restart - Reset tamagotchi to new egg", "", 0, nil)
	listHelp.AddItem("Ctrl+C: Quit - Exit the game", "", 0, nil)
	listHelp.AddItem("", "", 0, nil) // Empty line

	listHelp.AddItem("🎮 NAVIGATION", "", 0, nil)
	listHelp.AddItem("• Use arrow keys to navigate lists", "", 0, nil)
	listHelp.AddItem("• Press Enter to select items", "", 0, nil)
	listHelp.AddItem("• Use Ctrl+key shortcuts for quick access", "", 0, nil)
	listHelp.AddItem("", "", 0, nil) // Empty line

	listHelp.AddItem("💾 SAVE SYSTEM", "", 0, nil)
	listHelp.AddItem("Your tamagotchi progress is automatically saved.", "", 0, nil)
	listHelp.AddItem("Data is stored in your config directory.", "", 0, nil)
	listHelp.AddItem("", "", 0, nil) // Empty line

	listHelp.AddItem("🔄 RESTART FEATURE", "", 0, nil)
	listHelp.AddItem("• Use Ctrl+R to restart with a new tamagotchi", "", 0, nil)
	listHelp.AddItem("• Confirmation modal will ask for your approval", "", 0, nil)
	listHelp.AddItem("• All progress will be reset to a fresh egg", "", 0, nil)
	listHelp.AddItem("• Useful if your tamagotchi dies or you want a fresh start", "", 0, nil)
	listHelp.AddItem("", "", 0, nil) // Empty line

	listHelp.AddItem("🐛 TROUBLESHOOTING", "", 0, nil)
	listHelp.AddItem("• If your tamagotchi dies, use Ctrl+R to restart", "", 0, nil)
	listHelp.AddItem("• Check the Events page for recent activity", "", 0, nil)
	listHelp.AddItem("• Monitor stats regularly to prevent issues", "", 0, nil)
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
