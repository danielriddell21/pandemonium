package gui

type Config struct {
	Seed          int64
	Width, Height int
}

func Available() bool { return true }
