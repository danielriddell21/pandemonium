package render

func countColored(fb []byte, cfg Config) int {
	n := 0
	for i := 0; i+3 < len(fb); i += 4 {
		if fb[i] != 0 || fb[i+1] != 0 || fb[i+2] != 0 {
			n++
		}
	}
	_ = cfg
	return n
}
