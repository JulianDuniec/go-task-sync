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
	wg := &sync.WaitGroup{}
	taskMutex := &sync.Mutex{}
	tasks := make([]*task, 0)

	return &Synchronizer{
		tasks, wg, timeout, taskMutex,
	}
}

// Synchronizer keeps track of a list of tasks
// and handles start/stop behaviour of these
type Synchronizer struct {
	tasks []*task
	// Waitgroup used to await stop of
	// all tasks
	wg *sync.WaitGroup

	// Timeout when waiting for task completion
	timeout time.Duration

	// Lock for the task-list
	taskMutex *sync.Mutex
}

// Allows chaining Every(duration).Do
func (this periodic) Do(task emptyfunction) {
	this.ts.addTask(newTask(func(quitChan chan bool) {
		for {
			// Keep track of duration of task
			t := time.Now()

			task()

			// Calculate delta-time and
			// use the difference as wait-time, in
			// order to match the desired interval
			// between executions of the task
			dt := time.Now().Sub(t)
			waitDuration := this.duration - dt

			// If waitDuration is negative, the task is
			// longer than the desired interval. This also means that the next execution will
			// be made immediately, unless we receive from quitchan.
			select {
			case <-quitChan:
				return
			case <-time.After(waitDuration):
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
	this.taskMutex.Lock()
	for _, t := range this.tasks {
		// NOTE: It's very important to parameterize
		// the goroutine, otherwise the task-value
		// will be overwritten by the next
		// value in the iteration
		go func(t *task) {
			this.wg.Add(1)
			t.fn(t.quitChan)
			defer this.wg.Done()
		}(t)
	}
	this.taskMutex.Unlock()
}

// Signal quit to all running tasks
// and await completion OR timeout
func (this *Synchronizer) Stop() bool {

	// Signal quit to all tasks
	this.taskMutex.Lock()
	for _, t := range this.tasks {
		// NOTE: It's very important to parameterize
		// the goroutine, otherwise the task-value
		// will be overwritten by the next
		// value in the iteration
		go func(t *task) {
			t.quitChan <- true
		}(t)

	}
	this.taskMutex.Unlock()

	doneChn := make(chan bool)

	// Await all tasks to complete and
	// signal completion to done-channel
	go func() {
		this.wg.Wait()
		doneChn <- true
	}()

	// Wait for either completion of all tasks,
	// or timeout
	select {
	case <-time.After(this.timeout):
		return true
	case <-doneChn:
		return false
	}
}

func (this *Synchronizer) addTask(t *task) {
	this.taskMutex.Lock()
	this.tasks = append(this.tasks, t)
	this.taskMutex.Unlock()
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
