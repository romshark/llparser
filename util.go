package parser

func isLineBreak(source []rune, index uint) int {
	switch source[index] {
	case '\n':
		return 1
	case '\r':
		next := index + 1
		if next < uint(len(source)) && source[next] == '\n' {
			return 2
		}
	}
	return -1
}
