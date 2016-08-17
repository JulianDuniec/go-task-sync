package tasksync

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Blocks until we catch either syscall.SIGINT or syscall.SIGTERM
func BlockUntilQuit() {
	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-stop:
		// Caught signal - unblock
		return
	}
}

// Factory method for synchronizer object
func NewSynchronizer(timeout time.Duration) *Synchronizer {
	wg := sync.WaitGroup{}
	tasks := make([]*task, 0)

	return &Synchronizer{
		tasks, &wg, timeout,
	}
}

// Synchronizer keeps track of a list of tasks
// and handles start/stop behaviour of these
type Synchronizer struct {
	tasks   []*task
	wg      *sync.WaitGroup
	timeout time.Duration
}

// Allows chaining Every(duration).Do
func (this periodic) Do(f emptyfunction) {
	this.ts.addTask(newTask(func(quitChan chan bool) {
		for {
			f()
			select {
			case <-quitChan:
				return
			case <-time.After(this.duration):

			}
		}
	}))
}

// Run a method periodically. This is a chainable call (returns periodic)
// which has a "Do" method to specify the method to run
func (this *Synchronizer) Every(duration time.Duration) periodic {
	return periodic{this, duration}
}

// Run a continous method with a custom
// stop function.
func (this *Synchronizer) Continous(run emptyfunction, stop emptyfunction) {
	this.addTask(newTask(func(quitChan chan bool) {
		donechn := make(chan bool)
		// Split new goroutine for the run-method.
		// Notify donechn when it finished running.
		go func() {
			run()
			donechn <- true
		}()

		// Await quit-msg, and run stop.
		select {
		case <-quitChan:
			stop()
		}

		// Await done from Run().
		select {
		case <-donechn:
			// run is done
			return
		}
	}))
}

// Add one to a waitgroup for
// every started task, which is later
// used to keep track if all tasks gracefully
// shut down.
func (this *Synchronizer) Run() {
	for _, t := range this.tasks {
		go func() {
			this.wg.Add(1)
			t.fn(t.quitChan)
			defer this.wg.Done()
		}()
	}
}

// Signal quit to all running tasks
// and await completion OR timeout
func (this *Synchronizer) Stop() bool {
	for _, t := range this.tasks {
		go func() {
			t.quitChan <- true
		}()

	}

	doneChn := make(chan bool)

	go func() {
		this.wg.Wait()
		doneChn <- true
	}()

	select {
	case <-time.After(this.timeout):
		return true
	case <-doneChn:
		return false
	}
}

func (this *Synchronizer) addTask(t *task) {
	this.tasks = append(this.tasks, t)
}

func newTask(taskFn taskfunction) *task {
	return &task{
		make(chan bool),
		taskFn,
	}
}

type task struct {
	quitChan chan bool
	fn       taskfunction
}

type emptyfunction func()

type taskfunction func(quitChan chan bool)

type periodic struct {
	ts       *Synchronizer
	duration time.Duration
}
