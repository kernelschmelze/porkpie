package alias

import (
	"fmt"
	"reflect"

	"github.com/kernelschmelze/porkpie/ids"
)

var (
	tags = map[string]string{}
)

type Layout struct {
	Format string
	Fields []string
}

type Message struct {
	Local             bool
	Timestamp         int64
	Time              string `pp:"time"`
	EventID           uint32 `pp:"event"`
	Priority          uint32 `pp:"priority"`
	Impact            uint8  `pp:"impact"`
	Protocol          string `pp:"protocol"`
	Destination       string `pp:"destination"`
	Source            string `pp:"source"`
	GID               uint64 `pp:"gid"`
	SID               uint64 `pp:"sid"`
	SIDMsg            string `pp:"msg"`
	SIDRevision       uint64 `pp:"revision"`
	SIDClassification string `pp:"class"`
	SIDPriority       uint64 // ?
	MMCountry         string `pp:"mmcountry"`
	MMCity            string `pp:"mmcity"`
	ASNNumber         int    `pp:"asnnumber"`
	ASNCountryCode    string `pp:"asncountrycode"`
	ASNDescription    string `pp:"asndescription"`
}

func init() {
	buildTags(Message{})
}

func Get(r *ids.Record) Message {
	return Message{
		Local:             r.IsLocal(),
		Timestamp:         r.GetTime().Unix(),
		Time:              r.GetTime().Format("2006-01-02 15:04:05.000000"),
		EventID:           r.EventID,
		Priority:          r.GetPriority(),
		Impact:            r.GetImpact(),
		Protocol:          r.GetProtocol(),
		Destination:       r.GetDestination().String(),
		Source:            r.GetSource().String(),
		GID:               r.GetGID(),
		SID:               r.GetSID(),
		SIDMsg:            r.SIDMap.Msg,
		SIDRevision:       r.SIDMap.Revision,
		SIDClassification: r.SIDMap.Classification,
		SIDPriority:       r.SIDMap.Priority,
		MMCountry:         r.MM.Country,
		MMCity:            r.MM.City,
		ASNNumber:         r.ASN.Number,
		ASNCountryCode:    r.ASN.CountryCode,
		ASNDescription:    r.ASN.Description,
	}
}

func Format(v interface{}, format string, fields ...string) string {

	var values = []any{}

	reflectValue := reflect.ValueOf(v)
	if reflectValue.Kind() != reflect.Struct {
		return fmt.Sprintf("bad type '%s'", reflectValue.Kind())
	}

	for i := range fields {

		name, exist := tags[fields[i]]
		if !exist {
			name = fields[i]
		}

		value := reflectValue.FieldByName(name)
		values = append(values, value)

	}

	return fmt.Sprintf(format, values...)

}

func (l *Layout) IsValid() bool {
	if len(l.Format) == 0 || len(l.Fields) == 0 {
		return false
	}
	return true
}

func buildTags(v interface{}) error {

	tags = make(map[string]string)

	reflectType := reflect.TypeOf(v)
	if reflectType.Kind() != reflect.Struct {
		return fmt.Errorf("bad type '%s'", reflectType.Kind())
	}

	for i := 0; i < reflectType.NumField(); i++ {

		field := reflectType.Field(i)

		tag := field.Tag.Get("pp")
		if len(tag) == 0 {
			continue
		}

		tags[tag] = field.Name

	}

	return nil

}
