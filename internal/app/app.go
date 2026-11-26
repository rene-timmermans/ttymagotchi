package app

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rene-timmermans/ttymagotchi/internal/config"
	"github.com/rivo/tview"
)

const defaultUpdateInterval = 8 * time.Hour / 100

// App contains the tview application, the layout for the display, and the loaded config
type App struct {
	TApp              *tview.Application
	TLayout           *tview.Flex
	Config            *config.Config
	viewsList         map[string]*tview.List
	infoView          *tview.TextView
	spriteView        *tview.TextView
	currentTamagotchi *Tamagotchi
	gameEvents        []GameEvent
	modal             *tview.Modal
	activity          *activity
	done              chan struct{}

	stateMu         sync.RWMutex
	eventsMu        sync.Mutex
	uiMu            sync.Mutex
	tuiRunning      bool
	activityActive  bool
	uiReady         bool
	uiReadyOnce     sync.Once
	needsRefresh    bool
	timeAccumulator time.Duration
	updateInterval  time.Duration
}

// NewApp returns an instance of the application, initialized with the provided config
func NewApp(cfg *config.Config) *App {
	cfg.App.Language = normalizeLanguage(cfg.App.Language)
	app := &App{
		TApp:           tview.NewApplication(),
		done:           make(chan struct{}),
		Config:         cfg,
		viewsList:      make(map[string]*tview.List),
		gameEvents:     make([]GameEvent, 0),
		updateInterval: hungerDuration(cfg.App.SpeedPercent) / 100,
	}

	app.initializeStateFromConfig()

	pages, info := app.getPagesInfo()

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(pages, 0, 1, true).
		AddItem(info, 3, 1, false)

	app.TApp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape || event.Key() == tcell.KeyCtrlS {
			app.returnToStatus()
			return nil
		}
		if app.activity != nil {
			return app.activityInput(event)
		}
		switch event.Key() {
		case tcell.KeyCtrlH:
			app.goToSection(helpSection, info)
		case tcell.KeyCtrlF:
			app.goToSection(feedSection, info)
		case tcell.KeyCtrlP:
			app.goToSection(playSection, info)
		case tcell.KeyCtrlL:
			app.goToSection(sleepSection, info)
		case tcell.KeyCtrlY:
			app.goToSection(hygieneSection, info)
		case tcell.KeyCtrlO:
			app.goToSection(optionsSection, info)
		case tcell.KeyCtrlE:
			app.goToSection(eventsSection, info)
		case tcell.KeyCtrlR:
			app.showRestartModal()
		}
		return event
	})

	app.TLayout = layout

	// Start the game loop
	go app.gameLoop()

	return app
}

func (a *App) initializeStateFromConfig() {
	if a.Config.App.LastLogin.IsZero() {
		a.currentTamagotchi = newDefaultTamagotchi(randomName())
		a.updateConfigFromState()
	} else {
		cfg := a.Config.Tamagotchi
		a.currentTamagotchi = &Tamagotchi{
			Diaper: cfg.Diaper, LastChange: cfg.LastChange,
			DiaperDayTicks: cfg.DiaperDayTicks, DiaperEventCount: cfg.DiaperEventCount, DiaperNextEvent: cfg.DiaperNextEvent, DiaperEvents: cfg.DiaperEvents,
			Species:      cfg.Species,
			HealthyTicks: cfg.HealthyTicks,
			Cleanliness:  cfg.Cleanliness, Grooming: cfg.Grooming, GroomTicks: cfg.GroomTicks, CareTicks: cfg.CareTicks, LastWash: cfg.LastWash, LastBrush: cfg.LastBrush,
			Name:      cfg.Name,
			Age:       cfg.Age,
			AgeTicks:  initialAgeTicks(cfg, a.Config.App.LastLogin),
			Hunger:    cfg.Hunger,
			Happiness: cfg.Happiness,
			Health:    cfg.Health,
			Energy:    cfg.Energy,
			Weight:    cfg.Weight,
			Stage:     cfg.Stage,
			Created:   cfg.Created,
			LastFed:   cfg.LastFed,
			LastPlay:  cfg.LastPlay,
			LastSleep: cfg.LastSleep,
			IsAlive:   cfg.IsAlive,
		}
	}

	if !knownSpecies(a.currentTamagotchi.Species) {
		a.currentTamagotchi.Species = randomSpecies()
	}
	a.currentTamagotchi.Age = int(a.currentTamagotchi.AgeTicks / ageTicksPerDay)
	a.updateStageLocked(a.currentTamagotchi.Stage)
	ensureDiaperSchedule(a.currentTamagotchi)
	a.timeAccumulator = maxDuration(0, minDuration(a.updateInterval-1, a.Config.App.TickRemainder))
	a.checkDeathLocked()
	a.applyOfflineProgress()
}

func (a *App) Run() {
	a.uiMu.Lock()
	a.tuiRunning = true
	a.uiMu.Unlock()

	a.TApp.SetAfterDrawFunc(func(screen tcell.Screen) {
		a.markUIReady()
		a.TApp.SetAfterDrawFunc(nil)
	})

	app := a.TApp.SetRoot(a.TLayout, true).EnableMouse(true)

	defer func() {
		close(a.done)
		a.uiMu.Lock()
		a.tuiRunning = false
		a.uiMu.Unlock()

		now := time.Now()
		a.Config.App.LastLogin = now
		a.Config.App.CurrentLogin = now

		a.updateConfigFromState()

		if err := config.SaveConfig(a.Config); err != nil {
			log.Printf("failed to save config: %v", err)
		}
	}()

	if err := app.Run(); err != nil {
		panic(err)
	}
}

func (a *App) showRestartModal() {
	if a.modal != nil {
		return // Modal already showing
	}

	modal := tview.NewModal().
		SetText(a.tr("Are you sure you want to restart?\n\nThis will reset your tamagotchi to a new egg.\nAll progress will be lost!")).
		AddButtons([]string{a.tr("Cancel"), a.tr("Restart")}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == a.tr("Restart") {
				a.restartTamagotchi()
			}
			a.TApp.SetRoot(a.TLayout, true).SetFocus(a.TLayout)
			a.modal = nil
		})

	a.modal = modal
	a.TApp.SetRoot(modal, true).SetFocus(modal)
}

func (a *App) restartTamagotchi() {
	newName := randomName()
	newTamagotchi := newDefaultTamagotchi(newName)

	a.stateMu.Lock()
	a.currentTamagotchi = newTamagotchi
	a.timeAccumulator = 0
	a.stateMu.Unlock()

	a.eventsMu.Lock()
	a.gameEvents = make([]GameEvent, 0)
	a.eventsMu.Unlock()

	a.addLocalizedEvent("RESTART", "Started a new tamagotchi named %s! 🥚", newName)
	a.updateConfigFromState()
}

func (a *App) gameLoop() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	lastTick := time.Now()
	for {
		var now time.Time
		select {
		case <-a.done:
			return
		case now = <-ticker.C:
		}
		elapsed := now.Sub(lastTick)
		lastTick = now
		if elapsed < 0 {
			elapsed = 0
		}
		a.advanceTime(elapsed)
		a.refreshUI()
	}
}

func (a *App) advanceTime(elapsed time.Duration) bool {
	if elapsed <= 0 {
		return false
	}

	var ticks int

	a.stateMu.Lock()
	if a.currentTamagotchi != nil {
		a.timeAccumulator += elapsed
		ticks = int(a.timeAccumulator / a.updateInterval)
		if ticks > 0 {
			a.timeAccumulator -= time.Duration(ticks) * a.updateInterval
			for i := 0; i < ticks; i++ {
				a.applyTickLocked()
				if a.currentTamagotchi == nil || !a.currentTamagotchi.IsAlive {
					break
				}
			}
		}
	}
	a.stateMu.Unlock()

	if ticks > 0 {
		a.requestRefresh()
		return true
	}

	return false
}

func (a *App) applyTickLocked() {
	if a.currentTamagotchi == nil || !a.currentTamagotchi.IsAlive {
		return
	}

	t := a.currentTamagotchi

	if t.AgeTicks < maxAgeTicks {
		t.AgeTicks++
	}
	t.Age = int(t.AgeTicks / ageTicksPerDay)
	a.updateStageLocked(t.Stage)
	t.Hunger = min(100, t.Hunger+1)
	t.Energy = max(0, t.Energy-1)
	t.CareTicks = (t.CareTicks + 1) % 6
	if t.CareTicks%3 == 0 {
		t.Cleanliness = max(0, t.Cleanliness-1)
	}
	if needsDiaper(*t) {
		a.advanceDiaperLocked()
	} else {
		t.GroomTicks = (t.GroomTicks + 1) % 21
		if t.GroomTicks == 0 {
			t.Grooming = max(0, t.Grooming-1)
		}
	}

	if t.Hunger >= 80 || t.Cleanliness <= 20 || ageCareLevel(*t) <= 20 || t.Energy < 20 {
		t.Happiness = max(0, t.Happiness-1)
	}
	damage := 0
	if t.Hunger >= 90 {
		damage += 2
	}
	damage += hygieneDamage(t.Cleanliness)
	damage += hygieneDamage(ageCareLevel(*t))
	t.Health = max(0, t.Health-damage)
	a.checkDeathLocked()
	if !t.IsAlive {
		return
	}

	a.applyRecoveryLocked()
}

func (a *App) updateStageLocked(previousStage string) {
	if a.currentTamagotchi == nil {
		return
	}

	t := a.currentTamagotchi
	level := 0
	for i, day := range growthDays {
		if t.Age >= day {
			level = i
		}
	}
	wasInDiapers := needsDiaper(*t)
	t.Stage = growthStages[level]
	if wasInDiapers && !needsDiaper(*t) {
		t.Grooming = 100
		t.GroomTicks = 0
	}

	if previousStage != a.currentTamagotchi.Stage {
		a.addLocalizedEvent("EVOLUTION", "Your tamagotchi evolved to %s! 🎉", translatedArgument(a.currentTamagotchi.Stage))
	}
}

func (a *App) applyOfflineProgress() {
	if a.Config.App.LastLogin.IsZero() {
		return
	}

	elapsed := a.Config.App.CurrentLogin.Sub(a.Config.App.LastLogin)
	if elapsed <= 0 {
		return
	}

	changed := a.advanceTime(elapsed)
	a.Config.App.LastLogin = a.Config.App.CurrentLogin
	a.updateConfigFromState()
	if changed {
		a.addLocalizedEvent("PROGRESS", "Time passed while you were away: %s.", durationArgument(elapsed))
	}
}

func (a *App) requestRefresh() {
	a.uiMu.Lock()
	a.needsRefresh = true
	ready := a.uiReady
	running := a.tuiRunning
	a.uiMu.Unlock()

	if ready && running {
		go a.refreshUI()
	}
}

func (a *App) refreshUI() {
	a.uiMu.Lock()
	if !a.tuiRunning || !a.uiReady || a.TApp == nil {
		a.uiMu.Unlock()
		return
	}
	refresh := a.needsRefresh
	a.needsRefresh = false
	a.uiMu.Unlock()

	a.TApp.QueueUpdateDraw(func() {
		a.tickActivity(time.Now())
		if a.activity == nil && a.modal == nil && a.spriteView != nil {
			name, _ := a.TLayout.GetItem(0).(*tview.Pages).GetFrontPage()
			if name == statusSection {
				a.updateSpriteView(a.spriteView)
			}
		}
		if !refresh {
			return
		}
		if list := a.viewsList["status"]; list != nil {
			a.generateStatusList(list)
		}
		if a.spriteView != nil {
			a.updateSpriteView(a.spriteView)
		}
		if list := a.viewsList["feed"]; list != nil {
			a.generateFeedList(list)
		}
		if list := a.viewsList["play"]; list != nil {
			a.generatePlayList(list)
		}
		if list := a.viewsList["sleep"]; list != nil {
			a.generateSleepList(list)
		}
		if list := a.viewsList["hygiene"]; list != nil {
			a.generateHygieneList(list)
		}
		if list := a.viewsList["options"]; list != nil {
			a.generateOptionsList(list)
		}
		if list := a.viewsList["events"]; list != nil {
			a.generateEventsList(list)
		}
	})
}

func (a *App) updateConfigFromState() {
	a.stateMu.RLock()
	if a.currentTamagotchi == nil {
		a.stateMu.RUnlock()
		return
	}
	t := *a.currentTamagotchi
	remainder := a.timeAccumulator
	a.stateMu.RUnlock()

	a.Config.App.TickRemainder = remainder
	a.Config.Tamagotchi.Species = t.Species
	a.Config.Tamagotchi.HealthyTicks = t.HealthyTicks
	a.Config.Tamagotchi.Cleanliness = t.Cleanliness
	a.Config.Tamagotchi.DiaperDayTicks = t.DiaperDayTicks
	a.Config.Tamagotchi.DiaperEventCount = t.DiaperEventCount
	a.Config.Tamagotchi.DiaperNextEvent = t.DiaperNextEvent
	a.Config.Tamagotchi.DiaperEvents = t.DiaperEvents
	a.Config.Tamagotchi.Diaper = t.Diaper
	a.Config.Tamagotchi.LastChange = t.LastChange
	a.Config.Tamagotchi.Grooming = t.Grooming
	a.Config.Tamagotchi.GroomTicks = t.GroomTicks
	a.Config.Tamagotchi.CareTicks = t.CareTicks
	a.Config.Tamagotchi.LastWash = t.LastWash
	a.Config.Tamagotchi.LastBrush = t.LastBrush
	a.Config.Tamagotchi.Name = t.Name
	a.Config.Tamagotchi.Age = t.Age
	ageTicks := t.AgeTicks
	a.Config.Tamagotchi.AgeTicks = &ageTicks
	a.Config.Tamagotchi.Hunger = t.Hunger
	a.Config.Tamagotchi.Happiness = t.Happiness
	a.Config.Tamagotchi.Health = t.Health
	a.Config.Tamagotchi.Energy = t.Energy
	a.Config.Tamagotchi.Weight = t.Weight
	a.Config.Tamagotchi.Stage = t.Stage
	a.Config.Tamagotchi.Created = t.Created
	a.Config.Tamagotchi.LastFed = t.LastFed
	a.Config.Tamagotchi.LastPlay = t.LastPlay
	a.Config.Tamagotchi.LastSleep = t.LastSleep
	a.Config.Tamagotchi.IsAlive = t.IsAlive
}

func (a *App) tamagotchiSnapshot() (Tamagotchi, bool) {
	a.stateMu.RLock()
	defer a.stateMu.RUnlock()

	if a.currentTamagotchi == nil {
		return Tamagotchi{}, false
	}

	return *a.currentTamagotchi, true
}

func (a *App) eventsSnapshot() []GameEvent {
	a.eventsMu.Lock()
	defer a.eventsMu.Unlock()

	if len(a.gameEvents) == 0 {
		return nil
	}

	events := make([]GameEvent, len(a.gameEvents))
	copy(events, a.gameEvents)
	return events
}

func newDefaultTamagotchi(name string) *Tamagotchi {
	now := time.Now()
	pet := &Tamagotchi{
		Species:     randomSpecies(),
		Cleanliness: 100, Diaper: 100, Grooming: 100,
		Name:      name,
		Age:       0,
		Hunger:    50,
		Happiness: 50,
		Health:    100,
		Energy:    100,
		Weight:    50.0,
		Stage:     "egg",
		Created:   now,
		LastFed:   now,
		LastPlay:  now,
		LastSleep: now,
		IsAlive:   true,
	}
	newDiaperSchedule(pet)
	return pet
}

func formatDuration(d time.Duration) string {
	if d >= 24*time.Hour {
		days := d / (24 * time.Hour)
		hours := (d % (24 * time.Hour)) / time.Hour
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	if d >= time.Hour {
		hours := d / time.Hour
		minutes := (d % time.Hour) / time.Minute
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	if d >= time.Minute {
		minutes := d / time.Minute
		seconds := (d % time.Minute) / time.Second
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	seconds := d / time.Second
	if seconds <= 0 {
		seconds = 1
	}
	return fmt.Sprintf("%ds", seconds)
}

func (a *App) addGameEvent(eventType, message string) {
	a.addGameEventRecord(GameEvent{Type: eventType, Message: message, Timestamp: time.Now()})
}
func (a *App) addGameEventRecord(event GameEvent) {
	a.eventsMu.Lock()
	a.gameEvents = append(a.gameEvents, event)

	if len(a.gameEvents) > 50 {
		a.gameEvents = a.gameEvents[len(a.gameEvents)-50:]
	}
	a.eventsMu.Unlock()

	a.requestRefresh()
}

func (a *App) markUIReady() {
	a.uiReadyOnce.Do(func() {
		a.uiMu.Lock()
		a.uiReady = true
		needs := a.needsRefresh
		a.uiMu.Unlock()

		if needs {
			go a.refreshUI()
		}
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Call while holding stateMu, or before starting the game loop.
func (a *App) checkDeathLocked() {
	t := a.currentTamagotchi
	if t != nil && t.Health <= 0 {
		t.Health = 0
		t.HealthyTicks = 0
		if t.IsAlive {
			t.IsAlive = false
			a.addLocalizedEvent("DEATH", "Your tamagotchi has passed away... 💔")
		}
	}
}
func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}
