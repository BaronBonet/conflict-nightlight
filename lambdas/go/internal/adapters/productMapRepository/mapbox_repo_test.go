package productMapRepository

import (
	conflict_nightlightv1 "github.com/BaronBonet/conflict-nightlight/generated/conflict_nightlight/v1"
	"reflect"
	"testing"
)

func TestUpdateMapOptionsList(t *testing.T) {
	testCases := []*struct {
		name           string
		mapOptionsList []*conflict_nightlightv1.MapOptions
		newOption      conflict_nightlightv1.MapOptions
		expected       []*conflict_nightlightv1.MapOptions
	}{
		{
			name: "Add new option",
			mapOptionsList: []*conflict_nightlightv1.MapOptions{
				{DisplayName: "January 2021", Url: "mapbox://example1", Key: "2021-01"},
				{DisplayName: "February 2021", Url: "mapbox://example2", Key: "2021-02"},
			},
			newOption: conflict_nightlightv1.MapOptions{DisplayName: "March 2021", Url: "mapbox://example3", Key: "2021-03"},
			expected: []*conflict_nightlightv1.MapOptions{
				{DisplayName: "January 2021", Url: "mapbox://example1", Key: "2021-01"},
				{DisplayName: "February 2021", Url: "mapbox://example2", Key: "2021-02"},
				{DisplayName: "March 2021", Url: "mapbox://example3", Key: "2021-03"},
			},
		},
		{
			name: "Update existing option",
			mapOptionsList: []*conflict_nightlightv1.MapOptions{
				{DisplayName: "January 2021", Url: "mapbox://example1", Key: "2021-01"},
				{DisplayName: "February 2021", Url: "mapbox://example2", Key: "2021-02"},
			},
			newOption: conflict_nightlightv1.MapOptions{DisplayName: "February 2021", Url: "mapbox://example3", Key: "2021-02"},
			expected: []*conflict_nightlightv1.MapOptions{
				{DisplayName: "January 2021", Url: "mapbox://example1", Key: "2021-01"},
				{DisplayName: "February 2021", Url: "mapbox://example3", Key: "2021-02"},
			},
		},
		{
			name: "Add new option and sort with unsorted input",
			mapOptionsList: []*conflict_nightlightv1.MapOptions{
				{DisplayName: "March 2021", Url: "mapbox://example3", Key: "2021-03"},
				{DisplayName: "January 2021", Url: "mapbox://example1", Key: "2021-01"},
			},
			newOption: conflict_nightlightv1.MapOptions{DisplayName: "February 2021", Url: "mapbox://example2", Key: "2021-02"},
			expected: []*conflict_nightlightv1.MapOptions{
				{DisplayName: "January 2021", Url: "mapbox://example1", Key: "2021-01"},
				{DisplayName: "February 2021", Url: "mapbox://example2", Key: "2021-02"},
				{DisplayName: "March 2021", Url: "mapbox://example3", Key: "2021-03"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := updateMapOptionsList(tc.mapOptionsList, &tc.newOption)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}
