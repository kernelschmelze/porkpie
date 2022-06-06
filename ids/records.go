package ids

import (
	"fmt"
	"net"
	"time"

	"github.com/jasonish/go-unified2"
)

type Record struct {
	Drop    bool
	EventID uint32
	File    string
	Offset  int64

	// sidmap
	SIDMap SIDMapItem

	// geo ip
	Country string
	City    string

	// payload
	eventRecord     *unified2.EventRecord
	packetRecord    []*unified2.PacketRecord
	extraDataRecord []*unified2.ExtraDataRecord
}

type Packet struct {
	Data []byte
}

type Packets []Packet

type SIDMapItem struct {
	GID            uint64
	SID            uint64
	Revision       uint64
	Classification string
	Priority       uint64
	Msg            string
	Reference      []string
}

func (r *Record) IsValid() bool {
	return r.eventRecord != nil
}

func (r *Record) SetEventRecord(e *unified2.EventRecord) {
	r.eventRecord = e
}

func (r *Record) AddPacketRecord(e *unified2.PacketRecord) {
	if e != nil {
		r.packetRecord = append(r.packetRecord, e)
	}
}

func (r *Record) AddExtraDataRecord(e *unified2.ExtraDataRecord) {
	if e != nil {
		r.extraDataRecord = append(r.extraDataRecord, e)
	}
}

func (r *Record) GetGID() uint64 {
	if r.eventRecord != nil {
		return uint64(r.eventRecord.GeneratorId)
	}
	return 0
}

func (r *Record) GetSID() uint64 {
	if r.eventRecord != nil {
		return uint64(r.eventRecord.SignatureId)
	}
	return 0
}

func (r *Record) GetDestination() net.IP {
	if r.eventRecord != nil {
		return r.eventRecord.IpDestination
	}
	return net.IP{}
}

func (r *Record) GetSource() net.IP {
	if r.eventRecord != nil {
		return r.eventRecord.IpSource
	}
	return net.IP{}
}

func (r *Record) GetPriority() uint32 {
	if r.eventRecord != nil {
		return r.eventRecord.Priority
	}
	return 0
}

func (r *Record) GetImpact() uint8 {
	if r.eventRecord != nil {
		return r.eventRecord.Impact
	}
	return 0
}

func (r *Record) GetProtocol() string {

	if r.eventRecord == nil {
		return ""
	}

	switch r.eventRecord.Protocol {

	case 1:
		return "ICMP"

	case 6:
		return "TCP"

	case 17:
		return "UDP"

	}

	return fmt.Sprintf("%d", r.eventRecord.Protocol)

}

func (r *Record) GetTime() time.Time {

	if r.eventRecord == nil {
		return time.Unix(0, 0)
	}

	return time.Unix(int64(r.eventRecord.EventSecond), int64(r.eventRecord.EventMicrosecond*1000*1000))

}

func (r *Record) GetPackets() Packets {

	var packets Packets

	for i := range r.packetRecord {

		record := r.packetRecord[i]

		if record != nil && len(record.Data) > 0 {
			packet := Packet{
				Data: record.Data,
			}
			packets = append(packets, packet)
		}

	}

	for i := range r.extraDataRecord {

		record := r.extraDataRecord[i]

		if record != nil && len(record.Data) > 0 {
			packet := Packet{
				Data: record.Data,
			}
			packets = append(packets, packet)
		}

	}

	return packets

}

func (r *Record) IsLocal() bool {

	if ip := r.GetDestination(); IsPublicIP(ip) {
		return false
	}

	if ip := r.GetSource(); IsPublicIP(ip) {
		return false
	}

	return true

}
