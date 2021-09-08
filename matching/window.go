/*
Copyright (C) 2021 Global Art Exchange, LLC (GAX). All Rights Reserved.
You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/

package matching

import (
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

func Get(m []byte, i int64) bool {
	return m[i/8]&tA[i%8] != 0
}

func (b Bitmap) Get(i int64) bool {
	return Get(b, i)
}

func newWindow(min, max int64) Window {
	return Window{
		Min:    min,
		Max:    max,
		Cap:    max - min,
		Bitmap: New(max - min),
	}
}

func New(l int64) Bitmap {
	return NewSlice(l)
}

func NewSlice(l int64) []byte {
	remainder := l % 8
	if remainder != 0 {
		remainder = 1
	}
	return make([]byte, l/8+remainder)
}

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
		return fmt.Errorf("expired val %v, current Window [%v-%v]", val, w.Min, w.Max)
	} else if val > w.Max {
		delta := val - w.Max
		w.Min = w.Min + delta
		w.Max = w.Max + delta
		w.Bitmap.Set(val%w.Cap, true)
	} else if w.Bitmap.Get(val % w.Cap) {
		return fmt.Errorf("existed val %v", val)
	} else {
		w.Bitmap.Set(val%w.Cap, true)
	}

	return nil
}
