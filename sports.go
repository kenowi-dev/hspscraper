package hspscraper

import (
	"errors"
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/antchfx/xpath"
	"strings"
)

type Sport struct {
	Name        string
	Href        string
	InFlexiCard bool
	ExtraInfo   string
}

var xPathSports = xpath.MustCompile("//main//table//li")

var hspAtoZUrl = "https://www.hochschulsport.uni-hamburg.de/sportcampus/vona-z.html"
var hspSportTemplate = "https://buchung.hochschulsport-hamburg.de/angebote/Wintersemester_2023_2024/_%s.html"
var hspFlexiCardIndicator = "♥"

func FindSport(sport string) (*Sport, error) {
	// This needs to iterate over all sports, since that is the only place where the extraInfo is available.
	sports, err := GetAllSports()
	if err != nil {
		return nil, err
	}

	for _, s := range sports {
		if strings.Contains(s.Name, sport) {
			return s, nil
		}
	}
	return nil, errors.New("no sport found")
}

func GetAllFlexiCardSports() ([]*Sport, error) {
	flexiSports := make([]*Sport, 0)
	sports, err := GetAllSports()
	if err != nil {
		return nil, err
	}
	for _, sport := range sports {
		if sport.InFlexiCard {
			flexiSports = append(flexiSports, sport)
		}
	}
	return flexiSports, nil
}

func GetAllSports() ([]*Sport, error) {
	doc, err := htmlquery.LoadURL(hspAtoZUrl)
	if err != nil {
		return nil, err
	}

	lis := htmlquery.QuerySelectorAll(doc, xPathSports)
	// Cannot use len(lis), since improper sports sites will be ignored
	sports := make([]*Sport, 0)
	for _, li := range lis {
		a := li.FirstChild
		if a == nil {
			continue
		}
		sportName := a.FirstChild.Data
		extraInfo := ""
		if a.NextSibling != nil {
			extraInfo = a.NextSibling.Data
		}
		inFlexiCard := strings.Contains(extraInfo, hspFlexiCardIndicator)
		href := htmlquery.SelectAttr(a, "href")
		if !strings.HasPrefix(href, "https://") {
			// If the prefix does not exist, the sport is not a proper booking side.
			// e.g. /sportcampus/kinder.html
			continue
			//href = hspUrlBase + href
		}
		sport := Sport{
			Name:        sportName,
			Href:        href,
			InFlexiCard: inFlexiCard,
			ExtraInfo:   extraInfo,
		}
		sports = append(sports, &sport)
	}
	return sports, nil
}

func getHspSportUrl(sport string) string {
	return fmt.Sprintf(hspSportTemplate, strings.ReplaceAll(sport, " ", "_"))
}
