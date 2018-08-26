package gimli

import (
	"bytes"
	"encoding/hex"
	"io"
	"strconv"
	"testing"
)

func writeChunks(dst io.Writer, src []byte, size int) {
	if size == 0 {
		panic("bad size")
	}
	for len(src) >= size {
		l, err := dst.Write(src[:size])
		if err != nil {
			panic(err)
		}
		if l != size {
			panic("short write")
		}
		src = src[size:]
	}
	l, err := dst.Write(src)
	if err != nil {
		panic(err)
	}
	if l != len(src) {
		panic("short write")
	}
}

func TestHash(t *testing.T) {
	vectors := []struct {
		text, hexout string
	}{
		{
			"There's plenty for the both of us, may the best Dwarf win.",
			"4afb3ff784c7ad6943d49cf5da79facfa7c4434e1ce44f5dd4b28f91a84d22c8",
		},
		{
			"If anyone was to ask for my opinion, which I note they're not, I'd say we were taking the long way around.",
			"ba82a16a7b224c15bed8e8bdc88903a4006bc7beda78297d96029203ef08e07c",
		},
		{
			"Speak words we can all understand!",
			"8dd4d132059b72f8e8493f9afb86c6d86263e7439fc64cbb361fcbccf8b01267",
		},
		{
			"It's true you don't see many Dwarf-women. And in fact, they are so alike in voice and appearance, that they are often mistaken for Dwarf-men. And this in turn has given rise to the belief that there are no Dwarf-women, and that Dwarves just spring out of holes in the ground! Which is, of course, ridiculous.",
			"8887a5367d961d6734ee1a0d4aee09caca7fd6b606096ff69d8ce7b9a496cd2f",
		},
		{
			"",
			"b0634b2c0b082aedc5c0a2fe4ee3adcfc989ec05de6f00addb04b3aaac271f67",
		},
	}

	var h Hash
	for i := range vectors {
		h.Reset()
		in := vectors[i].text
		io.WriteString(&h, in)
		out := h.Sum(nil)
		wantout, err := hex.DecodeString(vectors[i].hexout)
		if err != nil {
			panic(err)
		}
		if !bytes.Equal(out, wantout) {
			t.Errorf("failed for test vector %d", i)
			t.Errorf("want: %x", wantout)
			t.Errorf("got:  %x", out)
			continue
		}
		if !bytes.Equal(out, h.Sum(nil)) {
			t.Error("h.Sum(...) not idempotent")
		}

		sum := Sum256([]byte(vectors[i].text))
		if !bytes.Equal(sum[:], wantout) {
			t.Errorf("Sum256 failed for test vector %d", i)
			t.Errorf("want: %x", wantout)
			t.Errorf("got: %x", sum[:])
			continue
		}

		// exercise writes at odd alignments
		for _, size := range []int{
			1, 3, 7, 9, 13, 15,
		} {
			h.Reset()
			writeChunks(&h, []byte(in), size)
			out = h.Sum(out[:0])
			if !bytes.Equal(out, wantout) {
				t.Errorf("test vector %d, alignment=%d failed", i, size)
				t.Errorf("want: %x", wantout)
				t.Errorf("got:  %x", out)
			}
		}
	}
}

func BenchmarkHash(b *testing.B) {
	sizes := []int{
		100,
		256,
		1024,
		4096,
	}
	data := make([]byte, sizes[len(sizes)-1])
	for _, size := range sizes {
		b.Run("Write" + strconv.Itoa(size), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(size))
			var h Hash
			for i := 0; i < b.N; i++ {
				h.Reset()
				h.Write(data[:size])
				h.Sum(data[:0])
			}
		})
		b.Run("Sum" + strconv.Itoa(size), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(size))
			for i := 0; i<b.N; i++ {
				_ = Sum256(data[:size])
			}
		})
	}
}
