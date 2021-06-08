# Ledge

I made this library because I wanted a colorized logger with toggle-able
debugging, plus convenience functions for recording performance of various
functions.

I often find myself writing this code in research code to measure performance,
so I broke it out into a library.

The code in `example/ledge_example.go` demonstrates its usage:

```go
package main

import (
	"fmt"
	"time"

	"github.com/semaj/ledge"
)

func main() {
	// Creates a new book-keeper. Takes any number of strings.
	// This will produce prefix "[juicebox A]"
	log := ledge.New("juicebox", "A")
	// This will always print out (all to stderr)
	log.Always("Test %d", 1)
	// Debugging is off by default
	log.Debug("Don't show me %d", 1)
	// Turn it on
	log.DebugOn()
	log.Debug("Show me %d", 1)
	// We can globally turn off debugging for all ledges
	ledge.DebugOff()
	log.Debug("Don't show me %d", 1)
	// Turn global back on
	ledge.DebugOn()
	log.Debug("Show me %d", 2)
	// Stats is also off by default, these two lines will do nothing
	log.Record("tag1", func() {})
	log.Count("tag1")
	// Turn stats on
	log.StatsOn()
	// Measure one function and print out the results
	log.Measure("one-off", func() {
		time.Sleep(13 * time.Millisecond)
	})
	// Measure one function and if it takes longer than 13ms, print out
	log.MeasureAbove("dont-show", 13*time.Millisecond, func() {
		time.Sleep(10 * time.Millisecond)
	})
	// Won't print
	log.MeasureAbove("show", 13*time.Millisecond, func() {
		time.Sleep(15 * time.Millisecond)
	})
	// We can also save times for later based on a tag
	fmt.Println("Should 100 stats:")
	for i := 0; i < 100; i++ {
		log.Record("tag1", func() {
			time.Sleep(10 * time.Millisecond)
		})
	}
	// And print out the stats
	log.Count("tag1")
	log.Mean("tag1")
	log.Median("tag1")
	log.Perc("tag1", 75)
	log.Min("tag1")
	log.Max("tag1")
	log.Variance("tag1")
	// Clear records for a tag
	log.ClearRecords("tag1")
	fmt.Println("Should 50 stats:")
	for i := 0; i < 50; i++ {
		log.Record("tag1", func() {
			time.Sleep(2 * time.Millisecond)
		})
	}
	// This is a convenience function which prints out basic stats
	log.Stats("tag1")
	log.ClearRecords("tag1")
	fmt.Println("Should see just count 0:")
	// None of these will get recorded since they are below threshold
	for i := 0; i < 50; i++ {
		log.RecordAbove("tag1", 10*time.Millisecond, func() {
			time.Sleep(5 * time.Millisecond)
		})
	}
	// This will only print a count of zero
	log.Stats("tag1")
	fmt.Println("Should see just panic:")
	// Turning stats off means no record will get recorded or printed
	ledge.StatsOff()
	for i := 0; i < 50; i++ {
		log.Record("tag1", func() {
			time.Sleep(2 * time.Millisecond)
		})
	}
	log.Stats("tag1")
	// Our panic
	log.Panic("PANICKING %f", 0.5)
}
```
