package gimli

import (
	"encoding/binary"
	"unsafe"
)

const (
	stateBytes = 384 / 8
)

type state [stateBytes / 4]uint32

var round = portableRound

var hashrounds = portableRounds

func le32(b []byte) uint32 {
	return binary.LittleEndian.Uint32(b)
}

func (s *state) bytes() *[stateBytes]byte {
	return (*[stateBytes]byte)(unsafe.Pointer(s))
}

// mix blocks of 16 bytes into the state 'blocks' times
func portableRounds(st *state, src []byte, blocks int) {
	for i := 0; i < blocks; i++ {
		st[0] ^= le32(src[(i*16)+0:])
		st[1] ^= le32(src[(i*16)+4:])
		st[2] ^= le32(src[(i*16)+8:])
		st[3] ^= le32(src[(i*16)+12:])
		round(st)
	}
}

func portableRound(b *state) {
	for round := 24; round > 0; round-- {
		for col := 0; col < 4; col++ {
			x := b[col]
			y := b[col+4]
			z := b[col+8]
			x = (x << 24) | (x >> 8)
			y = (y << 9) | (y >> 23)

			b[col+8] = x ^ (z << 1) ^ ((y & z) << 2)
			b[col+4] = y ^ x ^ ((x | z) << 1)
			b[col+0] = z ^ y ^ ((x & y) << 3)
		}
		switch round & 3 {
		case 0:
			b[0], b[1] = b[1], b[0]
			b[2], b[3] = b[3], b[2]
			b[0] ^= (uint32(0x9e377900) | uint32(round))
		case 2:
			b[0], b[2] = b[2], b[0]
			b[1], b[3] = b[3], b[1]
		}
	}
}
