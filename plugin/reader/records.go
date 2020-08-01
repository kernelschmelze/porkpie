package reader

import (
	"time"

	"github.com/kernelschmelze/porkpie/ids"

	manager "github.com/kernelschmelze/pkg/plugin/manager"

	"github.com/jasonish/go-unified2"
)

type job struct {
	File   string
	Offset int64
	Data   interface{}
}

var (
	defaultTreshold = 1 * time.Second
)

func (p *Plugin) handleRecord(currentFile string, offset int64, record interface{}) {

	job := job{
		File:   currentFile,
		Offset: offset,
		Data:   record,
	}

	p.jobs <- job

}

func (p *Plugin) handleJobs() {

	defer func() {
		p.wgJobs.Done()
	}()

	var r *ids.Record
	idleMode := 60 * time.Second
	threshold := defaultTreshold
	for {

		select {

		case <-p.kill:
			return

		case <-time.After(threshold):

			if r != nil && r.EventID > 0 {
				p.handleEvent(r)
				r = &ids.Record{}
			}

			if threshold == idleMode {
				p.storeBookmark()
				continue
			}

			// idle mode
			threshold = idleMode

		case job := <-p.jobs:

			switch data := job.Data.(type) {

			case *unified2.EventRecord:

				if r != nil {
					p.handleEvent(r)
				}

				r = &ids.Record{
					File:    job.File,
					Offset:  job.Offset,
					EventID: data.EventId,
				}

				r.SetEventRecord(data)

				threshold = defaultTreshold

			case *unified2.PacketRecord:

				if r == nil {
					continue
				}

				if data.EventId != r.EventID {
					continue
				}

				r.AddPacketRecord(data)
				threshold = defaultTreshold

			case *unified2.ExtraDataRecord:

				if r == nil {
					continue
				}

				if data.EventId != r.EventID {
					continue
				}

				r.AddExtraDataRecord(data)
				threshold = defaultTreshold

			}

		}

	}
}

func (p *Plugin) handleEvent(record *ids.Record) {

	manager.Dispatch(record)

}

func (p *Plugin) processRecord(v interface{}) error {

	switch data := v.(type) {

	case *ids.Record:
		p.setBookmark(data.File, data.Offset)

	}

	return nil
}
