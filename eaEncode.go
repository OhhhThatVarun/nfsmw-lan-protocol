package main

import "strings"

func eaEncode(data []byte) string {
	const t = "!\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}"
	var sb strings.Builder
	for i := 0; i < len(data); i += 3 {
		b0, b1, b2 := int(data[i]), 0, 0
		if i+1 < len(data) {
			b1 = int(data[i+1])
		}
		if i+2 < len(data) {
			b2 = int(data[i+2])
		}
		n := (b0 << 16) | (b1 << 8) | b2
		sb.WriteByte(t[(n>>18)&0x3f])
		sb.WriteByte(t[(n>>12)&0x3f])
		sb.WriteByte(t[(n>>6)&0x3f])
		sb.WriteByte(t[n&0x3f])
	}
	return sb.String()
}
