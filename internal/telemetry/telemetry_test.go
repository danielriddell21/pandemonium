package telemetry

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

// capture is a test subscriber that records everything it receives.
type capture struct {
	events   []PlayerEvent
	paths    []PathSummary
	profiles []RunProfile
}

func (c *capture) OnEvent(e PlayerEvent)       { c.events = append(c.events, e) }
func (c *capture) OnPathSummary(p PathSummary) { c.paths = append(c.paths, p) }
func (c *capture) OnRunProfile(p RunProfile)   { c.profiles = append(c.profiles, p) }

// fixedClock returns a clock that advances one second per call, for
// deterministic timestamps.
func fixedClock() func() time.Time {
	base := time.Unix(1_000_000, 0)
	n := 0
	return func() time.Time {
		t := base.Add(time.Duration(n) * time.Second)
		n++
		return t
	}
}

func TestBusDispatchesAndAggregates(t *testing.T) {
	c := &capture{}
	b := NewBus(c)
	b.now = fixedClock()

	b.BeginLevel(42, 0)
	b.Observe(sim.Observation{Tick: 1, Kind: sim.ObsMove, At: world.Coord{X: 1, Y: 1}})
	b.Observe(sim.Observation{Tick: 2, Kind: sim.ObsMove, At: world.Coord{X: 2, Y: 1}})
	b.Observe(sim.Observation{Tick: 3, Kind: sim.ObsMove, At: world.Coord{X: 1, Y: 1}}) // backtrack
	b.Observe(sim.Observation{
		Tick: 4, Kind: sim.ObsMarker, At: world.Coord{X: 2, Y: 1},
		Marker:  world.MarkerJunction,
		Taken:   world.Coord{X: 3, Y: 1},
		Optimal: world.Coord{X: 3, Y: 1}, // took the optimal branch
		Ignored: []world.Coord{{X: 2, Y: 2}},
	})
	b.Observe(sim.Observation{Tick: 5, Kind: sim.ObsDoor, At: world.Coord{X: 4, Y: 1}, WrongDoor: true})
	b.Observe(sim.Observation{Tick: 6, Kind: sim.ObsExit, At: world.Coord{X: 5, Y: 1}})

	if len(c.events) != 6 {
		t.Errorf("events = %d, want 6", len(c.events))
	}
	if len(c.paths) != 1 {
		t.Fatalf("path summaries = %d, want 1", len(c.paths))
	}
	p := c.paths[0]
	if p.TilesVisited != 2 {
		t.Errorf("TilesVisited = %d, want 2", p.TilesVisited)
	}
	if p.BacktrackCount != 1 {
		t.Errorf("BacktrackCount = %d, want 1", p.BacktrackCount)
	}
	if p.JunctionsSeen != 1 || p.OptimalChoices != 1 {
		t.Errorf("junctions=%d optimal=%d, want 1/1", p.JunctionsSeen, p.OptimalChoices)
	}
	if p.DoorsOpened != 1 || p.WrongDoors != 1 {
		t.Errorf("doors=%d wrong=%d, want 1/1", p.DoorsOpened, p.WrongDoors)
	}
	if !p.Completed {
		t.Error("path not marked completed after exit")
	}

	rp := b.Profile()
	if rp.LevelsCleared != 1 {
		t.Errorf("LevelsCleared = %d, want 1", rp.LevelsCleared)
	}
	if len(rp.ChoicesPerLevel) != 1 || rp.ChoicesPerLevel[0] != 1 {
		t.Errorf("ChoicesPerLevel = %v, want [1]", rp.ChoicesPerLevel)
	}
	if rp.ExploreScore <= 0 || rp.ExploreScore > 1 {
		t.Errorf("ExploreScore = %v, want (0,1]", rp.ExploreScore)
	}
}

func TestBusCountsKillsItemsSecrets(t *testing.T) {
	c := &capture{}
	b := NewBus(c)
	b.now = fixedClock()

	b.BeginLevel(7, 0)
	b.Observe(sim.Observation{Kind: sim.ObsKill, At: world.Coord{X: 1, Y: 1}})
	b.Observe(sim.Observation{Kind: sim.ObsKill, At: world.Coord{X: 2, Y: 1}})
	b.Observe(sim.Observation{Kind: sim.ObsItem, At: world.Coord{X: 3, Y: 1}})
	b.Observe(sim.Observation{Kind: sim.ObsSecret, At: world.Coord{X: 4, Y: 1}})
	b.Observe(sim.Observation{Kind: sim.ObsExit, At: world.Coord{X: 5, Y: 1}})

	if len(c.paths) != 1 {
		t.Fatalf("path summaries = %d, want 1", len(c.paths))
	}
	p := c.paths[0]
	if p.Kills != 2 || p.ItemsTaken != 1 || p.SecretsFound != 1 {
		t.Errorf("kills=%d items=%d secrets=%d, want 2/1/1", p.Kills, p.ItemsTaken, p.SecretsFound)
	}

	rp := b.Profile()
	if rp.TotalKills != 2 || rp.TotalItems != 1 || rp.TotalSecrets != 1 {
		t.Errorf("totals kills=%d items=%d secrets=%d, want 2/1/1", rp.TotalKills, rp.TotalItems, rp.TotalSecrets)
	}
}

func TestBusBeginLevelFinalisesPrevious(t *testing.T) {
	c := &capture{}
	b := NewBus(c)
	b.now = fixedClock()

	b.BeginLevel(1, 0)
	b.Observe(sim.Observation{Kind: sim.ObsMove, At: world.Coord{X: 1, Y: 1}})
	b.BeginLevel(2, 1) // abandons level 0 without exit

	if len(c.paths) != 1 {
		t.Fatalf("path summaries = %d, want 1 (abandoned level finalised)", len(c.paths))
	}
	if c.paths[0].Completed {
		t.Error("abandoned level should not be marked completed")
	}
}

func TestNopSubscriberSatisfiesInterface(_ *testing.T) {
	var _ Subscriber = NopSubscriber{}
	b := NewBus(NopSubscriber{})
	b.BeginLevel(1, 0)
	b.Observe(sim.Observation{Kind: sim.ObsMove, At: world.Coord{X: 1, Y: 1}})
	b.Observe(sim.Observation{Kind: sim.ObsExit, At: world.Coord{X: 1, Y: 1}})
}

func TestEventJSONRoundTrip(t *testing.T) {
	in := PlayerEvent{
		Tick: 9, TimestampMS: 1234, Type: "marker", LevelSeed: 42, LevelIndex: 3,
		Cell: [2]int{5, 6},
		Marker: &MarkerInfo{
			Kind: "junction", X: 5, Y: 6, TakenX: 6, TakenY: 6,
			OptimalX: 6, OptimalY: 6, Ignored: [][2]int{{5, 7}},
		},
	}
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out PlayerEvent
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	if out.Marker == nil || out.Marker.Kind != "junction" || out.Cell != in.Cell {
		t.Errorf("round trip mismatch: %+v", out)
	}
}

func TestProfileJSONRoundTrip(t *testing.T) {
	in := RunProfile{
		SessionID: "abc", LevelsCleared: 2, TotalSteps: 100,
		ChoicesPerLevel: []int{1, 3}, ExploreScore: 0.42,
		Paths: []PathSummary{{LevelSeed: 1, Completed: true}},
	}
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out RunProfile
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	if out.SessionID != "abc" || len(out.Paths) != 1 || out.ExploreScore != 0.42 {
		t.Errorf("round trip mismatch: %+v", out)
	}
}
