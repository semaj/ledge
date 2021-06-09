package ledge

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	. "github.com/logrusorgru/aurora/v3"
	"github.com/montanaflynn/stats"
	"github.com/tevino/abool"
)

type Ledge struct {
	records     map[string][]float64
	recordsLock *sync.RWMutex
	logger      *log.Logger
	debug       *abool.AtomicBool
	stats       *abool.AtomicBool
}

var globalDebug = *abool.NewBool(true)
var globalStats = *abool.NewBool(true)

func New(prefixComponents ...string) *Ledge {
	prefix := fmt.Sprintf("[%s] ", strings.Join(prefixComponents, " "))
	if len(prefixComponents) == 0 {
		prefix = ""
	}
	return &Ledge{
		records:     make(map[string][]float64),
		recordsLock: &sync.RWMutex{},
		logger:      log.New(os.Stderr, fmt.Sprintf("%s", Green(prefix)), log.Lmsgprefix|log.Lmicroseconds),
		debug:       abool.NewBool(false),
		stats:       abool.NewBool(false),
	}
}

func DebugOn() {
	globalDebug.Set()
}

func DebugOff() {
	globalDebug.UnSet()
}

func StatsOn() {
	globalStats.Set()
}

func StatsOff() {
	globalStats.UnSet()
}

func (l *Ledge) DebugOff() {
	l.debug.UnSet()
}

func (l *Ledge) DebugOn() {
	l.debug.Set()
}

func (l *Ledge) StatsOff() {
	l.stats.UnSet()
}

func (l *Ledge) StatsOn() {
	l.stats.Set()
}

func (l *Ledge) Always(format string, v ...interface{}) {
	l.logger.Println(fmt.Sprintf(format, v...))
}

func (l *Ledge) Debug(format string, v ...interface{}) {
	if l.debug.IsSet() && globalDebug.IsSet() {
		formatString := fmt.Sprintf("%s %s", Cyan("[DEBUG]"), format)
		s := fmt.Sprintf(formatString, v...)
		l.logger.Println(s)
	}
}

func (l *Ledge) Panic(format string, v ...interface{}) {
	formatString := fmt.Sprintf("%s %s", Red("[PANIC]"), format)
	s := fmt.Sprintf(formatString, v...)
	l.logger.Panicln(s)
}

func (l *Ledge) Check(err error, format string, v ...interface{}) {
	if err != nil {
		l.Panic(format, v...)
	}
}

func toMillis(d time.Duration) float64 {
	return d.Seconds() * 1000.0
}

func (l *Ledge) Measure(tag string, f func()) {
	if l.stats.IsSet() && globalStats.IsSet() {
		t0 := time.Now()
		f()
		elapsed := time.Since(t0)
		tagString := fmt.Sprintf("[%s MEASURE]", tag)
		s := fmt.Sprintf("%s %s", Yellow(tagString), elapsed)
		l.logger.Println(s)
	}
}

func (l *Ledge) MeasureAbove(tag string, above time.Duration, f func()) {
	if l.stats.IsSet() && globalStats.IsSet() {
		t0 := time.Now()
		f()
		elapsed := time.Since(t0)
		if elapsed > above {
			tagString := fmt.Sprintf("[%s MEASURE]", tag)
			s := fmt.Sprintf("%s %s", Yellow(tagString), elapsed)
			l.logger.Println(s)
		}
	}
}

func (l *Ledge) Record(tag string, f func()) {
	if l.stats.IsSet() && globalStats.IsSet() {
		t0 := time.Now()
		f()
		elapsed := time.Since(t0)
		l.recordsLock.Lock()
		defer l.recordsLock.Unlock()
		if stats, ok := l.records[tag]; ok {
			l.records[tag] = append(stats, toMillis(elapsed))
		} else {
			l.records[tag] = []float64{toMillis(elapsed)}
		}
	}
}

func (l *Ledge) RecordAbove(tag string, above time.Duration, f func()) {
	if l.stats.IsSet() && globalStats.IsSet() {
		t0 := time.Now()
		f()
		elapsed := time.Since(t0)
		l.recordsLock.Lock()
		defer l.recordsLock.Unlock()
		if elapsed > above {
			if stats, ok := l.records[tag]; ok {
				l.records[tag] = append(stats, toMillis(elapsed))
			} else {
				l.records[tag] = []float64{toMillis(elapsed)}
			}
		}
	}
}

func (l *Ledge) ClearRecords(tag string) {
	l.recordsLock.Lock()
	defer l.recordsLock.Unlock()
	l.records[tag] = make([]float64, 0)
}

func (l *Ledge) Stats(tag string) {
	l.Count(tag)
	l.Min(tag)
	l.Median(tag)
	l.Perc(tag, 99)
	l.Max(tag)
	l.Mean(tag)
	l.Variance(tag)
}

func (l *Ledge) Count(tag string) {
	if l.stats.IsSet() && globalStats.IsSet() {
		l.recordsLock.RLock()
		records := l.records[tag]
		l.recordsLock.RUnlock()
		tagString := fmt.Sprintf("[%s COUNT]", tag)
		s := fmt.Sprintf("%s %d", Magenta(tagString), len(records))
		l.logger.Println(s)
	}
}

func (l *Ledge) Mean(tag string) {
	if l.stats.IsSet() && globalStats.IsSet() {
		l.recordsLock.RLock()
		records, ok := l.records[tag]
		l.recordsLock.RUnlock()
		if !ok || len(records) == 0 {
			return
		}
		r, e := stats.Mean(records)
		if e != nil {
			panic(e)
		}
		tagString := fmt.Sprintf("[%s MEAN]", tag)
		s := fmt.Sprintf("%s %f", Magenta(tagString), r)
		l.logger.Println(s)
	}
}

func (l *Ledge) Median(tag string) {
	if l.stats.IsSet() && globalStats.IsSet() {
		l.recordsLock.RLock()
		records, ok := l.records[tag]
		l.recordsLock.RUnlock()
		if !ok || len(records) == 0 {
			return
		}
		r, e := stats.Median(records)
		if e != nil {
			panic(e)
		}
		tagString := fmt.Sprintf("[%s MEDIAN]", tag)
		s := fmt.Sprintf("%s %f", Magenta(tagString), r)
		l.logger.Println(s)
	}
}

func (l *Ledge) Perc(tag string, perc float64) {
	if l.stats.IsSet() && globalStats.IsSet() {
		l.recordsLock.RLock()
		records, ok := l.records[tag]
		l.recordsLock.RUnlock()
		if !ok || len(records) == 0 {
			return
		}
		r, e := stats.PercentileNearestRank(records, perc)
		if e != nil {
			panic(e)
		}
		tagString := fmt.Sprintf("[%s PERC-%d]", tag, uint(perc))
		s := fmt.Sprintf("%s %f", Magenta(tagString), r)
		l.logger.Println(s)
	}
}

func (l *Ledge) Min(tag string) {
	if l.stats.IsSet() && globalStats.IsSet() {
		l.recordsLock.RLock()
		records, ok := l.records[tag]
		l.recordsLock.RUnlock()
		if !ok || len(records) == 0 {
			return
		}
		r, e := stats.Min(records)
		if e != nil {
			panic(e)
		}
		tagString := fmt.Sprintf("[%s MIN]", tag)
		s := fmt.Sprintf("%s %f", Magenta(tagString), r)
		l.logger.Println(s)
	}
}

func (l *Ledge) Max(tag string) {
	if l.stats.IsSet() && globalStats.IsSet() {
		l.recordsLock.RLock()
		records, ok := l.records[tag]
		l.recordsLock.RUnlock()
		if !ok || len(records) == 0 {
			return
		}
		r, e := stats.Max(records)
		if e != nil {
			panic(e)
		}
		tagString := fmt.Sprintf("[%s MAX]", tag)
		s := fmt.Sprintf("%s %f", Magenta(tagString), r)
		l.logger.Println(s)
	}
}

func (l *Ledge) Variance(tag string) {
	if l.stats.IsSet() && globalStats.IsSet() {
		l.recordsLock.RLock()
		records, ok := l.records[tag]
		l.recordsLock.RUnlock()
		if !ok || len(records) == 0 {
			return
		}
		r, e := stats.Variance(records)
		if e != nil {
			panic(e)
		}
		tagString := fmt.Sprintf("[%s VARIANCE]", tag)
		s := fmt.Sprintf("%s %f", Magenta(tagString), r)
		l.logger.Println(s)
	}
}
