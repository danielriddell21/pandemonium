package sim

import "github.com/danielriddell21/pandemonium/internal/world"

// newSecrets builds the set of yet-undiscovered secret cells from a level.
func newSecrets(l *world.Level) map[world.Coord]bool {
	if len(l.Secrets) == 0 {
		return nil
	}
	m := make(map[world.Coord]bool, len(l.Secrets))
	for _, c := range l.Secrets {
		m[c] = true
	}
	return m
}

// checkSecret announces and records a secret the first time the player steps
// onto its cell.
func (g *Game) checkSecret(cell world.Coord) {
	if !g.secrets[cell] {
		return
	}
	delete(g.secrets, cell)
	g.found++
	g.setNotice("You found a secret area!")
	g.emit(Observation{Kind: ObsSecret, At: cell})
}
