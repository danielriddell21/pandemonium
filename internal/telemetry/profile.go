package telemetry

type RunProfile struct {
	SessionID        string        `json:"session_id"`
	StartedMS        int64         `json:"started_ms"`
	LevelsCleared    int           `json:"levels_cleared"`
	Deaths           int           `json:"deaths"`
	TotalSteps       int           `json:"total_steps"`
	TotalDoorsOpened int           `json:"total_doors_opened"`
	TotalWrongDoors  int           `json:"total_wrong_doors"`
	TotalKills       int           `json:"total_kills"`
	TotalItems       int           `json:"total_items"`
	TotalSecrets     int           `json:"total_secrets"`
	ChoicesPerLevel  []int         `json:"choices_per_level"`
	ExploreScore     float64       `json:"explore_score"`
	Paths            []PathSummary `json:"paths"`
}
