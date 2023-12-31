package hspscraper

import (
	"errors"
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/antchfx/xpath"
	"golang.org/x/net/html"
	"strings"
	"time"
)

var xPathCourseRowTemplate = "//table[@class='bs_kurse']/tbody//td[text() = '%s']/parent::tr"

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
var xPathCourseID = xpath.MustCompile("//td[@class='bs_sbuch']/input")

var hspUrlBase = "https://buchung.hochschulsport-hamburg.de"

type CourseState string

const (
	CourseStateOpen        = "Vormerkliste"
	CourseStateWaitingList = "Warteliste"
)

type Course struct {
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
	Updated  time.Time
}

func FindCourse(sport string, courseNumber string) (*Course, error) {
	doc, err := htmlquery.LoadURL(getHspSportUrl(sport))
	if err != nil {
		return nil, err
	}

	xPathCourseRow, err := xpath.Compile(fmt.Sprintf(xPathCourseRowTemplate, courseNumber))
	if err != nil {
		return nil, err
	}

	if tr := htmlquery.QuerySelector(doc, xPathCourseRow); tr != nil {
		return parseCourseRow(tr, true)
	}
	return nil, errors.New("course not found")
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
		course, err := parseCourseRow(tr, getDates)
		if err != nil {
			return nil, err
		}
		courses = append(courses, course)
	}

	return courses, nil
}

func parseCourseRow(tr *html.Node, getDates bool) (*Course, error) {
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
	var err error = nil
	if getDates {
		dates = make([]CourseDate, 0)
		n := htmlquery.QuerySelector(tr, xPathCourseDates)
		href := htmlquery.SelectAttr(n, "href")
		if dates, err = parseDates(href); err != nil {
			return nil, err
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
		id = htmlquery.SelectAttr(n, "name")
	}

	course := Course{
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
	return &course, nil
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
			Updated:  time.Now(),
		}
		courseDates = append(courseDates, courseDate)
	}
	return courseDates, nil
}
