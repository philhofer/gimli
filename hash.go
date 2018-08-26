package gimli

import (
	"hash"
)

var _ hash.Hash = &Hash{}

// Hash is a 256-bit hash function that
// uses the Gimli permutation.
type Hash struct {
	in    int64
	state state
	final state
}

// NewHash returns a Hash with a 256-bit output.
func NewHash256() *Hash {
	return &Hash{}
}

const (
	rate     = 16
	hashsize = stateBytes - rate
)

// BlockSize implements hash.Hash.BlockSize
//
// BlockSize always returns 16
func (h *Hash) BlockSize() int { return rate }

// Size implements hash.Hash.Size.
//
// Size always returns 32
func (h *Hash) Size() int { return hashsize }

// Reset implements hash.Hash.Reset
// Reset sets h to the zero value of Hash.
func (h *Hash) Reset() {
	*h = Hash{}
}

// xor a single byte into the state at a given position
func xorbyte(dst *state, v byte, pos uint) {
	dst.bytes()[pos] ^= v
}

// finalize copies the state into a new state that
// is ready for "squeezing"
func (h *Hash) finalize() {
	copy(h.final[:], h.state[:])

	// input gets a trailing ^=0x1f;
	// last byte of the block gets a trailing ^=0x80
	xorbyte(&h.final, 0x1f, uint(h.in&(rate-1)))
	xorbyte(&h.final, 0x80, rate-1)
	round(&h.final)
}

// Write implements hash.Hash.Write
//
// Write always returns len(b), nil
func (h *Hash) Write(b []byte) (int, error) {
	l := len(b)

	// if the previous write wasn't a multiple
	// of the hash block size, complete that block
	if pre := h.in & (rate - 1); pre != 0 {
		written := int64(0)
		for i := range b {
			xorbyte(&h.state, b[i], uint(written+pre))
			written++
			if written+pre == rate {
				round(&h.state)
				break
			}
		}
		b = b[written:]
	}

	aligned := len(b) >> 4
	if aligned > 0 {
		hashrounds(&h.state, b, aligned)
	}
	b = b[aligned<<4:]
	for i := range b {
		xorbyte(&h.state, b[i], uint(i))
	}

	h.in += int64(l)
	return l, nil
}

// Sum implements hash.Hash.Sum
//
// Sum appends h.Size() bytes to b.
// It does not change the state of h.
func (h *Hash) Sum(b []byte) []byte {
	var out [2 * rate]byte
	h.finalize()
	copy(out[0:], h.final.bytes()[:rate])
	round(&h.final)
	copy(out[rate:], h.final.bytes()[:rate])
	return append(b, out[:]...)
}

func Sum256(b []byte) (sum [32]byte) {
	var st state
	aligned := len(b) >> 4
	if aligned > 0 {
		hashrounds(&st, b, aligned)
		b = b[aligned<<4:]
	}
	for i := range b {
		xorbyte(&st, b[i], uint(i))
	}

	xorbyte(&st, 0x1f, uint(len(b))&(rate-1))
	xorbyte(&st, 0x80, rate-1)
	round(&st)

	copy(sum[:], st.bytes()[:rate])
	round(&st)
	copy(sum[rate:], st.bytes()[:rate])
	return sum
}
