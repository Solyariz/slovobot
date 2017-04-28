package util

import "unicode/utf8"

func ToRunes(ss string) []rune {
	runes := make([]rune, 0)
	s := []byte(ss)
	for utf8.RuneCount(s) > 0 {
		//r, size := utf8.DecodeRune(s)
		//s = s[size:]
		nextR, size := utf8.DecodeRune(s)
		runes = append(runes, nextR)
		s = s[size:]
		//fmt.Print(r == nextR, ",")
	}
	return runes
}
