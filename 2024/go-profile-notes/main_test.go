package main

import "testing"

func TestDoParallelTask(t *testing.T) {
	DoTasks()
}

func BenchmarkCalculate(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Calculate()
	}
}
