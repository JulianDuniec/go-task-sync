package tasksync

import (
	"testing"
	"time"
)

/*
   Test executing
*/
func TestPeriodic(t *testing.T) {
	executionCount := 0
	synchronizer := NewSynchronizer(1 * time.Second)

	synchronizer.Every(10 * time.Millisecond).Do(func() {
		executionCount++
	})

	synchronizer.Run()
	time.Sleep(1 * time.Second)
	timeout := synchronizer.Stop()
	if timeout {
		t.Log("Unexpected timeout")
		t.Fail()
	}

	if executionCount < 10 && executionCount > 11 {
		t.Log("Unexpected execution count")
		t.Log(executionCount)
		t.Fail()
	}
}

/*
   Asserts that periodics are interrupted in their waiting fase
*/
func TestLongPeriodic(t *testing.T) {
	executionCount := 0
	synchronizer := NewSynchronizer(1 * time.Second)

	synchronizer.Every(1 * time.Hour).Do(func() {
		executionCount++
	})

	synchronizer.Run()
	time.Sleep(1 * time.Second)
	timeout := synchronizer.Stop()
	if timeout {
		t.Log("Unexpected timeout")
		t.Fail()
	}

	if executionCount > 1 {
		t.Log("Unexpected execution count")
		t.Log(executionCount)
		t.Fail()
	}
}

/*
   Run a continous function and specify how to stop it
*/
func TestContinous(t *testing.T) {
	r := &Runner{true, false, 0}
	synchronizer := NewSynchronizer(10 * time.Second)

	synchronizer.Continous(r.Run, r.Stop)

	synchronizer.Run()

	time.Sleep(1 * time.Second)

	timeout := synchronizer.Stop()
	if timeout {
		t.Log("Unexpected timeout")
		t.Fail()
	}
	if !r.graceful {
		t.Log("Ungraceful shutdown")
	}
	if r.counter < 10 {
		t.Log("Unexpected execution count")
		t.Log(r.counter)
		t.Fail()
	}
}

type Runner struct {
	running  bool
	graceful bool
	counter  int
}

func (r *Runner) Run() {
	for r.running {
		r.counter++
		time.Sleep(10 * time.Millisecond)
	}
	r.graceful = true
}

func (r *Runner) Stop() {
	r.running = false
}
