package telemetry

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
