package telemetry

// RunProfile is the cumulative, cross-level picture of a single play session: how
// far the player has progressed, how they tend to behave (thorough explorer vs.
// rusher), and the per-level summaries that make up the run.
type RunProfile struct {
	SessionID        string        `json:"session_id"`
	StartedMS        int64         `json:"started_ms"`
	LevelsCleared    int           `json:"levels_cleared"`
	Deaths           int           `json:"deaths"`
	TotalSteps       int           `json:"total_steps"`
	TotalDoorsOpened int           `json:"total_doors_opened"`
	TotalWrongDoors  int           `json:"total_wrong_doors"`
	ChoicesPerLevel  []int         `json:"choices_per_level"`
	ExploreScore     float64       `json:"explore_score"`
	Paths            []PathSummary `json:"paths"`
}
