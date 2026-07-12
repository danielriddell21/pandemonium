package audio

import "testing"

func TestMusicLengthAndDeterminism(t *testing.T) {
	want := int(musicLoop*SampleRate) * bytesPerFrame
	a := Music(0)
	if len(a) != want {
		t.Fatalf("music length = %d, want %d", len(a), want)
	}
	b := Music(0)
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("music byte %d differs between builds", i)
		}
	}
}

// TestMusicLoopsSeamlessly checks the first and last frames are close, so the
// infinite loop has no audible click.
func TestMusicLoopsSeamlessly(t *testing.T) {
	for band := 0; band <= maxMusicBand; band++ {
		pcm := Music(band)
		first := int16(pcm[0]) | int16(pcm[1])<<8
		last := int16(pcm[len(pcm)-4]) | int16(pcm[len(pcm)-3])<<8
		if d := int(first) - int(last); d < -1200 || d > 1200 {
			t.Errorf("band %d loop seam too large: first=%d last=%d", band, first, last)
		}
	}
}

// TestMusicBandProgression checks the band rises with depth and then holds at
// the floor.
func TestMusicBandProgression(t *testing.T) {
	if MusicBand(0) != 0 {
		t.Error("a fresh run should start at band 0")
	}
	if MusicBand(3) <= MusicBand(0) {
		t.Error("the band should rise with depth")
	}
	if MusicBand(10_000) != maxMusicBand {
		t.Errorf("deep runs should hold at the floor band %d", maxMusicBand)
	}
}

// TestMusicDarkensWithBand checks deeper bands differ from shallow ones (the bed
// actually changes, not just plays the same loop).
func TestMusicDarkensWithBand(t *testing.T) {
	shallow, deep := Music(0), Music(maxMusicBand)
	same := true
	for i := range shallow {
		if shallow[i] != deep[i] {
			same = false
			break
		}
	}
	if same {
		t.Error("the deepest band should sound different from the first")
	}
}

// TestMusicBandClamps checks out-of-range bands do not panic and clamp.
func TestMusicBandClamps(t *testing.T) {
	if len(Music(-5)) != len(Music(0)) {
		t.Error("negative band should clamp to 0")
	}
	if len(Music(99)) != len(Music(maxMusicBand)) {
		t.Error("excessive band should clamp to the floor")
	}
}
