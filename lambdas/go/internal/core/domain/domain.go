package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/BaronBonet/conflict-nightlight/internal/infrastructure"
)

type Map struct {
	Source  MapSource
	Date    Date
	MapType MapType
	Bounds  Bounds
}

func (m *Map) String() string {
	mapTypeString := strings.Replace(m.MapType.String(), "MapType", "", 1)
	boundsString := strings.Replace(m.Bounds.String(), "Bounds", "", 1)
	return fmt.Sprintf("%s-%s_%d-%d-%d", mapTypeString, boundsString[:10], m.Date.Year, m.Date.Month, m.Date.Day)
}

type LocalMap struct {
	Filepath string
	Map      Map
}

type PublishedMap struct {
	Url string
	Map Map
}

type MapSource struct {
	URL         string
	MapProvider MapProvider
}

type Date struct {
	Day   int
	Month time.Month
	Year  int
}

//go:generate stringer -type=MapProvider
type MapProvider int

const (
	MapProviderUnspecified MapProvider = iota
	MapProviderEogdata
)

func StringToMapProvider(s string) MapProvider {
	for i := MapProviderUnspecified; i <= MapProviderEogdata; i++ {
		if infrastructure.CleanStrings(i.String()) == infrastructure.CleanStrings(s) {
			return i
		}
	}
	return MapProviderUnspecified
}

//go:generate stringer -type=MapType
type MapType int

const (
	MapTypeUnspecified MapType = iota
	MapTypeDaily
	MapTypeMonthly
)

func StringToMapType(s string) MapType {
	for i := MapTypeUnspecified; i <= MapTypeMonthly; i++ {
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
//	in the externalmapsrepo
//
//go:generate stringer -type=Bounds
type Bounds int

const (
	BoundsUnspecified Bounds = iota
	BoundsUkraineAndAround
	BoundsGazaAndAround
)

func StringToBounds(s string) Bounds {
	for i := BoundsUnspecified; i <= BoundsGazaAndAround; i++ {
		if infrastructure.CleanStrings(i.String()) == infrastructure.CleanStrings(s) {
			return i
		}
	}
	return BoundsUnspecified
}

type SelectedDates struct {
	Months []time.Month
	Years  []int
}

type SyncMapRequest struct {
	SelectedDates SelectedDates
	MapType       MapType
	Bounds        Bounds
}
