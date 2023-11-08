package logger

import (
	"log"
	"sync/atomic"
	"time"
)

type service struct {
	scannedCount  uint64
	deletedCount  uint64
	dumpedCount   uint64
	restoredCount uint64

	start       time.Time
	doneChannel chan bool
}

type Service interface {
	Start(reportPeriod time.Duration)
	Stop()
	IncScannedCounter(delta uint64)
	IncDumpedCounter(delta uint64)
	IncDeletedCounter(delta uint64)
	IncRestoredCounter(delta uint64)
	Report()
}

func CreateService() Service {
	return &service{
		doneChannel: make(chan bool),
	}
}

func (r *service) Start(reportPeriod time.Duration) {
	atomic.StoreUint64(&r.scannedCount, 0)
	atomic.StoreUint64(&r.deletedCount, 0)
	atomic.StoreUint64(&r.dumpedCount, 0)
	atomic.StoreUint64(&r.restoredCount, 0)

	r.start = time.Now()
	go r.report(reportPeriod)
}

func (r *service) Stop() {
	r.doneChannel <- true
}

func (r *service) IncScannedCounter(delta uint64) {
	atomic.AddUint64(&r.scannedCount, delta)
}

func (r *service) IncDeletedCounter(delta uint64) {
	atomic.AddUint64(&r.deletedCount, delta)
}

func (r *service) IncDumpedCounter(delta uint64) {
	atomic.AddUint64(&r.dumpedCount, delta)
}
func (r *service) IncRestoredCounter(delta uint64) {
	atomic.AddUint64(&r.restoredCount, delta)
}

func (r *service) Report() {
	log.Printf(
		"Scanned Keys: %d Dumped Entries: %d Restored Entries: %d Deleted Entries: %d after %s\n",
		atomic.LoadUint64(&r.scannedCount),
		atomic.LoadUint64(&r.dumpedCount),
		atomic.LoadUint64(&r.restoredCount),
		atomic.LoadUint64(&r.deletedCount),
		time.Since(r.start),
	)
}

func (r *service) report(reportTicker time.Duration) {
	timer := time.NewTicker(reportTicker)
	for {
		select {
		case <-timer.C:
			r.Report()
		case <-r.doneChannel:
			timer.Stop()
			break
		}
	}
}
