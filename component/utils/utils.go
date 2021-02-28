package utils

func RemoveDuplicateStrings(s []string) (result []string, resultMap map[string]struct{}) {
	result = make([]string, 0)
	resultMap = make(map[string]struct{})
	for _, v := range s {
		if _, ok := resultMap[v]; ok { // Duplicated element found
			continue
		} else {
			resultMap[v] = struct{}{}
			result = append(result, v)
		}
	}
	return
}

func DeleteEmptyString(s []string) []string {
	r := make([]string, 0, len(s))
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}
