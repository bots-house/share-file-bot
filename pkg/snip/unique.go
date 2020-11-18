package snip

func UniqueizeStrings(ss []string) []string {
	set := make(map[string]struct{}, len(ss))

	for _, s := range ss {
		set[s] = struct{}{}
	}

	result := make([]string, 0, len(set))
	for k := range set {
		result = append(result, k)
	}

	return result
}
