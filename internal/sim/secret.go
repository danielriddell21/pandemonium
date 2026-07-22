package sim

import "github.com/danielriddell21/pandemonium/internal/world"

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

func (g *Game) checkSecret(cell world.Coord) {
	if !g.secrets[cell] {
		return
	}
	delete(g.secrets, cell)
	g.found++
	g.setNotice("You found a secret area!")
	g.emit(Observation{Kind: ObsSecret, At: cell})
}
