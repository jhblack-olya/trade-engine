package matching

import (
	"errors"
	"fmt"
)

var (
	tA = [8]byte{1, 2, 4, 8, 16, 32, 64, 128}
	tB = [8]byte{254, 253, 251, 247, 239, 223, 191, 127}
)

type Window struct {
	Min    int64
	Max    int64
	Cap    int64
	Bitmap Bitmap
}

type Bitmap []byte

func Set(m []byte, i int64, v bool) {
	index := i / 8
	bit := i % 8
	if v {
		m[index] = m[index] | tA[bit]
	} else {
		m[index] = m[index] & tB[bit]
	}

}
func (b Bitmap) Set(i int64, v bool) {
	Set(b, i, v)
}
func (w Window) put(val int64) error {
	if val <= w.Min {
		return errors.New(fmt.Sprintf("expired val %v, current Window [%v-%v]", val, w.Min, w.Max))
	} else if val > w.Max {
		delta := val - w.Max
		w.Min += delta
		w.Max += delta
		w.Bitmap.Set(val%w.Cap, true)
	}

	return nil
}
