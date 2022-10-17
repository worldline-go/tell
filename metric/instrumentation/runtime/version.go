package runtime

const Version = "v0.0.0"

// SemVersion is the semantic version to be supplied to tracer/meter creation.
func SemVersion() string {
	return "semver:" + Version
}
