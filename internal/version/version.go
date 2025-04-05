package version

// Build information. Populated at build-time.
var (
	// Version is the current version of the application
	Version = "dev"
	// Commit is the git commit SHA at build time
	Commit = "unknown"
	// BuildDate is the date when the binary was built
	BuildDate = "unknown"
	// GoVersion is the version of Go used to build the binary
	GoVersion = "unknown"
)

// Info returns version, commit, build date, and Go version information
func Info() string {
	return Version + " (commit: " + Commit + ", built: " + BuildDate + ", " + GoVersion + ")"
}
