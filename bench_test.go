package main

import (
	"cmp"
	"math/rand/v2"
	"testing"
)

func BenchmarkQuickSelect(b *testing.B) {
	rng := rand.New(rand.NewPCG(0, 0))
	myTestCase := make([]int, 3840*2160)
	for b.Loop() {
		b.StopTimer()
		for i := range myTestCase {
			myTestCase[i] = rng.Int()
		}
		b.StartTimer()
		quickSelect(myTestCase, len(myTestCase)/2, cmp.Compare)
	}
}
