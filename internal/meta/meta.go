package meta

var (
	// Version is the compile-time set version of Watchtower
	Version = "v1.8.1"

	// UserAgent is the http client identifier derived from Version
	UserAgent string
)

func init() {
	UserAgent = "Watchtower/" + Version
}
