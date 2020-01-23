package main

import (
    "fmt"
    "sync"
    "github.com/shirou/gopsutil/cpu"
)

/* CPU Helper Methods */
type lastCPUTimes struct {
    sync.Mutex
    times cpu.TimesStat
}
var lastCPU lastCPUTimes

func init() {
    t, _ := cpu.Times(false) // get totals, not per-cpu stats
    lastCPU.Lock()
    lastCPU.times = t[0]
    lastCPU.Unlock()
}

func getCPUTimes() (cpu.TimesStat, cpu.TimesStat, error) {
    currentTimes, err := cpu.Times(false) // get totals, not per-cpu stats
    if err != nil {
        return cpu.TimesStat{}, cpu.TimesStat{}, err
    }

    lastTimes := lastCPU.times

    lastCPU.Lock()
    lastCPU.times = currentTimes[0] // update lastTimes to the currentTimes
    lastCPU.Unlock()

    return currentTimes[0], lastTimes, nil
}

// main function to boot up everything
func main() {
    currentTime, lastTime, err := getCPUTimes()
    if err != nil {
            fmt.Println("fail")            
        }
    fmt.Println(currentTime)
    fmt.Println(lastTime)
}

