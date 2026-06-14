package render

import "testing"

func TestViewmodelDrawsWeapon(t *testing.T) {
	cfg := Config{Width: 200, Height: 120, FOV: 1.152}
	tx := defaultTextures()
	fb := make([]byte, cfg.Width*cfg.Height*4)

	drawViewmodel(fb, cfg, tx.weapon[1], tx.flash, false, 0) // pistol, not firing
	if countColored(fb, cfg) == 0 {
		t.Error("weapon viewmodel drew nothing")
	}
}

func TestMuzzleFlashAddsPixels(t *testing.T) {
	cfg := Config{Width: 200, Height: 120, FOV: 1.152}
	tx := defaultTextures()

	noFire := make([]byte, cfg.Width*cfg.Height*4)
	firing := make([]byte, cfg.Width*cfg.Height*4)
	drawViewmodel(noFire, cfg, tx.weapon[1], tx.flash, false, 0)
	drawViewmodel(firing, cfg, tx.weapon[1], tx.flash, true, 0)

	if countColored(firing, cfg) <= countColored(noFire, cfg) {
		t.Error("muzzle flash should add lit pixels when firing")
	}
}
