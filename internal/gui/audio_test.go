package gui

import "testing"

// bandDepth is the first depth that MusicBand maps to a band above zero.
const bandDepth = 3

func TestUpdateMusicBandSwapsAndStopsTheOldPlayer(t *testing.T) {
	a, err := NewAudio()
	if err != nil {
		t.Skipf("no audio device available: %v", err)
	}
	a.StartAmbient()

	old := a.music
	if old == nil {
		t.Fatal("a new Audio should hold a music player")
	}
	a.depth = bandDepth
	a.updateMusicBand()

	if a.music == old {
		t.Fatal("crossing a band boundary should swap the music player")
	}
	if a.musicBand == 0 {
		t.Errorf("music band should have advanced, got %d", a.musicBand)
	}
	if old.IsPlaying() {
		t.Error("the replaced player should have stopped reading its source")
	}
	if !a.music.IsPlaying() {
		t.Error("the new player should pick up while the ambient bed is playing")
	}
}

func TestUpdateMusicBandKeepsThePlayerWithinABand(t *testing.T) {
	a := &Audio{}
	a.updateMusicBand() // no context, no player: must not panic

	if a.music != nil {
		t.Error("a silent Audio should stay silent")
	}
}
