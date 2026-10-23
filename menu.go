package main

import (
	"github.com/egoist/mygo"
	"github.com/ntnyq/cyber-game/locales"
)

type command struct {
	label   func() string
	enabled bool
	run     func()
}

func (w *window) commands() []command {
	return []command{
		{locales.StartGame, w.loadErr == nil, func() { w.navigate(setupPage) }},
		{locales.ContinueGame, w.session != nil && w.loadErr == nil, func() { w.navigate(pausePage) }},
		{locales.Records, true, func() { w.navigate(recordsPage) }},
		{locales.Settings, true, w.openSettings},
		{locales.QuitGame, true, w.requestQuit},
	}
}
func (w *window) syncMenu() {
	if w.win == nil {
		return
	}
	mygo.App.SetMenu(mygo.NewMenu([]*mygo.MenuItem{
		{Label: appName, Submenu: []*mygo.MenuItem{{Label: locales.Menu(), Click: func(*mygo.MenuItem, *mygo.Window) { w.dispatch(func() { w.pause(); w.navigate(menuPage) }) }}, {Label: locales.QuitGame(), Accelerator: "CmdOrCtrl+Q", Click: func(*mygo.MenuItem, *mygo.Window) { w.dispatch(w.requestQuit) }}}},
	}))
}
