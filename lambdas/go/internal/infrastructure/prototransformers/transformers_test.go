package prototransformers

import (
	conflict_nightlightv1 "github.com/BaronBonet/conflict-nightlight/generated/conflict_nightlight/v1"
	"github.com/BaronBonet/conflict-nightlight/internal/core/domain"
	"github.com/go-playground/assert/v2"
	"testing"
)

func TestDomainToDownloadAndCropTifRequest(t *testing.T) {
	m := domain.Map{
		Date:    domain.Date{Month: 1, Year: 2022},
		MapType: domain.MapTypeDaily,
		Bounds:  domain.BoundsUkraineAndAround,
		Source:  domain.MapSource{URL: "http://example.com", MapProvider: domain.MapProviderEogdata},
	}
	p := DomainToProto(m)
	assert.Equal(t, p.MapType.String(), "MAP_TYPE_DAILY")
	assert.Equal(t, p.Bounds.String(), "BOUNDS_UKRAINE_AND_AROUND")
	assert.Equal(t, p.MapSource.MapProvider.String(), "MAP_PROVIDER_EOGDATA")
	assert.Equal(t, int(p.Date.Month), 1)
	assert.Equal(t, int(p.Date.Year), 2022)
}

func TestToSnakeCase(t *testing.T) {
	assert.Equal(t, CamelToSnakeCase("helloWorld"), "hello_world")
	assert.Equal(t, CamelToSnakeCase("helloFullWorld"), "hello_full_world")
}

func TestProtoToDomain(t *testing.T) {
	p := conflict_nightlightv1.Map{
		Date: &conflict_nightlightv1.Date{
			Day:   0,
			Month: 1,
			Year:  2023,
		},
		MapType: 1,
		Bounds:  1,
		MapSource: &conflict_nightlightv1.MapSource{
			MapProvider: 1,
			Url:         "example.com",
		},
	}
	m := ProtoToDomain(&p)
	assert.Equal(t, m.Bounds.String(), "BoundsUkraineAndAround")
	assert.Equal(t, m.Source.URL, "example.com")
	assert.Equal(t, m.Source.MapProvider.String(), "MapProviderEogdata")
	assert.Equal(t, m.Date.Day, 0)
	assert.Equal(t, m.Date.Month, 1)
	assert.Equal(t, m.Date.Year, 2023)
	assert.Equal(t, m.MapType.String(), "MapTypeDaily")

}
