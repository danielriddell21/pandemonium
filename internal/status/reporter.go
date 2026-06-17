package status

import (
	"github.com/danielriddell21/pandemonium/internal/hud"
	"github.com/danielriddell21/pandemonium/internal/telemetry"
)

// rushThreshold is the explore-score below which a run reads as beelining rather
// than exploring.
const rushThreshold = 0.5

// Reporter is a telemetry subscriber that posts HUD messages chosen by a Source.
type Reporter struct {
	overlay *hud.Overlay
	src     Source
	profile telemetry.RunProfile // latest cumulative run profile

	level   int  // level index of the events seen so far
	sawKill bool // a kill has been remarked on this level
	sawItem bool // an item pickup has been remarked on this level
}

// Compile-time check that Reporter consumes telemetry.
var _ telemetry.Subscriber = (*Reporter)(nil)

// New builds a Reporter that posts to overlay using src.
func New(overlay *hud.Overlay, src Source) *Reporter {
	return &Reporter{overlay: overlay, src: src}
}

// OnEvent maps a player event to a cue, folds in the run so far, and asks the
// source for a line, which it delivers to the overlay via emit (now or later).
func (r *Reporter) OnEvent(e telemetry.PlayerEvent) {
	cue, ok := r.cueFor(e)
	if !ok {
		return
	}
	r.enrich(&cue)
	r.src.Request(cue, r.emit)
}

// enrich folds the cumulative run profile into a cue.
func (r *Reporter) enrich(c *Cue) {
	p := r.profile
	c.LevelsCleared = p.LevelsCleared
	c.Deaths = p.Deaths
	c.WrongDoorsTotal = p.TotalWrongDoors
	c.ExploreScore = p.ExploreScore
	c.Rushing = p.ExploreScore > 0 && p.ExploreScore < rushThreshold
}

// emit posts a line to the overlay. It is safe to call from any goroutine.
func (r *Reporter) emit(line Line) {
	r.overlay.Post(line.Text, line.Frames, line.Channel)
}

// OnPathSummary is unused for now.
func (r *Reporter) OnPathSummary(telemetry.PathSummary) {}

// OnRunProfile records the latest cumulative profile so later cues can react to it.
func (r *Reporter) OnRunProfile(p telemetry.RunProfile) { r.profile = p }

// cueFor derives a cue from a player event, or reports false to ignore it. Kills
// and item pickups happen constantly, so only the first of each per level is
// remarked on; deaths and secret finds are rare enough to always surface.
func (r *Reporter) cueFor(e telemetry.PlayerEvent) (Cue, bool) {
	if e.LevelIndex != r.level {
		r.level, r.sawKill, r.sawItem = e.LevelIndex, false, false
	}
	switch e.Type {
	case "exit":
		return Cue{Kind: CueExit, Level: e.LevelIndex}, true
	case "death":
		return Cue{Kind: CueDeath, Level: e.LevelIndex}, true
	case "secret":
		return Cue{Kind: CueSecret, Level: e.LevelIndex}, true
	case "kill":
		if r.sawKill {
			break
		}
		r.sawKill = true
		return Cue{Kind: CueKill, Level: e.LevelIndex}, true
	case "item":
		if r.sawItem {
			break
		}
		r.sawItem = true
		return Cue{Kind: CueItem, Level: e.LevelIndex}, true
	case "door":
		if e.Marker != nil && e.Marker.WrongDoor {
			return Cue{Kind: CueWrongDoor, Level: e.LevelIndex}, true
		}
	case "marker":
		if e.Marker == nil {
			break
		}
		switch e.Marker.Kind {
		case "decoy_exit":
			return Cue{Kind: CueDecoy, Level: e.LevelIndex}, true
		case "junction":
			optimal := e.Marker.TakenX == e.Marker.OptimalX && e.Marker.TakenY == e.Marker.OptimalY
			return Cue{Kind: CueFork, Level: e.LevelIndex, OptimalChoice: optimal}, true
		}
	}
	return Cue{}, false
}
