package sloghelpers

import (
	"context"
	"os"
	"sync"
	"time"
)

type LogFileWriter struct {
	name string

	mu     sync.Mutex
	ticker *time.Ticker
	file   *os.File
}

func (o *LogFileWriter) Write(p []byte) (n int, err error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	return o.file.Write(p)
}

func (o *LogFileWriter) openFile() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.file != nil {
		o.file.Close()
	}

	t := time.Now()

	f, err := os.OpenFile(o.name+"-"+t.Format("2006-01-02-15")+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	o.file = f

	return nil
}

func (o *LogFileWriter) tickerLoop(ctx context.Context) {
	now := time.Now()

	nextHour := now.Truncate(time.Hour).Add(time.Hour)
	untilNextHour := nextHour.Sub(now) + time.Millisecond*10

	tReset := sync.OnceFunc(func() {
		o.ticker.Reset(time.Hour)
	})

	o.ticker = time.NewTicker(untilNextHour)
	defer o.ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			o.mu.Lock()
			defer o.mu.Unlock()

			o.file.Close()
			return
		case <-o.ticker.C:
			o.openFile()
			tReset()
		}
	}
}

func NewLogFileWriter(ctx context.Context, name string) *LogFileWriter {
	res := &LogFileWriter{name: name}
	res.openFile()
	go res.tickerLoop(ctx)

	return res
}
