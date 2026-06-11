package render

import "os"

// diagnosticsEnv is the environment variable that enables on-screen diagnostic
// (playtest) messages. When unset or falsey, diagnostics are hidden.
const diagnosticsEnv = "PANDEMONIUM_DEBUG"

// diagnosticsFromEnv reports whether diagnostics are enabled via the environment.
func diagnosticsFromEnv() bool {
	switch os.Getenv(diagnosticsEnv) {
	case "", "0", "false", "no":
		return false
	default:
		return true
	}
}
