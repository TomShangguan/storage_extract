package trie

// Trie keys are dealt with in three distinct encodings:
//
// KEYBYTES encoding contains the actual key and nothing else. This encoding is the
// input to most API functions.
//
// HEX encoding contains one byte for each nibble of the key and an optional trailing
// 'terminator' byte of value 0x10 which indicates whether or not the node at the key
// contains a value. Hex key encoding is used for nodes loaded in memory because it's
// convenient to access.
//
// COMPACT encoding is defined by the Ethereum Yellow Paper (it's called "hex prefix
// encoding" there) and contains the bytes of the key and a flag. The high nibble of the
// first byte contains the flag; the lowest bit encoding the oddness of the length and
// the second-lowest encoding whether the node at the key is a value node. The low nibble
// of the first byte is zero in the case of an even number of nibbles and the first nibble
// in the case of an odd number. All remaining nibbles (now an even number) fit properly
// into the remaining bytes. Compact encoding is used for nodes stored on disk.

func hexToCompact(hex []byte) []byte {
	terminator := byte(0)
	if hasTerm(hex) {
		terminator = 1
		hex = hex[:len(hex)-1]
	}
	buf := make([]byte, len(hex)/2+1)
	buf[0] = terminator << 5 // the flag byte
	if len(hex)&1 == 1 {
		buf[0] |= 1 << 4 // odd flag
		buf[0] |= hex[0] // first nibble is contained in the first byte
		hex = hex[1:]
	}
	decodeNibbles(hex, buf[1:])
	return buf
}

func keybytesToHex(str []byte) []byte {
	l := len(str)*2 + 1
	var nibbles = make([]byte, l)
	for i, b := range str {
		nibbles[i*2] = b / 16
		nibbles[i*2+1] = b % 16
	}
	nibbles[l-1] = 16
	return nibbles
}

func decodeNibbles(nibbles []byte, bytes []byte) {
	for bi, ni := 0, 0; ni < len(nibbles); bi, ni = bi+1, ni+2 {
		bytes[bi] = nibbles[ni]<<4 | nibbles[ni+1]
	}
}

// prefixLen returns the length of the common prefix of a and b.
func prefixLen(a, b []byte) int {
	var i, length = 0, len(a)
	if len(b) < length {
		length = len(b)
	}
	for ; i < length; i++ {
		if a[i] != b[i] {
			break
		}
	}
	return i
}

// hasTerm returns whether a hex key has the terminator flag.
func hasTerm(s []byte) bool {
	return len(s) > 0 && s[len(s)-1] == 16
}

func compactToHex(compact []byte) []byte {
	if len(compact) == 0 {
		return compact
	}
	base := keybytesToHex(compact)
	// delete terminator flag
	if base[0] < 2 {
		base = base[:len(base)-1]
	}
	// apply odd flag
	chop := 2 - base[0]&1
	return base[chop:]
}
