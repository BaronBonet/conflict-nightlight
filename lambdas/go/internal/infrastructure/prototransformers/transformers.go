package prototransformers

import (
	"fmt"
	conflict_nightlightv1 "github.com/BaronBonet/conflict-nightlight/generated/conflict_nightlight/v1"
	"github.com/BaronBonet/conflict-nightlight/internal/core/domain"
	"regexp"
	"strings"
)

func DomainToProto(m domain.Map) conflict_nightlightv1.Map {
	return conflict_nightlightv1.Map{
		Date: &conflict_nightlightv1.Date{
			Day:   int32(m.Date.Day),
			Month: int32(m.Date.Month),
			Year:  int32(m.Date.Year),
		},
		MapType: conflict_nightlightv1.MapType(conflict_nightlightv1.MapType_value[fmt.Sprintf("MAP_TYPE_%s", strings.ToUpper(m.MapType.String()))]),
		Bounds:  conflict_nightlightv1.Bounds(conflict_nightlightv1.Bounds_value[fmt.Sprintf("BOUNDS_%s", strings.ToUpper(ToSnakeCase(m.Bounds.String())))]),
		MapSource: &conflict_nightlightv1.MapSource{
			MapProvider: conflict_nightlightv1.MapProvider(conflict_nightlightv1.MapProvider_value[fmt.Sprintf("MAP_PROVIDER_%s", strings.ToUpper(m.Source.MapProvider.String()))]),
			Url:         m.Source.URL,
		},
	}
}

func ProtoToDomain(mp *conflict_nightlightv1.Map) domain.Map {
	return domain.Map{
		Date:    domain.Date{Day: int(mp.Date.Day), Month: int(mp.Date.Month), Year: int(mp.Date.Year)},
		MapType: domain.MapType(mp.MapType),
		Bounds:  domain.Bounds(mp.Bounds),
		Source:  domain.MapSource{MapProvider: domain.MapProvider(mp.MapSource.MapProvider), URL: mp.MapSource.Url},
	}
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
