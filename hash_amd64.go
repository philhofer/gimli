// +build amd64

package gimli

//go:noescape
func roundAVX(st *state)

//go:noescape
func hashroundsAVX(st *state, src []byte, rounds int)

func init() {
	// TODO: test for presence of AVX
	round = roundAVX
	hashrounds = hashroundsAVX
}
