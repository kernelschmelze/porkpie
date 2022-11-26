package ip2asn

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	gzip "github.com/klauspost/pgzip"
)

type asn struct {
	Start       string
	End         string
	Number      int
	CountryCode string
	Description string
}

func (p *Plugin) readDB(db string) error {

	if _, err := os.Stat(db); os.IsNotExist(err) {
		return err
	}

	file, err := os.Open(db)
	if err != nil {
		return err
	}

	var reader *bufio.Reader

	extension := filepath.Ext(db)

	switch extension {

	case ".gz":

		r, err := gzip.NewReader(file)
		if err != nil {
			return err
		}

		defer r.Close()

		reader = bufio.NewReader(r)

	case ".tsv":

		reader = bufio.NewReader(file)

	default:

		return fmt.Errorf("no reader interface for file extension '%s'", extension)

	}

	p.guard.Lock()
	defer p.guard.Unlock()

	p.lookup = make(map[string][]asn)

	for {

		line, err := reader.ReadBytes('\n')

		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}

		data := strings.Split(string(line), "\t")

		if len(data) < 5 {
			continue
		}

		number, err := strconv.Atoi(data[2])
		if err == nil && number == 0 {
			continue
		}

		asn := asn{
			data[0],
			data[1],
			number,
			data[3],
			strings.TrimSpace(data[4]), // expensive op
		}

		index := p.getTableIndex(data[0])

		p.lookup[index] = append(p.lookup[index], asn)

	}

	return nil

}

func (p *Plugin) getTableIndex(ip string) string {

	if offset := strings.IndexAny(ip, ".:"); offset > 0 {
		return ip[:offset]
	}

	return ip
}

func (p *Plugin) ipBetween(ip net.IP, start net.IP, end net.IP) bool {

	if ip == nil || start == nil || end == nil {
		return false
	}

	ip = ip.To16()
	start = start.To16()
	end = end.To16()

	if ip == nil || start == nil || end == nil {
		return false
	}

	if bytes.Compare(ip, start) >= 0 && bytes.Compare(ip, end) <= 0 {
		return true
	}

	return false
}

func (p *Plugin) getASN(ip net.IP) (asn, bool) {

	if ip == nil {
		return asn{}, false
	}

	index := p.getTableIndex(ip.String())

	p.guard.RLock()
	defer p.guard.RUnlock()

	if table, exist := p.lookup[index]; exist {

		for i := range table {

			if p.ipBetween(ip, net.ParseIP(table[i].Start), net.ParseIP(table[i].End)) {
				return table[i], true
			}

		}

	}

	return asn{}, false

}
