package main

import (
	"context"
	"errors"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/cyber-game/internal/game"
	"github.com/ntnyq/cyber-game/internal/records"
	"github.com/ntnyq/cyber-game/locales"
)

func useLocale(t *testing.T, tag string) {
	t.Helper()
	old := i18n.Locale()
	i18n.SetLocaleFunc(func() string { return tag })
	t.Cleanup(func() { i18n.SetLocaleFunc(func() string { return old }) })
}

func writeViewSnapshot(t *testing.T, tt *ui.Tester, name string) {
	t.Helper()
	dir := os.Getenv("NEURAL_BREACH_SNAPSHOTS")
	if dir == "" {
		return
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(dir, name+".png"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, tt.Image()); err != nil {
		t.Fatal(err)
	}
}
func fixtureWindow() *window {
	d := records.Empty()
	s := game.New(game.TutorialPuzzle(), game.Standard, false, "fixture", time.Now())
	for _, cell := range s.Puzzle.Solution {
		s.Select(cell, time.Now())
	}
	d.History = []game.Session{s}
	w := newWindow(nil, d, nil)
	w.result = &s
	return w
}
func TestTutorialThroughPauseAndResult(t *testing.T) {
	useLocale(t, "en")
	w := fixtureWindow()
	tt := ui.NewTester(w.view, 1440, 900)
	tt.Frame()
	if err := tt.Click(locales.StartGame()); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if w.page != setupPage {
		t.Fatal("start did not open setup")
	}
	if err := tt.Click(locales.Tutorial()); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if w.session == nil || w.session.Mode != game.Tutorial {
		t.Fatal("tutorial absent")
	}
	if err := tt.Click("R1 C2 1C"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if len(w.session.Steps) != 1 {
		t.Fatal("cell click failed")
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if w.page != pausePage || tt.HasText(locales.Targets()) || tt.HasText("R1 C2 1C") {
		t.Fatal("pause leaked board")
	}
	if err := tt.Click(locales.Settings()); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if err := tt.Click(locales.Cancel()); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if w.page != pausePage {
		t.Fatal("settings did not return to pause")
	}
	if err := tt.Click(locales.Resume()); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	for _, cell := range game.TutorialPuzzle().Solution[1:] {
		w.selectCell(cell)
	}
	tt.Frame()
	if w.page != resultPage || w.result.Reason != game.Complete {
		t.Fatal("tutorial not completed")
	}
	if len(w.data.History) != 1 {
		t.Fatal("tutorial polluted history")
	}
}
func TestTimerPauseAndLateInput(t *testing.T) {
	w := fixtureWindow()
	s := game.New(game.TutorialPuzzle(), game.Standard, false, "clock", time.Now())
	w.session = &s
	w.page = gamePage
	w.lastTick = time.Now().Add(-time.Hour)
	w.advance(time.Now())
	if s.ElapsedMS != 0 {
		t.Fatal("observation charged")
	}
	s.Select(1, time.Now())
	w.lastTick = time.Now().Add(-time.Second)
	w.pause()
	elapsed := s.ElapsedMS
	if elapsed < 1000 || elapsed > 1050 {
		t.Fatal(elapsed)
	}
	w.tick(time.Now().Add(time.Hour))
	if s.ElapsedMS != elapsed {
		t.Fatal("pause charged time")
	}
	w.resume()
	w.tick(w.lastTick.Add(60 * time.Second))
	if w.page != resultPage || w.result.Reason != game.Timeout {
		t.Fatal("timeout missing")
	}
}

func TestKeyboardNavigationAndMute(t *testing.T) {
	useLocale(t, "en")
	w := fixtureWindow()
	tt := ui.NewTester(w.view, 1440, 900)
	tt.Frame()
	tt.Key(0, ui.KeyDown)
	tt.Frame()
	if w.menuIndex != 2 {
		t.Fatalf("menu did not skip unavailable continue: %d", w.menuIndex)
	}
	tt.Key(0, ui.KeyM)
	tt.Frame()
	if !w.data.Settings.Muted {
		t.Fatal("mute shortcut failed")
	}
	w.tutorial()
	tt.Frame()
	tt.Key(0, ui.KeyRight)
	tt.Frame()
	if w.focusCell != 2 {
		t.Fatalf("board focus did not move: %d", w.focusCell)
	}
	tt.Key(0, ui.KeyLeft)
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if len(w.session.Steps) != 1 {
		t.Fatal("keyboard selection failed")
	}
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if len(w.session.Steps) != 1 {
		t.Fatal("repeated confirmation consumed a slot")
	}
}

func TestBoardAccessibility(t *testing.T) {
	useLocale(t, "en")
	w := fixtureWindow()
	w.tutorial()
	tt := ui.NewTester(w.view, 1100, 700)
	tt.Frame()
	tt.UseScreenReader()

	puzzle := w.session.Puzzle
	size := puzzle.Difficulty.Rules().Size
	if grid, ok := tt.AccessByRole(ui.RoleGrid, locales.Matrix()); !ok || grid.RowCount != size || grid.ColumnCount != size {
		t.Fatalf("matrix accessibility grid: %+v", grid)
	}
	cellName := func(index int) string {
		return fmt.Sprintf("R%d C%d %s", index/size+1, index%size+1, puzzle.Board[index])
	}
	first := puzzle.Solution[0]
	firstName := cellName(first)
	if cell, ok := tt.AccessByRole(ui.RoleCell, firstName); !ok || cell.Disabled || cell.Selected || cell.RowIndex != 1 || cell.ColumnIndex != first+1 || cell.Description != locales.SuggestedCell() {
		t.Fatalf("first tutorial cell: %+v", cell)
	}
	if cell, ok := tt.AccessByRole(ui.RoleCell, cellName((first+1)%size)); !ok || !cell.Disabled {
		t.Fatalf("other cell in the first row: %+v", cell)
	}
	if err := tt.Click(firstName); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if cell, ok := tt.AccessByRole(ui.RoleCell, firstName); !ok || !cell.Disabled || !cell.Selected {
		t.Fatalf("used cell: %+v", cell)
	}
	if got := tt.Announcements(); !slices.Contains(got, locales.ChooseColumn(first%size+1)) {
		t.Fatalf("missing column announcement: %q", got)
	}
	for _, index := range puzzle.Solution[1:3] {
		if err := tt.Click(cellName(index)); err != nil {
			t.Fatal(err)
		}
		tt.Frame()
	}
	progress := fmt.Sprintf("%s 1/%d", locales.Uploaded(), len(puzzle.Targets))
	nextRow := locales.ChooseRow(puzzle.Solution[1]/size + 1)
	if got := tt.Announcements(); !slices.Contains(got, progress) || !slices.Contains(got, nextRow) {
		t.Fatalf("expected announcements for %q and %q, got %q", nextRow, progress, got)
	}
	tt.Frame()
	if got := tt.Announcements(); len(got) > 0 {
		t.Fatalf("unchanged statuses were announced again: %q", got)
	}
	w.pause()
	tt.Frame()
	if _, ok := tt.Access(firstName); ok {
		t.Fatal("paused view exposed a matrix cell to assistive technology")
	}
	w.openReplay(*w.session)
	tt.Frame()
	if _, ok := tt.AccessByRole(ui.RoleTable, locales.Matrix()); !ok {
		t.Fatal("replay matrix is missing its table semantics")
	}
	if cell, ok := tt.AccessByRole(ui.RoleCell, firstName); !ok || !cell.Disabled {
		t.Fatalf("replay cell: %+v", cell)
	}
}
func TestViewSnapshots(t *testing.T) {
	if err := registerFonts(); err != nil {
		t.Fatal(err)
	}
	for _, tag := range locales.Catalog.Languages() {
		t.Run(tag, func(t *testing.T) {
			useLocale(t, tag)
			for _, size := range [][3]int{{1100, 700, 100}, {1440, 900, 100}, {1650, 1050, 150}} {
				for _, page := range []page{menuPage, setupPage, gamePage, pausePage, resultPage, recordsPage, replayPage, settingsPage} {
					w := fixtureWindow()
					w.data.Settings.Scale = size[2]
					w.draft = w.data.Settings
					w.page = page
					w.focusMenu = false
					w.glitchStart = time.Time{}
					p, err := game.Generate(context.Background(), 2077, 3)
					if err != nil {
						t.Fatal(err)
					}
					s := game.New(p, game.Standard, false, "active", time.Now())
					w.session = &s
					tt := ui.NewTester(w.view, size[0], size[1])
					tt.SetLocale(tag)
					tt.Frame()
					if len(tt.Texts()) == 0 {
						t.Fatal("empty view")
					}
					if page == gamePage {
						label := fmt.Sprintf("R8 C8 %s", p.Board[63])
						r, ok := tt.Find(label)
						if !ok || r.H < 44 || r.X+r.W > float32(size[0]) || r.Y+r.H > float32(size[1])-60 {
							t.Fatalf("grid clipped %v %v %v", size, r, ok)
						}
					}
					if page == pausePage && tt.HasText(locales.Targets()) {
						t.Fatal("pause leaks targets")
					}
					if tag == "en" || tag == "zh-Hans" {
						writeViewSnapshot(t, tt, fmt.Sprintf("%s-%d-%d-%d", tag, page, size[0], size[2]))
					}
				}
			}
		})
	}
}

func TestViewSnapshotsProtectedData(t *testing.T) {
	if err := registerFonts(); err != nil {
		t.Fatal(err)
	}
	for _, tag := range locales.Catalog.Languages() {
		t.Run(tag, func(t *testing.T) {
			useLocale(t, tag)
			for _, size := range [][3]int{{1100, 700, 100}, {1650, 1050, 150}} {
				w := fixtureWindow()
				w.loadErr = errors.New("unsupported data version")
				w.saveErr = fmt.Errorf("data protected: %w", w.loadErr)
				w.data.Settings.Scale = size[2]
				w.glitchStart = time.Time{}
				tt := ui.NewTester(w.view, size[0], size[1])
				tt.Frame()
				if tt.HasText(locales.RetrySave()) {
					t.Fatal("protected store offered a retry that cannot succeed")
				}
				for _, label := range []string{locales.Repair(), locales.DiscardQuit()} {
					r, ok := tt.Find(label)
					if !ok || r.X < 0 || r.Y < 0 || r.X+r.W > float32(size[0]) || r.Y+r.H > float32(size[1]) {
						t.Fatalf("recovery action clipped: %s %v %v", label, size, r)
					}
				}
				if tag == "en" || tag == "zh-Hans" {
					writeViewSnapshot(t, tt, fmt.Sprintf("%s-protected-%d-%d", tag, size[0], size[2]))
				}
			}
		})
	}
}

func TestViewSnapshotsGameSaveFailure(t *testing.T) {
	if err := registerFonts(); err != nil {
		t.Fatal(err)
	}
	puzzle, err := game.Generate(context.Background(), 2077, 3)
	if err != nil {
		t.Fatal(err)
	}
	for _, tag := range locales.Catalog.Languages() {
		t.Run(tag, func(t *testing.T) {
			useLocale(t, tag)
			for _, scale := range []int{100, 125, 150} {
				for _, p := range []page{gamePage, replayPage} {
					width, height := 1100*scale/100, 700*scale/100
					w := fixtureWindow()
					w.data.Settings.Scale = scale
					w.page = p
					w.glitchStart = time.Time{}
					s := game.New(puzzle, game.Practice, true, "active", time.Now())
					w.session, w.result = &s, &s
					w.saveErr = errors.New("replace /temporary/neural-breach/data.json: permission denied")
					tt := ui.NewTester(w.view, width, height)
					tt.SetLocale(tag)
					tt.Frame()
					for _, label := range []string{locales.SaveFailed(), locales.RetrySave(), locales.DiscardQuit()} {
						recovery, ok := tt.Find(label)
						if !ok || recovery.X < 0 || recovery.Y < 0 || recovery.X+recovery.W > float32(width) || recovery.Y+recovery.H > float32(height) {
							t.Fatalf("recovery action clipped: page=%d scale=%d label=%s rect=%+v", p, scale, label, recovery)
						}
						for cell, code := range puzzle.Board {
							board, ok := tt.Find(fmt.Sprintf("R%d C%d %s", cell/8+1, cell%8+1, code))
							if !ok || board.W+0.01 < 44*float32(scale)/100 || board.H+0.01 < 44*float32(scale)/100 {
								t.Fatalf("board cell missing or too small: page=%d scale=%d cell=%d rect=%+v", p, scale, cell, board)
							}
							if recovery.X < board.X+board.W && recovery.X+recovery.W > board.X && recovery.Y < board.Y+board.H && recovery.Y+recovery.H > board.Y {
								t.Fatalf("recovery overlaps the board: page=%d scale=%d label=%s cell=%d", p, scale, label, cell)
							}
						}
						if p == gamePage {
							for _, action := range []string{locales.Hint(), locales.Undo(), locales.Finish()} {
								control, ok := tt.Find(action)
								if !ok || recovery.X < control.X+control.W && recovery.X+recovery.W > control.X && recovery.Y < control.Y+control.H && recovery.Y+recovery.H > control.Y {
									t.Fatalf("recovery overlaps or hides %s: scale=%d", action, scale)
								}
							}
						}
					}
					writeViewSnapshot(t, tt, fmt.Sprintf("%s-save-failed-%d-%d", tag, p, scale))
				}
			}
		})
	}
}
