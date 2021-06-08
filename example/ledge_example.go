package main

import (
	"fmt"
	"time"

	"github.com/semaj/ledge"
)

func main() {
	// ledge.New in external
	log := ledge.New("juicebox", "A")
	log.Always("Test %d", 1)
	log.Debug("Don't show me %d", 1)
	log.DebugOn()
	log.Debug("Show me %d", 1)
	ledge.DebugOff()
	log.Debug("Don't show me %d", 1)
	ledge.DebugOn()
	log.Debug("Show me %d", 2)
	log.Record("tag1", func() {})
	log.Count("tag1")
	log.StatsOn()
	log.Measure("one-off", func() {
		time.Sleep(13 * time.Millisecond)
	})
	log.MeasureAbove("dont-show", 13*time.Millisecond, func() {
		time.Sleep(10 * time.Millisecond)
	})
	log.MeasureAbove("show", 13*time.Millisecond, func() {
		time.Sleep(15 * time.Millisecond)
	})
	fmt.Println("Should 100 stats:")
	for i := 0; i < 100; i++ {
		log.Record("tag1", func() {
			time.Sleep(10 * time.Millisecond)
		})
	}
	log.Count("tag1")
	log.Mean("tag1")
	log.Median("tag1")
	log.Perc("tag1", 75)
	log.Min("tag1")
	log.Max("tag1")
	log.Variance("tag1")
	log.ClearRecords("tag1")
	fmt.Println("Should 50 stats:")
	for i := 0; i < 50; i++ {
		log.Record("tag1", func() {
			time.Sleep(2 * time.Millisecond)
		})
	}
	log.Stats("tag1")
	log.ClearRecords("tag1")
	fmt.Println("Should see just count 0:")
	for i := 0; i < 50; i++ {
		log.RecordAbove("tag1", 10*time.Millisecond, func() {
			time.Sleep(5 * time.Millisecond)
		})
	}
	log.Stats("tag1")
	fmt.Println("Should see just panic:")
	ledge.StatsOff()
	for i := 0; i < 50; i++ {
		log.Record("tag1", func() {
			time.Sleep(2 * time.Millisecond)
		})
	}
	log.Stats("tag1")
	log.Panic("PANICKING %f", 0.5)
}
