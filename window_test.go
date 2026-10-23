package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/cyber-game/internal/game"
	"github.com/ntnyq/cyber-game/internal/records"
	"github.com/ntnyq/cyber-game/locales"
)

// Run real background commits, but deliver their callbacks on the test's UI
// thread just as Window.Update does. No sleeps or concurrent state polling.
func persistentWindow(t *testing.T, path string) (*window, func()) {
	t.Helper()
	store, data, err := records.Open(path)
	w := newWindow(store, data, err)
	queue := make(chan func(), 32)
	w.postUpdate = func(fn func()) { queue <- fn }
	t.Cleanup(w.close)
	return w, func() {
		t.Helper()
		timer := time.NewTimer(5 * time.Second)
		defer timer.Stop()
		for w.task.Running() || (w.savedRevision != w.data.Revision && w.saveErr == nil) {
			select {
			case fn := <-queue:
				fn()
			case <-timer.C:
				t.Fatal("save completion was not dispatched")
			}
		}
	}
}

func clickView(t *testing.T, tt *ui.Tester, label string) {
	t.Helper()
	tt.Frame()
	if label == locales.Confirm() {
		// The dialog heading has the same text as its action. Select the
		// focusable button rather than the first matching text rectangle.
		for range 4 {
			if tt.Focused(label) {
				tt.Key(0, ui.KeyEnter)
				tt.Frame()
				return
			}
			tt.Key(0, ui.KeyTab)
		}
		t.Fatal("confirmation button was not keyboard reachable")
	}
	if err := tt.Click(label); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
}

func TestCorruptStoreCanQuitWithoutSaving(t *testing.T) {
	useLocale(t, "en")
	path := filepath.Join(t.TempDir(), "data.json")
	const original = "broken data to preserve"
	if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	w, waitSave := persistentWindow(t, path)
	w.openSettings()
	w.draft.Muted = true
	w.leaveSettings(true)
	waitSave()
	if w.loadErr == nil || w.saveErr == nil {
		t.Fatal("expected read protection and failed settings save")
	}
	w.requestQuit()
	if w.confirmAction == nil || w.allowClose {
		t.Fatal("unsaved settings should require confirmation")
	}
	tt := ui.NewTester(w.view, 1100, 700)
	clickView(t, tt, locales.Confirm())
	waitSave()
	tt.Frame()
	if w.quitting || w.allowClose || !tt.HasText(locales.DataDamaged()) || !tt.HasText(locales.DiscardQuit()) {
		t.Fatal("failed quit hid the protected data warning or escape route")
	}
	clickView(t, tt, locales.DiscardQuit())
	if w.confirmAction == nil || w.allowClose {
		t.Fatal("discard must require a second explicit confirmation")
	}
	clickView(t, tt, locales.Cancel())
	if w.allowClose || w.confirmAction != nil {
		t.Fatal("cancel unexpectedly allowed closing")
	}
	clickView(t, tt, locales.DiscardQuit())
	clickView(t, tt, locales.Confirm())
	if !w.allowClose {
		t.Fatal("confirmed discard did not allow closing")
	}
	b, err := os.ReadFile(path)
	if err != nil || string(b) != original {
		t.Fatalf("discard changed protected bytes: %q, %v", b, err)
	}
	backups, err := filepath.Glob(path + ".backup-*")
	if err != nil || len(backups) != 0 {
		t.Fatalf("discard unexpectedly repaired the store: %v, %v", backups, err)
	}
}

func TestTutorialPersistsAcrossQuitAndContinue(t *testing.T) {
	useLocale(t, "en")
	path := filepath.Join(t.TempDir(), "data.json")
	w, waitSave := persistentWindow(t, path)
	w.tutorial()
	waitSave()
	_, data, err := records.Open(path)
	if err != nil || data.Active == nil || data.Active.Mode != game.Tutorial || len(data.Active.Steps) != 0 {
		t.Fatalf("new tutorial was not saved: %+v, %v", data.Active, err)
	}
	for _, cell := range w.session.Puzzle.Solution[:3] {
		w.selectCell(cell)
		waitSave()
	}
	w.pause()
	waitSave()
	steps, elapsed := len(w.session.Steps), w.session.ElapsedMS
	tt := ui.NewTester(w.view, 1440, 900)
	clickView(t, tt, locales.SaveMenu())
	waitSave()
	if w.page != menuPage {
		t.Fatal("saved tutorial did not return to menu")
	}
	w.requestQuit()
	clickView(t, tt, locales.Confirm())
	waitSave()
	if !w.allowClose || w.saveErr != nil {
		t.Fatal("tutorial quit did not wait for successful persistence")
	}
	w.close()

	restored, waitRestored := persistentWindow(t, path)
	if restored.session == nil || restored.session.Mode != game.Tutorial || !restored.session.Unlimited || len(restored.session.Steps) != steps || restored.session.ElapsedMS != elapsed {
		t.Fatalf("tutorial state lost on restart: %+v", restored.session)
	}
	tt = ui.NewTester(restored.view, 1440, 900)
	clickView(t, tt, locales.ContinueGame())
	if restored.page != pausePage || tt.HasText(locales.Targets()) {
		t.Fatal("continued tutorial did not enter hidden pause state")
	}
	restored.tick(time.Now().Add(24 * time.Hour))
	if restored.session.ElapsedMS != elapsed {
		t.Fatal("offline time charged to tutorial")
	}
	clickView(t, tt, locales.Resume())
	waitRestored()
	if restored.session.Resumes != 1 || restored.focusCell != restored.session.Puzzle.Solution[steps] {
		t.Fatal("tutorial did not resume at the next guided step")
	}
	// Cell 4 is legal after the third scripted step, but not the tutorial's next step.
	if !restored.session.Legal(4) {
		t.Fatal("invalid test fixture")
	}
	restored.selectCell(4)
	if len(restored.session.Steps) != steps {
		t.Fatal("restored tutorial accepted a non-scripted move")
	}
	for _, cell := range restored.session.Puzzle.Solution[steps:] {
		restored.selectCell(cell)
		waitRestored()
	}
	if restored.page != resultPage || restored.result.Reason != game.Complete {
		t.Fatal("restored tutorial could not complete")
	}
	_, data, err = records.Open(path)
	if err != nil || data.Active != nil || len(data.History) != 0 {
		t.Fatalf("tutorial completion left a slot or score record: %+v, %v", data, err)
	}
}
