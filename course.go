package hspscraper

import (
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/antchfx/xpath"
	"strings"
)

const courseStateOpen = "Vormerkliste"
const courseStateWaitingList = "Warteliste"

var xPathCourseIDTemplate = "//a[@id='K%s']/following-sibling::input"

type Course struct {
	Sport        string
	CourseNumber string
	CourseOpen   bool
	courseID     string
}

func FindCourse(sport string, courseNumber string) (*Course, error) {
	doc, err := htmlquery.LoadURL(getHspSportUrl(sport))
	if err != nil {
		return nil, err
	}

	xPathCourseID, err := xpath.Compile(fmt.Sprintf(xPathCourseIDTemplate, courseNumber))
	if err != nil {
		return nil, err
	}
	courseState := getValue(doc, xPathCourseID)
	courseID := getAtrValue(doc, xPathCourseID, "name")

	return &Course{
		Sport:        sport,
		CourseNumber: courseNumber,
		CourseOpen:   courseState == courseStateOpen,
		courseID:     courseID,
	}, nil
}

func getHspSportUrl(sport string) string {
	return fmt.Sprintf("https://buchung.hochschulsport-hamburg.de/angebote/Wintersemester_2023_2024/_%s.html", strings.ReplaceAll(sport, " ", "_"))
}
