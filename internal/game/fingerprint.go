package game

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// Fingerprint identifies the playable content, independent of its seed or witness.
// JSON preserves the boundaries and order of the board and target sequences.
func (p Puzzle) Fingerprint() string {
	b, _ := json.Marshal(struct {
		Version    int        `json:"version"`
		Difficulty Difficulty `json:"difficulty"`
		Board      []string   `json:"board"`
		Targets    [][]string `json:"targets"`
	}{p.Version, p.Difficulty, p.Board, p.Targets})
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
