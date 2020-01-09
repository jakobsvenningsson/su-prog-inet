package main

import (
    "1.1/t"
    "time"
)

func main() {
    t1, t2 := t.New(1), t.New(2)

    // Start thread T1
    t1.Start()
    time.Sleep(time.Second * 5)

    // Start thread T2
    t2.Start()
    time.Sleep(time.Second * 5)
    
    // Stop thread T1
    t1.Stop()
    time.Sleep(time.Second * 5)

    // Stop thread T2
    t2.Stop()
}
