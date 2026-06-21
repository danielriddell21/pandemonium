package sim

import "testing"

func TestFaceTurnsTowardDamage(t *testing.T) {
	g := newTestGame(t, 7)
	g.Player.Pos = Vec2{X: 10, Y: 10}
	g.Player.Angle = 0 // facing +X; the strafe-right axis is +Y

	g.faceHurt(Vec2{X: 12, Y: 10}) // dead ahead
	if g.Player.FaceDir() != 0 {
		t.Errorf("hit from ahead should face forward, got %d", g.Player.FaceDir())
	}
	g.faceHurt(Vec2{X: 10, Y: 12}) // to the player's right
	if g.Player.FaceDir() != 1 {
		t.Errorf("hit from the right should face right, got %d", g.Player.FaceDir())
	}
	g.faceHurt(Vec2{X: 10, Y: 8}) // to the player's left
	if g.Player.FaceDir() != -1 {
		t.Errorf("hit from the left should face left, got %d", g.Player.FaceDir())
	}
}

func TestFaceReturnsForwardAfterLapse(t *testing.T) {
	g := newTestGame(t, 7)
	g.Player.Pos = Vec2{X: 10, Y: 10}
	g.Player.Angle = 0
	g.faceHurt(Vec2{X: 10, Y: 12})
	if g.Player.FaceDir() == 0 {
		t.Fatal("face should be turned right after the hit")
	}
	g.Player.hurtTTL = 0 // lapse
	if g.Player.FaceDir() != 0 {
		t.Error("face should return to forward once the reaction lapses")
	}
}
