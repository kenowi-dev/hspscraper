package hspscraper

import (
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/antchfx/xpath"
	"strings"
	"time"
)

const courseStateOpenx = "Vormerkliste"
const courseStateWaitingListx = "Warteliste"

var xPathCourseButtonTemplate = "//a[@id='K%s']/following-sibling::input"

var xPathSports = xpath.MustCompile("//main//table//li")
var xPathCourseRows = xpath.MustCompile("//table[@class='bs_kurse']/tbody/tr")
var xPathCourseNumber = xpath.MustCompile("//td[@class='bs_sknr']/text()")
var xPathCourseDetails = xpath.MustCompile("//td[@class='bs_sdet']/text()")
var xPathCourseDay = xpath.MustCompile("//td[@class='bs_stag']/text()")
var xPathCourseTime = xpath.MustCompile("//td[@class='bs_szeit']/text()")
var xPathCourseLocation = xpath.MustCompile("//td[@class='bs_sort']/a/text()")
var xPathCourseDates = xpath.MustCompile("//td[@class='bs_szr']/a")
var xPathCourseDate = xpath.MustCompile("//td[2]//text()")
var xPathCourseDateTime = xpath.MustCompile("//td[3]//text()")

var xPathCourseManagement = xpath.MustCompile("//td[@class='bs_skl']/text()")
var xPathCoursePrice = xpath.MustCompile("//td[@class='bs_spreis']///text()")
var xPathCourseState = xpath.MustCompile("//td[@class='bs_sbuch']/input")
var xPathCourseID = xpath.MustCompile("//td[@class='bs_sbuch']/a")

var hspUrlBase = "https://buchung.hochschulsport-hamburg.de"
var hspAtoZUrl = "https://www.hochschulsport.uni-hamburg.de/sportcampus/vona-z.html"
var hspSportTemplate = "https://buchung.hochschulsport-hamburg.de/angebote/Wintersemester_2023_2024/_%s.html"
var hspFlexiCardIndicator = "♥"

type CourseState string

const (
	CourseStateOpen        = "Vormerkliste"
	CourseStateWaitingList = "Warteliste"
)

type Course struct {
	Sport      string
	Number     string
	Details    string
	Day        string
	Time       string
	Location   string
	Dates      []CourseDate
	Management string
	Price      string
	State      CourseState
	id         string
}

type CourseDate struct {
	Date     time.Time
	Duration *time.Duration
}

type Sport struct {
	Name        string
	Href        string
	InFlexiCard bool
	ExtraInfo   string
}

func FindCourse(sport string, courseNumber string) (*Course, error) {
	doc, err := htmlquery.LoadURL(getHspSportUrl(sport))
	if err != nil {
		return nil, err
	}

	xPathCourseID, err := xpath.Compile(fmt.Sprintf(xPathCourseButtonTemplate, courseNumber))
	if err != nil {
		return nil, err
	}
	courseState := getValue(doc, xPathCourseID)
	courseID := getAtrValue(doc, xPathCourseID, "name")

	return &Course{
		Sport:  sport,
		Number: courseNumber,
		id:     courseID,
		State:  CourseState(courseState),
	}, nil
}
func GetAllCoursesWithDates(sport *Sport) ([]*Course, error) {
	return getAllCourses(sport, true)
}
func GetAllCourses(sport *Sport) ([]*Course, error) {
	return getAllCourses(sport, false)
}
func getAllCourses(sport *Sport, getDates bool) ([]*Course, error) {
	doc, err := htmlquery.LoadURL(sport.Href)
	if err != nil {
		return nil, err
	}
	trs := htmlquery.QuerySelectorAll(doc, xPathCourseRows)

	var courses = make([]*Course, 0)
	for _, tr := range trs {
		number := ""
		if n := htmlquery.QuerySelector(tr, xPathCourseNumber); n != nil {
			number = n.Data
		}
		details := ""
		if n := htmlquery.QuerySelector(tr, xPathCourseDetails); n != nil {
			details = n.Data
		}
		day := ""
		if n := htmlquery.QuerySelector(tr, xPathCourseDay); n != nil {
			day = n.Data
		}
		t := ""
		if n := htmlquery.QuerySelector(tr, xPathCourseTime); n != nil {
			t = n.Data
		}
		location := ""
		if n := htmlquery.QuerySelector(tr, xPathCourseLocation); n != nil {
			location = n.Data
		}

		var dates []CourseDate = nil
		if getDates {
			dates = make([]CourseDate, 0)
			if n := htmlquery.QuerySelector(tr, xPathCourseDates); n != nil {
				href := htmlquery.SelectAttr(n, "href")
				if href != "" {
					if dates, err = parseDates(href); err != nil {
						return nil, err
					}
				}
			}
		}

		management := ""
		if n := htmlquery.QuerySelector(tr, xPathCourseManagement); n != nil {
			management = n.Data
		}
		price := ""
		if n := htmlquery.QuerySelector(tr, xPathCoursePrice); n != nil {
			price = n.Data
		}
		state := ""
		if n := htmlquery.QuerySelector(tr, xPathCourseState); n != nil {
			state = htmlquery.SelectAttr(n, "value")
		}
		id := ""
		if n := htmlquery.QuerySelector(tr, xPathCourseID); n != nil {
			id = htmlquery.SelectAttr(n, "id")
		}

		course := Course{
			Sport:      sport.Name,
			Number:     number,
			Details:    details,
			Day:        day,
			Time:       t,
			Location:   location,
			Dates:      dates,
			Management: management,
			Price:      price,
			State:      CourseState(state),
			id:         id,
		}
		courses = append(courses, &course)
	}

	return courses, nil
}

func parseDates(href string) ([]CourseDate, error) {
	datesDoc, err := htmlquery.LoadURL(hspUrlBase + href)
	if err != nil {
		return nil, err
	}
	trs := htmlquery.QuerySelectorAll(datesDoc, xPathCourseRows)
	courseDates := make([]CourseDate, 0)
	for _, tr := range trs {
		var date time.Time
		if n := htmlquery.QuerySelector(tr, xPathCourseDate); n != nil {
			date, err = time.Parse("02.01.2006", n.Data)
			if err != nil {
				return nil, err
			}
		}
		var duration *time.Duration = nil
		if n := htmlquery.QuerySelector(tr, xPathCourseDateTime); n != nil {
			f, t, ok := strings.Cut(n.Data, "-")
			if !ok {
				goto creation
			}
			from, err := time.Parse("15.04", f)
			if err != nil {
				goto creation
			}
			to, err := time.Parse("15.04", t)
			if err != nil {
				goto creation
			}
			d := to.Sub(from)
			duration = &d
		}
	creation:
		courseDate := CourseDate{
			Date:     date,
			Duration: duration,
		}
		courseDates = append(courseDates, courseDate)
	}
	return courseDates, nil
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
