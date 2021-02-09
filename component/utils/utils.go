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
