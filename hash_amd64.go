// +build amd64

package gimli

//go:noescape
func roundAVX(state *[StateWords]uint32)

//go:noescape
func hashroundsAVX(state *[StateWords]uint32, src []byte, rounds int)

func init() {
	// TODO: test for presence of AVX
	round = roundAVX
	hashrounds = hashroundsAVX
}
