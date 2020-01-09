// Package t provides a struct T and methods which fulfils the requirements of assigment 1.1
package t 

import (
    "fmt"
    "time"
)

// Command represents instructions which is used to communicate with the thread associated with each struct T.
// The thread associated with each T struct is started in the New() instruction.
type Command uint32

// The availible instructions.
const (
    Stop Command = iota
    Start
)

// T represents a thread. The thread associated with T can be communicated with the channel `ch`.
type T struct {
    id uint32
    ch chan Command
}

// Start will start printing inside of the print-loop.
func (t *T) Start() {
    t.ch <- Start
}

// Stop will stop printing inside of the print-loop.
func (t *T) Stop() {
    t.ch <- Stop
}

// New initializes and returns the address of new struct T. 
// It also starts a new thread (print-loop) that will be responsible for printing to stdout. 
func New(id uint32) *T {
    t := &T{id, make(chan Command) }
    go t.loop()
    return t
}

func (t *T) loop() {
    stopCh := make(chan struct{})
    for command := range t.ch {
        switch command {
            case Start:
                go func(stop chan struct{}) {
                    for {
                        select {        
                        case <- time.After(time.Second * 1):
                            fmt.Printf("Tråd T%d: Tråd %d\n", t.id, t.id)
                        case <-stop:
                            return
                        }
                    }
                }(stopCh)
            case Stop:
                stopCh <- struct{}{}
                return
        }
    }
}
