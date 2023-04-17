package domain

import (
	"github.com/go-playground/assert/v2"
	"testing"
)

func TestStringToBounds(t *testing.T) {
	assert.Equal(t, StringToBounds("BoundsUkraineAndAround"), BoundsUkraineAndAround)
	assert.Equal(t, StringToBounds("bounds_ukraine_and_around"), BoundsUkraineAndAround)
	assert.Equal(t, StringToBounds("fake"), BoundsUnspecified)
}

func TestStringToMapType(t *testing.T) {
	assert.Equal(t, StringToMapType("MapTypeDaily"), MapTypeDaily)
	assert.Equal(t, StringToMapType("map_type_daily"), MapTypeDaily)
	assert.Equal(t, StringToMapType("fake"), MapTypeUnspecified)
}

func TestStringToMapProvider(t *testing.T) {
	assert.Equal(t, StringToMapProvider("MapProviderEogdata"), MapProviderEogdata)
	assert.Equal(t, StringToMapProvider("map_provider_eogdata"), MapProviderEogdata)
	assert.Equal(t, StringToMapProvider("fake"), MapProviderUnspecified)
}

func TestMap_String(t *testing.T) {
	m := Map{
		Bounds:  BoundsUkraineAndAround,
		MapType: MapTypeMonthly,
		Date:    Date{Day: 1, Month: 1, Year: 2022},
		Source: MapSource{
			MapProvider: MapProviderEogdata,
			URL:         "testMetadata",
		},
	}
	s := m.String()
	assert.Equal(t, s, "Monthly-UkraineAnd_2022-1-1")
}
