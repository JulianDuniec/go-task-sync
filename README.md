# go-task-sync

Simple synchronization of background goroutines. Helpful when you want graceful shutdown. 

### Installation
    go get github.com/JulianDuniec/go-task-sync

### Example, periodic tasks:
    import("github.com/JulianDuniec/go-task-sync")
    ...
    // Create a synchronizer with 1 second timeout
    synchronizer := tasksync.NewSynchronizer(1 * time.Second)

    // Specify a task to run once every hour
    synchronizer.Every(1 * time.Hour).Do(func() {
        // Run some task
    })

    // Start running tasks (non-blocking)
    synchronizer.Run()

    // Convenience method - block until program receives quit signal
    tasksync.BlockUntilQuit()
    
    // Signal synchronizer to stop eg. when catching a quit signal
    // Blocks until either all tasks are finished or until timeout
    timeout := synchronizer.Stop()

### Example, continous task tasks:
    
    import("github.com/JulianDuniec/go-task-sync")
    ...

    // Create a synchronizer with 1 second timeout
    synchronizer := tasksync.NewSynchronizer(1 * time.Second)
    
    // Used to implement a custom shutdown behaviour
    running := true

    // First parameter is the task to execute, 
    // Second parameter is fired when asked to shut down
    synchronizer.Continous(func() {
        for running {
            // Do some work
        }
    }, func() {
        running = false
    })

    synchronizer.Run()  (non-blocking)

    tasksync.BlockUntilQuit()
    timeout := synchronizer.Stop()

### Example - HTTP-server

    // This example uses github.com/hydrogen18/stoppableListener to shut down
    // connections

    synchronizer := tasksync.NewSynchronizer(10 * time.Second)
    
    listener, _ := net.Listen("tcp", port)
    stoppableListener, _ := stoppableListener.New(listener)
    server := http.Server{}

    start := func() {
        server.Serve(stoppableListener)
    }

    stop := func() {
        stoppableListener.Stop()
    }


    synchronizer.Continous(start, stop)

    tasksync.BlockUntilQuit()
    timeout := synchronizer.Stop()

