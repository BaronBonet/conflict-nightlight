package prototransformers

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	conflict_nightlightv1 "github.com/BaronBonet/conflict-nightlight/generated/conflict_nightlight/v1"
	"github.com/BaronBonet/conflict-nightlight/internal/core/domain"
)

func DomainToProto(m domain.Map) conflict_nightlightv1.Map {
	return conflict_nightlightv1.Map{
		Date: &conflict_nightlightv1.Date{
			Day:   uint32(m.Date.Day),
			Month: uint32(m.Date.Month),
			Year:  uint32(m.Date.Year),
		},
		MapType: conflict_nightlightv1.MapType(conflict_nightlightv1.MapType_value[fmt.Sprintf(strings.ToUpper(camelToSnakeCase(m.MapType.String())))]),
		Bounds:  conflict_nightlightv1.Bounds(conflict_nightlightv1.Bounds_value[fmt.Sprintf(strings.ToUpper(camelToSnakeCase(m.Bounds.String())))]),
		MapSource: &conflict_nightlightv1.MapSource{
			MapProvider: conflict_nightlightv1.MapProvider(conflict_nightlightv1.MapProvider_value[fmt.Sprintf(strings.ToUpper(camelToSnakeCase(m.Source.MapProvider.String())))]),
			Url:         m.Source.URL,
		},
	}
}

func ProtoToDomain(mp *conflict_nightlightv1.Map) domain.Map {
	return domain.Map{
		Date:    domain.Date{Day: int(mp.Date.Day), Month: time.Month(mp.Date.Month), Year: int(mp.Date.Year)},
		MapType: domain.MapType(mp.MapType),
		Bounds:  domain.Bounds(mp.Bounds),
		Source:  domain.MapSource{MapProvider: domain.MapProvider(mp.MapSource.MapProvider), URL: mp.MapSource.Url},
	}
}

func ProtoToSyncMapsRequest(mp *conflict_nightlightv1.SyncMapRequest) domain.SyncMapRequest {
	return domain.SyncMapRequest{
		SelectedDates: domain.SelectedDates{
			Years:  ConvertInts[int](mp.SelectedYears),
			Months: ConvertInts[time.Month](mp.SelectedMonths),
		},
		MapType: domain.MapType(mp.MapType),
		Bounds:  domain.Bounds(mp.Bounds),
	}
}

func camelToSnakeCase(str string) string {
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := matchAllCap.ReplaceAllString(str, "${1}_${2}")
	return strings.ToLower(snake)
}

type Int interface {
	~int | ~int32 | ~int64
}

func ConvertInts[U, T Int](s []T) (out []U) {
	out = make([]U, len(s))
	for i := range s {
		out[i] = U(s[i])
	}
	return out
}
