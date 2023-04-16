package domain

import (
	"fmt"
	"github.com/BaronBonet/conflict-nightlight/internal/infrastructure"
)

type Map struct {
	Date    Date
	MapType MapType
	Bounds  Bounds
	Source  MapSource
}

func (m *Map) String() string {
	return fmt.Sprintf("%s-%s_%d-%d-%d", m.MapType.String(), m.Bounds.String()[:10], m.Date.Year, m.Date.Month, m.Date.Day)
}

type LocalMap struct {
	Filepath string
	Map      Map
}

type MapSource struct {
	MapProvider MapProvider
	URL         string
}

type Date struct {
	Day   int
	Month int
	Year  int
}

//go:generate stringer -type=MapProvider
type MapProvider int

const (
	MapProviderUnspecified MapProvider = iota
	Eogdata
)

//go:generate stringer -type=MapType
type MapType int

const (
	MapTypeUnspecified MapType = iota
	Daily
	Monthly
)

func StringToMapType(s string) MapType {
	for i := MapTypeUnspecified; i <= Monthly; i++ {
		if infrastructure.CleanStrings(i.String()) == infrastructure.CleanStrings(s) {
			return i
		}
	}
	return MapTypeUnspecified
}

// Bounds is how the internal map was cropped from the source map
//
//	this is implemented via a shp file, to prevent the specific implementation of
//	geodata from bleeding into our domain you'll need to map the shape file name to the tile id
//	in the service
//
//go:generate stringer -type=Bounds
type Bounds int

const (
	BoundsUnspecified Bounds = iota
	UkraineAndAround
)

func StringToBounds(s string) Bounds {
	for i := BoundsUnspecified; i <= UkraineAndAround; i++ {
		if infrastructure.CleanStrings(i.String()) == infrastructure.CleanStrings(s) {
			return i
		}
	}
	return BoundsUnspecified
}

type SelectedDates struct {
	Months []int `validate:"required,dive,gt=0,lte=12"`
	Years  []int `validate:"required,dive,gt=2010"`
}
