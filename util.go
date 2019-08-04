package main

// isDigit returns true if bt is a digit character, otherwise returns false
func isDigit(bt byte) bool {
	if bt >= 0x30 && bt <= 0x39 {
		// 0-9
		return true
	}
	return false
}

// isLatinUpperCase returns true if bt is an upper case latin letter,
// otherwise returns false
func isLatinUpperCase(bt byte) bool {
	if bt >= 0x41 && bt <= 0x5A {
		// A-Z
		return true
	}
	return false
}

// isLatinLowerCase returns true if bt is a lower case latin letter,
// otherwise returns false
func isLatinLowerCase(bt byte) bool {
	if bt >= 0x61 && bt <= 0x7A {
		// a-z
		return true
	}
	return false
}

// isLatinAlphanum return true if bt is either a digit, a lower or upper case
// latin letter, otherwise returns false
func isLatinAlphanum(bt byte) bool {
	if isDigit(bt) || isLatinLowerCase(bt) || isLatinUpperCase(bt) {
		return true
	}
	return false
}

// isSpace returns true if bt is either a whitespace or a tab,
// otherwise returns false
func isSpace(bt byte) bool {
	switch bt {
	case ' ':
		return true
	case '\t':
		return true
	}
	return false
}

// isLineBreak returns the end-index of either a line-feed
// or a carriage-return followed by a line-feed, otherwise returns -1
func isLineBreak(src string, index uint) int {
	switch src[index] {
	case '\n':
		return int(index) + 1
	case '\r':
		ix := index + 1
		if ix < uint(len(src)) && src[ix] == '\n' {
			return int(ix) + 1
		}
	}
	return -1
}

func isEOF(src string, index uint) bool {
	if index >= uint(len(src)) {
		return true
	}
	return false
}
