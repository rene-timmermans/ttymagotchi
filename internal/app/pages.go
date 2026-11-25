package app

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const footerText = "Esc / Ctrl+S Status | Ctrl+F Feed | Ctrl+P Play | Ctrl+L Sleep\nCtrl+Y Hygiene | Ctrl+O Options | Ctrl+E Events\nCtrl+H Help | Ctrl+R Restart | Ctrl+C Quit"

const (
	statusSection  = "Status"
	feedSection    = "Feed"
	playSection    = "Play"
	sleepSection   = "Sleep"
	hygieneSection = "Hygiene"
	optionsSection = "Options"
	eventsSection  = "Events"
	helpSection    = "Help"
)

func (a *App) getPagesInfo() (tview.Primitive, tview.Primitive) {
	pages := tview.NewPages()

	// Add all pages
	_, statusContent := a.statusPage()
	pages.AddPage(statusSection, statusContent, true, true)

	_, feedContent := a.feedPage()
	pages.AddPage(feedSection, feedContent, true, false)

	_, playContent := a.playPage()
	pages.AddPage(playSection, playContent, true, false)

	_, sleepContent := a.sleepPage()
	pages.AddPage(sleepSection, sleepContent, true, false)

	_, eventsContent := a.eventsPage()
	pages.AddPage(eventsSection, eventsContent, true, false)

	_, helpContent := a.helpPage()
	pages.AddPage(helpSection, helpContent, true, false)

	_, hygieneContent := a.hygienePage()
	pages.AddPage(hygieneSection, hygieneContent, true, false)
	_, optionsContent := a.optionsPage()
	pages.AddPage(optionsSection, optionsContent, true, false)

	// Info bar
	info := tview.NewTextView().
		SetDynamicColors(true).
		SetText(a.tr(footerText))
	a.infoView = info

	return pages, info
}

func (a *App) goToSection(section string, info tview.Primitive) {
	pages := a.TLayout.GetItem(0).(*tview.Pages)
	pages.SwitchToPage(section)
	pages.SendToFront(section)
}

func getList() *tview.List {
	list := tview.NewList()
	list.ShowSecondaryText(false)
	list.SetMainTextColor(tcell.ColorWhite)
	list.SetSelectedTextColor(tcell.ColorBlack)
	list.SetSelectedBackgroundColor(tcell.ColorYellow)
	list.SetBorder(true)
	list.SetTitle("ttymagotchi")
	return list
}

// Escape and Ctrl+S have the same destination, including from an activity.
func (a *App) returnToStatus() {
	if g := a.activity; g != nil {
		if g.phase != "result" {
			if isCare(g.kind) {
				a.finishCare()
			} else if g.kind == "food" {
				a.finishFeeding()
			}
		}
		a.closeActivity()
	}
	// Returning from the restart dialog cancels the pending restart.
	a.modal = nil
	a.TApp.SetRoot(a.TLayout, true)
	a.goToSection(statusSection, nil)
	a.TApp.SetFocus(a.TLayout)
	a.requestRefresh()
}
