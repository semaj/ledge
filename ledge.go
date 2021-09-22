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
	stdout      *log.Logger
	stderr      *log.Logger
	debug       *abool.AtomicBool
	stats       *abool.AtomicBool
}

func New(prefixComponents ...string) *Ledge {
	prefix := fmt.Sprintf("[%s] ", strings.Join(prefixComponents, " "))
	if len(prefixComponents) == 0 {
		prefix = ""
	}
	return &Ledge{
		records:     make(map[string][]float64),
		recordsLock: &sync.RWMutex{},
		stdout:      log.New(os.Stdout, fmt.Sprintf("%s", Green(prefix)), log.Lmsgprefix|log.Lmicroseconds),
		stderr:      log.New(os.Stderr, fmt.Sprintf("%s", BrightRed(prefix)), log.Lmsgprefix|log.Lmicroseconds),
		debug:       abool.NewBool(false),
		stats:       abool.NewBool(false),
	}
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

func (l *Ledge) Println(v ...interface{}) {
	l.stdout.Println(v...)
}

func (l *Ledge) Printf(format string, v ...interface{}) {
	l.stdout.Printf(format, v...)
}

func (l *Ledge) Debugf(format string, v ...interface{}) {
	if l.debug.IsSet() {
		formatString := fmt.Sprintf("%s %s", Cyan("[DEBUG]"), format)
		l.stderr.Printf(formatString, v...)
	}
}

func (l *Ledge) Debugln(v ...interface{}) {
	if l.debug.IsSet() {
		l.stderr.Println(append([]interface{}{Cyan("[DEBUG]")}, v...)...)
	}
}

func (l *Ledge) Panicf(format string, v ...interface{}) {
	formatString := fmt.Sprintf("%s %s", Red("[PANIC]"), format)
	l.stderr.Panicf(formatString, v...)
}

func (l *Ledge) Panicln(v ...interface{}) {
	l.stderr.Panicln(append([]interface{}{Red("[PANIC]")}, v...)...)
}

func (l *Ledge) Check(err error) {
	if err != nil {
		l.Panicf("%v", err)
	}
}

func (l *Ledge) CheckPrintf(err error, format string, v ...interface{}) {
	if err != nil {
		l.Panicf(format, v...)
	}
}

func (l *Ledge) CheckPrintln(err error, v ...interface{}) {
	if err != nil {
		v = append(v, err)
		l.Panicln(v)
	}
}

func toMillis(d time.Duration) float64 {
	return d.Seconds() * 1000.0
}

func (l *Ledge) Time(tag string, f func()) {
	t0 := time.Now()
	f()
	if l.stats.IsSet() {
		elapsed := time.Since(t0)
		tagString := fmt.Sprintf("[TIME %s]", tag)
		l.stdout.Printf("%s %s", Yellow(tagString), elapsed)
	}
}

func (l *Ledge) TimeAbove(tag string, above time.Duration, f func()) {
	t0 := time.Now()
	f()
	if l.stats.IsSet() {
		elapsed := time.Since(t0)
		if elapsed > above {
			tagString := fmt.Sprintf("[TIME-ABOVE %s]", tag)
			s := fmt.Sprintf("%s %s", Yellow(tagString), elapsed)
			l.stdout.Println(s)
		}
	}
}

func (l *Ledge) Record(tag string, f func()) {
	t0 := time.Now()
	f()
	if l.stats.IsSet() {
		elapsed := time.Since(t0)
		l.recordsLock.Lock()
		defer l.recordsLock.Unlock()
		if records, ok := l.records[tag]; ok {
			l.records[tag] = append(records, toMillis(elapsed))
		} else {
			l.records[tag] = []float64{toMillis(elapsed)}
		}
	}
}

func (l *Ledge) RecordAndPrint(tag string, f func()) {
	t0 := time.Now()
	f()
	if l.stats.IsSet() {
		elapsed := time.Since(t0)
		l.recordsLock.Lock()
		defer l.recordsLock.Unlock()
		tagString := fmt.Sprintf("[RECORD %s]", tag)
		s := fmt.Sprintf("%s %s", Yellow(tagString), elapsed)
		l.stdout.Println(s)
		if records, ok := l.records[tag]; ok {
			l.records[tag] = append(records, toMillis(elapsed))
		} else {
			l.records[tag] = []float64{toMillis(elapsed)}
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
	if l.stats.IsSet() {
		l.recordsLock.RLock()
		records := l.records[tag]
		l.recordsLock.RUnlock()
		tagString := fmt.Sprintf("[COUNT %s]", tag)
		l.stdout.Printf("%s %d", Magenta(tagString), len(records))
	}
}

func (l *Ledge) Mean(tag string) {
	if l.stats.IsSet() {
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
		tagString := fmt.Sprintf("[MEAN %s]", tag)
		l.stdout.Printf("%s %f", Magenta(tagString), r)
	}
}

func (l *Ledge) Median(tag string) {
	if l.stats.IsSet() {
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
		tagString := fmt.Sprintf("[MEDIAN %s]", tag)
		l.stdout.Printf("%s %f", Magenta(tagString), r)
	}
}

func (l *Ledge) Perc(tag string, perc float64) {
	if l.stats.IsSet() {
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
		tagString := fmt.Sprintf("[PERC-%d %s]", uint(perc), tag)
		l.stdout.Printf("%s %f", Magenta(tagString), r)
	}
}

func (l *Ledge) Min(tag string) {
	if l.stats.IsSet() {
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
		tagString := fmt.Sprintf("[MIN %s]", tag)
		l.stdout.Printf("%s %f", Magenta(tagString), r)
	}
}

func (l *Ledge) Max(tag string) {
	if l.stats.IsSet() {
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
		tagString := fmt.Sprintf("[MAX %s]", tag)
		l.stdout.Printf("%s %f", Magenta(tagString), r)
	}
}

func (l *Ledge) Variance(tag string) {
	if l.stats.IsSet() {
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
		tagString := fmt.Sprintf("[VARIANCE %s]", tag)
		l.stdout.Printf("%s %f", Magenta(tagString), r)
	}
}
