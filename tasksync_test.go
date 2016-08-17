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
	executionCount := 0
	synchronizer := NewSynchronizer(1 * time.Second)
	running := true

	synchronizer.Continous(func() {
		for running {
			time.Sleep(1 * time.Millisecond)
			executionCount++
		}
	}, func() {
		running = false
	})

	synchronizer.Run()

	time.Sleep(1 * time.Second)
	timeout := synchronizer.Stop()
	if timeout {
		t.Log("Unexpected timeout")
		t.Fail()
	}

	if executionCount < 10 {
		t.Log("Unexpected execution count")
		t.Log(executionCount)
		t.Fail()
	}
}
