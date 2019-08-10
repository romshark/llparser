package misc

func isLineBreak(source string, index uint) int {
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

func isSpecialChar(bt byte) bool {
	if bt >= 0x21 && bt <= 0x2F {
		// ! " # $ % & ' ( ) * + , - . /
		return true
	}
	if bt >= 0x3A && bt <= 0x40 {
		// : ; < = > ? @
		return true
	}
	if bt >= 0x5B && bt <= 0x60 {
		// [ \ ] ^ _ `
		return true
	}
	if bt >= 0x7B && bt <= 0x7E {
		// { | } ~
		return true
	}
	return false
}

func isDigit(bt byte) bool {
	if bt >= 0x30 && bt <= 0x39 {
		// 0-9
		return true
	}
	return false
}

func isSpace(bt byte) bool {
	if bt == ' ' || bt == '\t' {
		// whitespace or tab
		return true
	}
	return false
}

func isLatinLower(bt byte) bool {
	if bt >= 0x61 && bt <= 0x7A {
		// a-z
		return true
	}
	return false
}

func isLatinUpper(bt byte) bool {
	if bt >= 0x41 && bt <= 0x5A {
		// A-Z
		return true
	}
	return false
}

func isLatinAlphanum(bt byte) bool {
	if isDigit(bt) {
		return true
	}
	if isLatinLower(bt) {
		return true
	}
	if isLatinUpper(bt) {
		return true
	}
	return false
}
