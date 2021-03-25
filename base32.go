package geofire

import "strings"

const BITS_PER_BASE32_CHAR = 5
const BASE32_CHARS = "0123456789bcdefghjkmnpqrstuvwxyz"

func ToBase32Char(value int) byte {
	if value > len(BASE32_CHARS) {
		panic("Not a valid BASE32 value")
	}
	return BASE32_CHARS[value]
}

func ToBase32Value(b byte) int {
	return strings.IndexByte(BASE32_CHARS, b)
}
