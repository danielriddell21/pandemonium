package render

import "testing"

func TestDefaultTexturesBuilt(t *testing.T) {
	ts := defaultTextures()
	if ts.wall == nil || ts.door == nil || ts.fireball == nil || len(ts.demon) != 2 {
		t.Fatalf("incomplete texture set: %+v", ts)
	}
	if len(ts.demon[0].walk) < 2 || len(ts.demon[0].dead) < 2 {
		t.Errorf("demon variant missing animation frames: %+v", ts.demon[0])
	}
	if ts.wall.w != texSize || ts.wall.h != texSize {
		t.Errorf("wall texture is %dx%d, want %dx%d", ts.wall.w, ts.wall.h, texSize, texSize)
	}
}

func TestWallTextureHasDetail(t *testing.T) {
	w := defaultTextures().wall
	// A flat texture would have one colour; the brick pattern must vary.
	first := w.at(0, 0)
	varied := false
	for v := range texSize {
		for u := range texSize {
			if w.at(u, v) != first {
				varied = true
			}
		}
	}
	if !varied {
		t.Error("wall texture has no variation")
	}
}

func TestDemonSpriteHasTransparentBackground(t *testing.T) {
	s := defaultTextures().demon[0].walk[0]
	if a := s.at(0, 0).A; a != 0 {
		t.Errorf("sprite corner should be transparent, got alpha %d", a)
	}
	if a := s.at(texSize/2, texSize/2).A; a == 0 {
		t.Error("sprite centre should be opaque")
	}
}

func TestAtWraps(t *testing.T) {
	w := defaultTextures().wall
	if w.at(texSize, texSize) != w.at(0, 0) {
		t.Error("at should wrap out-of-range coordinates")
	}
}
