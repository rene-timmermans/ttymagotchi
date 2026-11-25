package app

import (
	"embed"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ezeoleaf/termagotchi/internal/config"
	"github.com/rivo/tview"
)

//go:embed locales/*.json
var localeFiles embed.FS

var languageCodes = []string{"en", "fr", "es", "de"}
var languageNames = map[string]string{"en": "English", "fr": "Français", "es": "Español", "de": "Deutsch"}
var translations = loadTranslations()

func loadTranslations() map[string]map[string]string {
	all := make(map[string]map[string]string)
	for _, code := range languageCodes {
		data, err := localeFiles.ReadFile("locales/" + code + ".json")
		if err != nil {
			panic(err)
		}
		catalog := make(map[string]string)
		if err := json.Unmarshal(data, &catalog); err != nil {
			panic(err)
		}
		all[code] = catalog
	}
	return all
}
func normalizeLanguage(code string) string {
	if _, ok := languageNames[code]; ok {
		return code
	}
	return "en"
}
func (a *App) tr(key string) string {
	if a.Config == nil {
		return key
	}
	if translated, ok := translations[normalizeLanguage(a.Config.App.Language)][key]; ok {
		return translated
	}
	return key
}
func (a *App) duration(d time.Duration) string {
	if d >= 24*time.Hour {
		return fmt.Sprintf(a.tr("%dd %dh"), d/(24*time.Hour), (d%(24*time.Hour))/time.Hour)
	}
	if d >= time.Hour {
		return fmt.Sprintf(a.tr("%dh %dm"), d/time.Hour, (d%time.Hour)/time.Minute)
	}
	if d >= time.Minute {
		return fmt.Sprintf(a.tr("%dm %ds"), d/time.Minute, (d%time.Minute)/time.Second)
	}
	return fmt.Sprintf(a.tr("%ds"), max(1, int(d/time.Second)))
}
func (a *App) stageText(t Tamagotchi) string {
	return fmt.Sprintf("%s (%d/10)", a.tr(t.Stage), growthIndex(t.Stage)+1)
}
func (a *App) saveSettings() {
	a.updateConfigFromState()
	a.Config.App.LastLogin = time.Now()
	a.Config.App.CurrentLogin = a.Config.App.LastLogin
	if err := config.SaveConfig(a.Config); err != nil {
		a.addLocalizedEvent("ERROR", "Could not save settings: %v", err.Error())
	}
}
func (a *App) setLanguage(code string) {
	a.Config.App.Language = normalizeLanguage(code)
	a.saveSettings()
	a.refreshLanguage()
	a.requestRefresh()
}
func (a *App) cycleLanguage(delta int) {
	current := normalizeLanguage(a.Config.App.Language)
	for i, code := range languageCodes {
		if code == current {
			a.setLanguage(languageCodes[(i+delta+len(languageCodes))%len(languageCodes)])
			return
		}
	}
}
func (a *App) refreshLanguage() {
	for name, generate := range map[string]func(*tview.List){
		"status": a.generateStatusList, "feed": a.generateFeedList, "play": a.generatePlayList,
		"sleep": a.generateSleepList, "hygiene": a.generateHygieneList, "options": a.generateOptionsList,
		"events": a.generateEventsList, "help": a.generateHelpList,
	} {
		if list := a.viewsList[name]; list != nil {
			generate(list)
		}
	}
	if a.infoView != nil {
		a.infoView.SetText(a.tr(footerText))
	}
	if a.spriteView != nil {
		a.spriteView.SetTitle(a.tr("Current Form"))
		a.updateSpriteView(a.spriteView)
	}
	if a.activity != nil {
		a.renderActivity(time.Now())
	}
}

// Typed arguments keep identifiers such as pet names intact when changing language.
type translatedArgument string
type durationArgument time.Duration

func (a *App) eventMessage(event GameEvent) string {
	if event.MessageKey == "" {
		return event.Message
	}
	args := make([]any, len(event.MessageArgs))
	for i, arg := range event.MessageArgs {
		switch value := arg.(type) {
		case translatedArgument:
			args[i] = a.tr(string(value))
		case durationArgument:
			args[i] = a.duration(time.Duration(value))
		default:
			args[i] = arg
		}
	}
	return fmt.Sprintf(a.tr(event.MessageKey), args...)
}
func (a *App) addLocalizedEvent(kind, key string, args ...any) {
	a.addGameEventRecord(GameEvent{Type: kind, MessageKey: key, MessageArgs: args, Timestamp: time.Now()})
}
