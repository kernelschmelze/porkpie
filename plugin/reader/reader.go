package reader

import (
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	log "github.com/kernelschmelze/pkg/logger"
	"github.com/kernelschmelze/pkg/path"

	"github.com/jasonish/go-unified2"
	"github.com/pkg/errors"
)

var (
	defaultReadInterval = 50 * time.Millisecond
)

func (p *Plugin) read(c config) {

	log.Debugf("%T start reader, path '%s', file prefix '%s' ", p, c.Path, c.FilePrefix)

	defer func() {

		log.Debugf("%T stop reader, path '%s', file prefix '%s' ", p, c.Path, c.FilePrefix)
		p.wgReader.Done()

	}()

	var getNextFile = func() (string, error) {

		nextFile, err := p.findNextFile(c)
		if err == nil && len(nextFile) > 0 {
			return nextFile, nil
		}

		return "", err
	}

	if len(c.CurrentFile) == 0 {
		c.CurrentFile, _ = getNextFile()
	}

	r, err := unified2.NewRecordReader(c.CurrentFile, c.Offset)
	if err != nil {
		log.Errorf("%T initi unified reader failed, err=%s", p, err)
		return
	}

	readInterval := defaultReadInterval

	for {

		select {

		case <-p.killReader:

			if r != nil {
				r.Close()
			}

			return

		case <-time.After(readInterval):

			record, err := r.Next()

			if err != nil {

				readInterval = 1 * time.Second

				if err == io.EOF {

					if nextFile, err := getNextFile(); err == nil && len(nextFile) > 0 {

						c.CurrentFile = nextFile
						c.Offset = 0

						r.Close()

						if r, err = unified2.NewRecordReader(c.CurrentFile, c.Offset); err != nil {
							log.Errorf("%T init unified reader failed, err=%s", p, err)
							return
						}

						readInterval = defaultReadInterval
					}

					continue

				}

				log.Errorf("%T unified reader read failed, err=%s", p, err)
				continue
			}

			readInterval = defaultReadInterval

			p.handleRecord(c.CurrentFile, r.Offset(), record)

		}

	}

}

func (p *Plugin) findNextFile(c config) (string, error) {

	var file string

	path, err := utils.ExpandPath(c.Path)
	if err != nil {
		return file, err
	}

	if !utils.Exists(path) {
		return file, errors.Errorf("path '%s' does not exist", path)
	}

	if files, err := ioutil.ReadDir(path); err == nil {

		currentFile := c.CurrentFile
		currentFile = strings.TrimPrefix(currentFile, path+"/")

		currentCounter := p.getCounter(currentFile, c.FilePrefix)

		for _, file := range files {

			name := file.Name()
			if strings.HasPrefix(name, c.FilePrefix) {

				if len(c.CurrentFile) == 0 {
					return path + "/" + name, nil
				}

				counter := p.getCounter(name, c.FilePrefix)
				if counter > currentCounter {
					return path + "/" + name, nil
				}

			}

		}

	}

	return file, nil
}

func (p *Plugin) getCounter(name string, prefix string) uint64 {

	counter := strings.TrimPrefix(name, prefix)
	counter = strings.TrimLeft(counter, ".")

	if u64, err := strconv.ParseUint(counter, 10, 32); err == nil {
		return u64
	}

	return 0

}
