package telemetry

// PathSummary is a rolling summary of how the player traversed a single level:
// how much of it they saw, how much they doubled back, how long they took, and
// how they handled doors and forks.
type PathSummary struct {
	LevelSeed      int64 `json:"level_seed"`
	LevelIndex     int   `json:"level_index"`
	Steps          int   `json:"steps"`
	TilesVisited   int   `json:"tiles_visited"`
	BacktrackCount int   `json:"backtrack_count"`
	TimeSpentMS    int64 `json:"time_spent_ms"`
	DoorsOpened    int   `json:"doors_opened"`
	WrongDoors     int   `json:"wrong_doors"`
	JunctionsSeen  int   `json:"junctions_seen"`
	OptimalChoices int   `json:"optimal_choices"`
	Kills          int   `json:"kills"`
	ItemsTaken     int   `json:"items_taken"`
	SecretsFound   int   `json:"secrets_found"`
	Completed      bool  `json:"completed"`
}
