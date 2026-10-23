package main

import (
	"embed"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/cyber-game/internal/records"
	"github.com/ntnyq/cyber-game/internal/sound"
)

const appName = "Neural Breach"
const version = "0.1.0"

//go:embed resources/fonts/*.ttf
var fonts embed.FS

//go:embed resources/menu-background.png
var menuBackgroundBytes []byte

var menuBackground *ui.Bitmap

func registerFonts() error {
	if menuBackground == nil {
		var err error
		menuBackground, err = ui.DecodeBitmap(menuBackgroundBytes)
		if err != nil {
			return err
		}
	}
	entries, err := fonts.ReadDir("resources/fonts")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		b, err := fonts.ReadFile("resources/fonts/" + entry.Name())
		if err != nil {
			return err
		}
		if err = ui.RegisterFont(b, "Rajdhani"); err != nil {
			return err
		}
	}
	return nil
}
func main() {
	path := flag.String("data", "", "Path to the game data JSON file")
	showVersion := flag.Bool("version", false, "Print version")
	flag.Parse()
	if *showVersion {
		fmt.Println(appName, version)
		return
	}
	if *path == "" {
		dir, err := os.UserConfigDir()
		if err != nil {
			slog.Error("config directory", "error", err)
			os.Exit(1)
		}
		*path = filepath.Join(dir, "neural-breach", "data.json")
	}
	store, data, loadErr := records.Open(*path)
	if err := registerFonts(); err != nil {
		slog.Error("load font", "error", err)
	}
	mygo.App.SetName(appName)
	mygo.App.SetVersion(version)
	if !mygo.App.RequestSingleInstanceLock() {
		return
	}
	w := newWindow(store, data, loadErr)
	mygo.App.OnBeforeQuit(func(e *mygo.QuitEvent) {
		if !w.allowClose {
			e.PreventDefault()
			w.dispatch(w.requestQuit)
		}
	})
	mygo.App.OnSecondInstance(func([]string, string) {
		mygo.RunOnMain(func() {
			if w.win != nil {
				w.win.Show()
				w.win.Focus()
			}
		})
	})
	mygo.App.OnQuit(w.close)
	mygo.App.WhenReady(func() {
		mygo.Theme.SetSource(mygo.ThemeSource("dark"))
		mygo.App.SetLocale(data.Settings.Language)
		w.win = mygo.NewWindow(mygo.WindowOptions{Title: appName, Width: 1440, Height: 900, MinWidth: 1100, MinHeight: 700, Content: ui.View(w.view)})
		w.postUpdate = w.win.Update
		w.win.OnClose(func(e *mygo.CloseEvent) {
			if !w.allowClose {
				e.PreventDefault()
				w.dispatch(w.requestQuit)
			}
		})
		w.win.OnClosed(w.close)
		w.win.OnBlur(func() { w.dispatch(func() { w.background = true; w.pause() }) })
		w.win.OnFocus(func() { w.dispatch(func() { w.background = false }) })
		w.win.OnMinimize(func() { w.dispatch(func() { w.background = true; w.pause() }) })
		w.applyScale()
		w.syncMenu()
		w.audio = sound.New()
		w.updateAudio()
		go w.loop()
		if capture := os.Getenv("NEURAL_BREACH_CAPTURE"); capture != "" {
			go func() {
				time.Sleep(time.Second)
				b, err := w.win.CapturePage()
				if err == nil {
					err = os.WriteFile(capture, b, 0o600)
				}
				if err != nil {
					slog.Error("capture", "error", err)
				}
				mygo.RunOnMain(func() { w.allowClose = true; mygo.App.Quit() })
			}()
		}
	})
	if err := mygo.App.Run(); err != nil {
		slog.Error("application", "error", err)
		os.Exit(1)
	}
}
