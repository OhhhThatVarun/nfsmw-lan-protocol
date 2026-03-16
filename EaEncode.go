package main

import "strings"

// Encodes a byte slice using EA's custom 64-character text encoding.
//
// The function converts the input into 3-byte (24-bit) blocks. Each block is
// then split into 4 values of 6 bits each, and each 6-bit value is mapped to
// a character from EA's custom 64-character alphabet.
//
// This is structurally similar to standard Base64, but differs in several ways:
//
//   - It uses EA's own alphabet instead of the standard Base64 alphabet.
//   - It does not use '=' padding characters.
//   - If the final block contains fewer than 3 bytes, the missing bytes are
//     treated as zero during packing, but the function still emits 4 output
//     characters for that block.
//
// Output length is always:
//
// ceil(len(data) / 3) * 4
//
func EaEncode(data []byte) string {
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
