package main

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

func (t *T) start() {
    t.ch <- Start
}

func (t *T) stop() {
    t.ch <- Stop
}

func NewT(id uint32) *T {
    t := &T{id, make(chan Command) }
    go t.loop()
    return t
}

func main() {
    t1, t2 := NewT(1), NewT(2)

    // Start thread T1
    t1.start()
    time.Sleep(time.Second * 5)

    // Start thread T2
    t2.start()
    time.Sleep(time.Second * 5)
    
    // Stop thread T1
    t1.stop()
    time.Sleep(time.Second * 5)

    // Stop thread T2
    t2.stop()
}
