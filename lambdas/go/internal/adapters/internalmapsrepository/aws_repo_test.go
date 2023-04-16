package internalmapsrepository

import (
	"github.com/BaronBonet/conflict-nightlight/internal/core/domain"
	"github.com/go-playground/assert/v2"
	"testing"
)

func TestAWSMapsRepo_extractDateFromFilepath(t *testing.T) {
	d, _ := extractDateFromFilepath("2022_1_0.tif")
	assert.Equal(t, domain.Date{
		Day:   0,
		Month: 1,
		Year:  2022,
	}, d)
}

func TestAWSMapsRepo_createFilepathFromDate(t *testing.T) {
	n := createFilepathFromMap(domain.Map{MapType: domain.Daily, Bounds: domain.UkraineAndAround, Source: domain.MapSource{
		MapProvider: domain.Eogdata,
		URL:         "test.com",
	}, Date: domain.Date{
		Day:   0,
		Month: 1,
		Year:  2022,
	}})
	assert.Equal(t, n, "Eogdata/UkraineAndAround/Daily/2022_1_0.tif")
}
