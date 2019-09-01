package misc

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

func isSpecialChar(bt rune) bool {
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

func isSpace(bt rune) bool {
	if bt == ' ' || bt == '\t' {
		// whitespace or tab
		return true
	}
	return false
}
