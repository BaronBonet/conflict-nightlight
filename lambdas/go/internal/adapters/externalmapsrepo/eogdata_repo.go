package externalmapsrepo

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/BaronBonet/conflict-nightlight/internal/core/domain"
	"github.com/BaronBonet/conflict-nightlight/internal/core/ports"
	"github.com/gocolly/colly"
)

type eogdataRepo struct {
	logger  ports.Logger
	scraper *colly.Collector
}

func NewEogdataExternalMapsRepository(logger ports.Logger, tmpWriteDir string) ports.ExternalMapProviderRepo {
	return &eogdataRepo{logger: logger, scraper: colly.NewCollector(
		colly.CacheDir(tmpWriteDir),
		colly.MaxDepth(5),
		colly.AllowedDomains("eogdata.mines.edu"),
	)}
}

func (repo *eogdataRepo) GetProvider() domain.MapProvider {
	return domain.MapProviderEogdata
}

func (repo *eogdataRepo) List(ctx context.Context, bounds domain.Bounds, mapType domain.MapType) ([]domain.Map, error) {

	tileId, err := boundsToTileID(bounds)
	if err != nil {
		repo.logger.Error(ctx, "Could not transform bounds to tileId")
		return nil, err
	}

	var urls []string
	repo.scraper.OnRequest(func(r *colly.Request) {
		repo.logger.Debug(ctx, "Visiting url", "url", r.URL.String())
	})

	repo.scraper.OnHTML("tr.even, tr.odd", func(e *colly.HTMLElement) {
		url := e.ChildAttr("td.indexcolicon > a", "href")
		url = e.Request.AbsoluteURL(url)
		if isTgzLink(url) {
			urls = append(urls, url)
		} else {
			if err := repo.scraper.Visit(url); err != nil {
				repo.logger.Error(ctx, "Error when trying to visit a URL", "error", err, "url", url)
			}
		}
	})
	urlToScrape, err := mapTypeToUrl(mapType)
	if err != nil {
		return nil, err
	}
	if err := repo.scraper.Visit(*urlToScrape); err != nil {
		repo.logger.Error(ctx, "Error when trying to visit a URL", "error", err, "url", *urlToScrape)
		return nil, err
	}

	var sourceMaps []domain.Map
	for _, url := range urls {
		if getTileId(url) == tileId.String() &&
			!strings.Contains(url, "vcmslcfg") {
			date, err := extractDateFromMonthlyLink(url)
			if err != nil {
				return nil, err
			}
			sourceMaps = append(sourceMaps, domain.Map{
				Date:    *date,
				MapType: mapType,
				Bounds:  bounds,
				Source: domain.MapSource{
					MapProvider: domain.MapProviderEogdata,
					URL:         url,
				},
			})
		}
	}
	return sourceMaps, nil
}

// EogdataTileId denotes one of the 6 tiles that the MapProviderEogdata service uses
// our bounding shape file needs to be mapped to the tile that it is apart of.
type EogdataTileId string

const (
	EOGDATATILEID_TILE2 EogdataTileId = "75N060W"
)

func (id EogdataTileId) String() string {
	return string(id)
}

func boundsToTileID(bounds domain.Bounds) (*EogdataTileId, error) {
	var tileId EogdataTileId
	switch bounds {
	case domain.BoundsUkraineAndAround:
		tileId = EOGDATATILEID_TILE2
	case domain.BoundsGazaAndAround:
		tileId = EOGDATATILEID_TILE2
	default:
		return nil, fmt.Errorf("unknown Bounds: %s", bounds)
	}
	return &tileId, nil
}

func mapTypeToUrl(mapType domain.MapType) (*string, error) {
	var urlToScrape string
	switch mapType {
	case domain.MapTypeDaily:
		urlToScrape = "https://eogdata.mines.edu/nighttime_light/nightly/rade9d/"
	case domain.MapTypeMonthly:
		urlToScrape = "https://eogdata.mines.edu/nighttime_light/monthly/v10/"
	default:
		return nil, errors.New("unknown map type")
	}
	return &urlToScrape, nil
}

func extractDateFromMonthlyLink(link string) (*domain.Date, error) {
	segments := strings.Split(link, "/")
	year, err := strconv.ParseInt(segments[6], 10, 32)
	if err != nil {
		return nil, errors.New("Error while extracting the year. Error: " + err.Error())
	}
	month, err := strconv.ParseInt(segments[7][4:6], 10, 32)
	if err != nil {
		return nil, errors.New("Error while extracting the month. Error: " + err.Error())
	}
	return &domain.Date{
		Year:  int(year),
		Month: time.Month(month),
		Day:   1,
	}, nil
}

func getTileId(link string) string {
	r, _ := regexp.Compile("[0-9]{2}[NS][0-9]{3}[EW]")
	return r.FindString(link)
}

func isTgzLink(link string) bool {
	s := strings.Split(link, ".")
	return len(s) > 0 && s[len(s)-1] == "tgz"
}
