package sgol

func StringSliceContains(s []string, i string) bool {
	for _, x := range s {
		if x == i {
			return true
		}
	}
	return false
}
