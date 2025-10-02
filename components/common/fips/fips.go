package fips

import (
	"crypto/fips140"
	"os"
)

// IsFIPS140Only checks if the application is running in FIPS 140 exclusive mode.
func IsFIPS140Only() bool {
	return fips140.Enabled() && os.Getenv("GODEBUG") == "fips140=only,tlsmlkem=0"
}
