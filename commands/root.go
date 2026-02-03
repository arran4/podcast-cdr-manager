package commands

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// PodcastCdrManager is a subcommand `podcast-cdr-manager`
// Flags:
//   profile: --profile (default: "default") The profile/user
func PodcastCdrManager(profile string) {
}

func GetProfile(p string) string {
	if p == "default" && version == "dev" {
		return "default-dev"
	}
	return p
}
