package externalmapsrepository

import (
	"testing"
)

func TestGetTileId(t *testing.T) {
	testCases := []struct {
		link     string
		expected string
	}{
		{
			link:     "https://eogdata.mines.edu/nighttime_light/monthly/v10/2013/201301/vcmcfg/SVDNB_npp_20130101-20130131_00N060W_vcmcfg_v10_c201605121529.tgz",
			expected: "00N060W",
		},
		{
			link:     "https://eogdata.mines.edu/nighttime_light/monthly/v10/2019/201908/vcmcfg/SVDNB_npp_20190801-20190831_75N180W_vcmcfg_v10_c201909051300.tgz",
			expected: "75N180W",
		},
		{
			link:     "https://eogdata.mines.edu/nighttime_light/monthly/v10/2022/202208/NOAA-20/vcmcfg/SVDNB_j01_20220801-20220831_00N060W_vcmcfg_v10_c202209231200.tgz",
			expected: "00N060W",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.link, func(t *testing.T) {
			result := getTileId(tc.link)
			if result != tc.expected {
				t.Errorf("Expected %s but got %s", tc.expected, result)
			}
		})
	}
}
