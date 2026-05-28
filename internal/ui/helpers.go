package ui

func repeatStr(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

func reptStr(s string, count int) string {
	return repeatStr(s, count)
}
