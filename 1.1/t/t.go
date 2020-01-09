// Package t ...
package t 

import (
    "fmt"
    "time"
)

type Command uint32

const (
    Stop Command = iota
    Start
)

type T struct {
    id uint32
    ch chan Command
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

func (t *T) Start() {
    t.ch <- Start
}

func (t *T) Stop() {
    t.ch <- Stop
}

func New(id uint32) *T {
    t := &T{id, make(chan Command) }
    go t.loop()
    return t
}
