package sim

import (
	"testing"

	"github.com/danielriddell21/pandemonium/internal/world"
)

func TestEnteringSecretFiresOnce(t *testing.T) {
	for seed := int64(0); seed < 100; seed++ {
		l, err := world.Generate(world.Config{Width: 48, Height: 32, Seed: seed})
		if err != nil {
			t.Fatal(err)
		}
		if len(l.Secrets) == 0 {
			continue
		}
		secret := l.Secrets[0]
		capt := &captureObserver{}
		g := New(l, WithObserver(capt))

		// Step onto the secret from the neighbouring cell.
		g.tracker.lastCell = world.Coord{X: secret.X, Y: secret.Y + 1}
		g.tracker.started = true
		g.Player.Pos = Vec2{X: float64(secret.X) + 0.5, Y: float64(secret.Y) + 0.5}
		g.observeMovement()

		if capt.count(ObsSecret) != 1 {
			t.Fatalf("seed %d: ObsSecret count = %d, want 1", seed, capt.count(ObsSecret))
		}
		if g.Notice() == "" {
			t.Errorf("seed %d: expected a secret-found notice", seed)
		}
		// Re-entering the same cell must not fire it again.
		g.tracker.lastCell = world.Coord{X: secret.X, Y: secret.Y + 1}
		g.observeMovement()
		if capt.count(ObsSecret) != 1 {
			t.Errorf("seed %d: secret re-fired: count = %d, want 1", seed, capt.count(ObsSecret))
		}
		return
	}
	t.Skip("no secrets generated in scanned seeds")
}
