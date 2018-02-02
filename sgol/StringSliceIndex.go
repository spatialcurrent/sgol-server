package sgol

func StringSliceIndex(s []string, value string) int {
	for i, x := range s {
		if x == value {
			return i
		}
	}
	return -1
}
