package loading

import "testing"

func TestLoadingProgressLifecycle(t *testing.T) {
	lp := NewLoadingProgress()
	taskA, taskB := make(chan int), make(chan int)
	lp.AddTask(taskA, "catalog")
	lp.AddTask(taskB, "installed")
	if got, want := lp.String(" done"), "[1/2] catalog...\n[2/2] installed...\n"; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
	lp.MarkCompleted(taskB)
	if got, want := lp.String(" done"), "[1/2] catalog...\n[2/2] installed... done\n"; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
	lp.Reset()
	if got := lp.String(" done"); got != "" {
		t.Fatalf("String() after reset = %q", got)
	}
}
