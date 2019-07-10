MARS - Minimal Asynchronous Runtime System
==========================================

MARS is an experimental task dispatcher library, which allows limiting concurrency.

In principle, the golang runtime already provides everything that's needed via go routines,
except I find myself writing the same few functions of boilerplate again and again.

Features:
* back pressure on task creation (currently blocks)
* error propagation
* extensible task types


## Installation

    go get github.com/michaelw/mars

## Usage

```go
// create pool with 4 workers, back pressure at 100 tasks
M := mars.New(context.Background(), mars.Config{
        Concurrency: 4,
        QueueSize:   100,
})

for _, w := range work {
    w := w
    M.SubmitFn(func(ctx context.Context) error {
        // Do concurrent work here.
        //
        // Returning an error terminates the pool early (also via ctx), and
        // the error is propagated to M.Wait() below.
        return nil
    })
}

M.Shutdown()
// wait for shutdown
if err := M.Wait(); err != nil {
    log.Fatal(err)
}
```


## TODO

* tests
* CI

### Features

* add synchronization mechanisms, e.g., task dependencies
* deadlock detection
* debugging
