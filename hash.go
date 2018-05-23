package gimli

import (
	"hash"
	"encoding/binary"
)

const HashBlockSize = 16

// DefaultHashSize is the default size of the hash function.
const DefaultHashSize = 32

var _ hash.Hash = &Hash{}

// Hash is a Gimli-based hash function that can
// output arbitrary-length hashes using a "sponge"
// construction with the Gimli permutation.
//
// The zero value of Hash is safe to use; it
// will default to using an output size of
// DefaultHashSize.
type Hash struct {
	in      int64
	state   [StateWords]uint32
	outsize int
}

// NewHash returns a Hash with a 256-bit output.
func NewHash() *Hash {
	return &Hash{outsize: DefaultHashSize}
}

// NewHashSize returns a Hash with the given output size
// (in bytes). Most users should just use the default size.
func NewHashSize(size int) *Hash {
	if size < 1 {
		panic("invalid hash size")
	}
	return &Hash{outsize: size}
}

// BlockSize implements hash.Hash.BlockSize
//
// For the Gimli Hash, the block size is 16 bytes.
func (h *Hash) BlockSize() int { return HashBlockSize }

// Size implements hash.Hash.Size.
//
// If h is the zero value of Hash, then DefaultHashSize is returned.
func (h *Hash) Size() int {
	if h.outsize == 0 {
		h.outsize = DefaultHashSize
	}
	return h.outsize
}

// Reset implements hash.Hash.Reset
//
// The output size of the hash is preserved when Reset is called.
func (h *Hash) Reset() {
	size := h.outsize
	*h = Hash{outsize: size}
}

// xor a single byte into the state at a given position
func xorbyte(dst *[StateWords]uint32, v byte, pos uint) {
	// TODO: remove endianness assumptions here
	dst[(pos&^3)>>2] ^= uint32(v) << ((pos&3)<<3)
}

// finalize copies the state into a new state that
// is ready for "squeezing"
func (h *Hash) finalize(dst *[StateWords]uint32) {
	copy(dst[:], h.state[:])

	// input gets a trailing ^=0x1f;
	// last byte of the block gets a trailing ^=0x80
	xorbyte(dst, 0x1f, uint(h.in & (HashBlockSize - 1)))
	xorbyte(dst, 0x80, HashBlockSize-1)
	round(dst)
}

// Write implements hash.Hash.Write
//
// Write always returns len(b), nil
func (h *Hash) Write(b []byte) (int, error) {
	l := len(b)

	// if the previous write wasn't a multiple
	// of the hash block size, complete that block
	if pre := h.in & (HashBlockSize - 1); pre != 0 {
		written := int64(0)
		for i := range b {
			xorbyte(&h.state, b[i], uint(written+pre))
			written++
			if written+pre == HashBlockSize {
				round(&h.state)
				break
			}
		}
		b = b[written:]
	}

	aligned := len(b) >> 4
	hashrounds(&h.state, b, aligned)
	b = b[aligned<<4:]
	for i := range b {
		xorbyte(&h.state, b[i], uint(i))
	}

	h.in += int64(l)
	return l, nil
}

func put32(dst []byte, v uint32) {
	binary.LittleEndian.PutUint32(dst, v)
}

// copy the first 16 bytes of state into dst
func statehead(dst []byte, state *[StateWords]uint32) {
	put32(dst[0 :], state[0])
	put32(dst[4 :], state[1])
	put32(dst[8 :], state[2])
	put32(dst[12:], state[3])
}

// Sum implements hash.Hash.Sum
//
// Sum appends h.Size() bytes to b.
// It does not change the state of h.
func (h *Hash) Sum(b []byte) []byte {
	var tmp [StateWords]uint32
	h.finalize(&tmp)

	outsize := h.Size()
	aligned := (outsize+HashBlockSize-1)&^(HashBlockSize-1)
	outbuf := make([]byte, aligned)

	// "squeeze"
	// copy the first part of the state
	// into the output and keep permuting
	// until we've produced enough output
	off := 0
	for off < outsize {
		if off != 0 {
			round(&tmp)
		}
		statehead(outbuf[off:], &tmp)
		off += HashBlockSize
	}
	return append(b, outbuf[:outsize]...)
}
