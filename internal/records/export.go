package records

import (
	"encoding/csv"
	"strconv"
	"strings"
	"time"

	"github.com/ntnyq/cyber-game/internal/game"
)

func CSV(history []game.Session) ([]byte, error) {
	var b strings.Builder
	w := csv.NewWriter(&b)
	_ = w.Write([]string{"id", "finished_at", "mode", "difficulty", "seed", "result", "score", "uploaded", "steps", "elapsed_ms", "resumes", "hints", "undos"})
	for _, s := range history {
		if err := w.Write([]string{s.ID, s.FinishedAt.Format(time.RFC3339), string(s.Mode), strconv.Itoa(int(s.Puzzle.Difficulty)), strconv.FormatUint(s.Puzzle.Seed, 10), string(s.Reason), strconv.Itoa(s.Score().Total), strconv.Itoa(s.Uploaded()), strconv.Itoa(len(s.Steps)), strconv.FormatInt(s.ElapsedMS, 10), strconv.Itoa(s.Resumes), strconv.Itoa(s.Hints), strconv.Itoa(s.Undos)}); err != nil {
			return nil, err
		}
	}
	w.Flush()
	return []byte(b.String()), w.Error()
}
