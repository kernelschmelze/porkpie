package sidmap

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kernelschmelze/pkg/path"

	"github.com/kernelschmelze/porkpie/ids"

	"github.com/pkg/errors"
)

func (p *Plugin) loadMap(file string) error {

	if !utils.Exists(file) {
		return errors.New("file does not exist")
	}

	p.mapsGuard.Lock()

	utils.ReadLine(file, func(line string) {

		if len(line) == 0 || strings.HasPrefix(line, "#") {
			return
		}

		items := strings.Split(line, "||")

		if len(items) > 5 {

			gid := toUint64(items[0])
			sid := toUint64(items[1])
			rev := toUint64(items[2])

			// v2
			if gid > 0 && sid > 0 && rev > 0 {

				item := ids.SIDMapItem{
					GID:            gid,
					SID:            sid,
					Revision:       rev,
					Priority:       toUint64(items[4]),
					Classification: strings.TrimSpace(items[3]),
					Msg:            strings.TrimSpace(items[5]),
				}

				for i := 6; i < len(items); i++ {
					reference := items[i]
					reference = strings.TrimSpace(reference)
					item.Reference = append(item.Reference, reference)
				}

				p.addItem(item)

				return

			}

		}

	})

	p.mapsGuard.Unlock()

	return nil
}

func (p *Plugin) addItem(item ids.SIDMapItem) {

	index := p.getIndex(item.GID, item.SID)
	p.maps[index] = item
}

func (p *Plugin) getIndex(gid uint64, sid uint64) string {
	return fmt.Sprintf("%d:%d", gid, sid)
}

func toUint64(data string) uint64 {

	data = strings.TrimSpace(data)
	if u64, err := strconv.ParseUint(data, 10, 32); err == nil {
		return u64
	}

	return 0
}
