package world

import "testing"

// build constructs a Level from an ASCII map for testing. '#' is wall, '.' is
// floor, 'S' spawn, 'E' exit, '+' door. All rows must be equal length.
func build(t *testing.T, rows []string) *Level {
	t.Helper()
	h := len(rows)
	if h == 0 {
		t.Fatal("empty map")
	}
	w := len(rows[0])
	l := newLevel(w, h, 0)
	for y, row := range rows {
		if len(row) != w {
			t.Fatalf("row %d has width %d, want %d", y, len(row), w)
		}
		for x, r := range row {
			switch r {
			case '#':
				l.set(x, y, TileWall)
			case '.':
				l.set(x, y, TileFloor)
			case 'S':
				l.set(x, y, TileSpawn)
				l.Spawn = Coord{x, y}
			case 'E':
				l.set(x, y, TileExit)
				l.Exit = Coord{x, y}
			case '+':
				l.set(x, y, TileDoor)
			default:
				t.Fatalf("unknown rune %q", r)
			}
		}
	}
	return l
}

func TestReachable(t *testing.T) {
	tests := []struct {
		name string
		rows []string
		want bool
	}{
		{
			name: "open path",
			rows: []string{
				"#####",
				"#S.E#",
				"#####",
			},
			want: true,
		},
		{
			name: "wall blocks path",
			rows: []string{
				"#####",
				"#S#E#",
				"#####",
			},
			want: false,
		},
		{
			name: "winding path",
			rows: []string{
				"#######",
				"#S#...#",
				"#.#.#.#",
				"#...#E#",
				"#######",
			},
			want: true,
		},
		{
			name: "closed door blocks (door is solid)",
			rows: []string{
				"#####",
				"#S+E#",
				"#####",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := build(t, tt.rows)
			if got := Reachable(l, l.Spawn, l.Exit); got != tt.want {
				t.Errorf("Reachable = %v, want %v\n%s", got, tt.want, l)
			}
		})
	}
}
