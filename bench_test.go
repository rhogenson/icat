package main

import (
	"cmp"
	"math/rand/v2"
	"testing"
)

func BenchmarkQuickSelect(b *testing.B) {
	r := rand.New(rand.NewPCG(0, 0))
	testSlice := make([]int, 3840*2160)
	for i := range testSlice {
		testSlice[i] = r.Int()
	}
	myTestSlice := make([]int, len(testSlice))
	for b.Loop() {
		b.StopTimer()
		copy(myTestSlice, testSlice)
		b.StartTimer()
		quickSelect(myTestSlice, len(myTestSlice)/2, cmp.Compare)
	}
}
