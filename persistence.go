package main

import (
	"context"
	"encoding/json"

	"github.com/egoist/mygo"
	"github.com/ntnyq/cyber-game/internal/game"
	"github.com/ntnyq/cyber-game/internal/records"
	"github.com/ntnyq/cyber-game/internal/storage"
	"github.com/ntnyq/cyber-game/locales"
)

func (w *window) persist(after func()) {
	if w.store == nil {
		w.completeSave(after)
		return
	}
	if w.session != nil {
		s := *w.session
		w.data.Active = &s
		if s.Mode != game.Tutorial {
			w.data.RememberPuzzle(s.Puzzle)
		}
	}
	w.data.Revision++
	snapshot := w.data.Clone()
	rev := snapshot.Revision
	w.saveErr = nil
	go func() {
		err := w.store.Commit(snapshot)
		w.dispatch(func() {
			if rev != w.data.Revision {
				return
			}
			w.saveErr = err
			if err == nil {
				w.savedRevision = rev
				w.completeSave(after)
			} else {
				w.quitting = false
			}
		})
	}()
}

func (w *window) completeSave(after func()) {
	// Quit intent belongs to the window, even when another save supersedes
	// the snapshot that first requested it.
	if w.quitting {
		w.quitWithoutSave()
	} else if after != nil {
		after()
	}
}

func (w *window) saveQuit() {
	w.task.Cancel()
	w.quitting = true
	w.persist(nil)
}

func (w *window) repair() {
	w.ask(locales.RepairConfirm, func() {
		var d records.Data
		w.run(func(context.Context) error { var err error; d, err = w.store.Repair(); return err }, func() {
			w.data = d
			w.loadErr = nil
			w.saveErr = nil
			w.session = nil
			w.savedRevision = d.Revision
			w.navigate(menuPage)
		})
	})
}

func (w *window) export(csv bool) {
	snapshot := w.data.Clone()
	w.run(func(ctx context.Context) error {
		ext := "json"
		if csv {
			ext = "csv"
		}
		path, err := mygo.Dialog.Save(mygo.SaveDialogOptions{Parent: w.win, Title: appName, DefaultPath: "neural-breach." + ext})
		if err != nil {
			return err
		}
		if path == "" {
			return context.Canceled
		}
		if err = ctx.Err(); err != nil {
			return err
		}
		var b []byte
		if csv {
			b, err = records.CSV(snapshot.History)
		} else {
			b, err = json.MarshalIndent(snapshot.History, "", "  ")
		}
		if err != nil {
			return err
		}
		return storage.WriteFile(path, b, 0o600)
	}, func() { w.message = locales.Exported })
}
