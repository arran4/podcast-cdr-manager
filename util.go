package podcast_cdr_manager

func SkipFirstN[T any](args []T, i int) []T {
	m := i
	if m >= len(args) {
		m = len(args)
	}
	return args[m:]
}
