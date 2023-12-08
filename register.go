package hspscraper

import (
	"errors"
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/antchfx/xpath"
	"strings"
	"time"
)

const bookingUrl = "https://buchung.hochschulsport-hamburg.de/cgi/anmeldung.fcgi"

var (
	xPathSex              = xpath.MustCompile("//input[@name='sex' and @checked]")
	xPathVorname          = xpath.MustCompile("//input[@name='vorname']")
	xPathName             = xpath.MustCompile("//input[@name='name']")
	xPathStrasse          = xpath.MustCompile("//input[@name='strasse']")
	xPathOrt              = xpath.MustCompile("//input[@name='ort']")
	xPathEmail            = xpath.MustCompile("//input[@name='email']")
	xPathTelefon          = xpath.MustCompile("//input[@name='telefon']")
	xPathStatusorig       = xpath.MustCompile("//select[@name='statusorig']/option[@selected]")
	xPathTimeSlot         = xpath.MustCompile("//input[@value='buchen']")
	xPathFid              = xpath.MustCompile("//input[@name='fid']")
	xPathFormData         = xpath.MustCompile("//input[@name='_formdata']")
	xPathPreisAnz         = xpath.MustCompile("//input[@name='preis_anz']")
	xPathTnbed            = xpath.MustCompile("//input[@name='tnbed']")
	xPathBuchungsLink     = xpath.MustCompile("//div[@class='bs_meldung']//a[1]")
	xPathBuchungsErrorMsg = xpath.MustCompile("//div[@class='bs_meldung']/div[1]/text()")
	xPathConfirmation     = xpath.MustCompile("//div[@class='content']/div/span[1]/text()")
)

func Register(course *Course, email string, pw string, date time.Time) error {

	if course.id == "" {
		return errors.New("course.id cannot be empty")
	}
	if course.Sport == "" {
		return errors.New("course.Sport cannot be empty")
	}

	if course.State != CourseStateOpen {
		return errors.New("course not open")
	}

	if email == "" {
		return errors.New("email cannot be empty")
	}

	if pw == "" {
		return errors.New("password cannot be empty")
	}

	node, err := bookingRequestWithReferer(map[string]string{
		course.id: CourseStateOpen,
	}, getHspSportUrl(course.Sport))
	if err != nil {
		return err
	}

	fid := getValue(node, xPathFid)
	timeSlotKey := getAtrValue(node, xPathTimeSlot, "name")
	timeSlot := strings.TrimPrefix(timeSlotKey, "BS_Termin_")

	firstBookableTime, err := time.Parse(time.DateOnly, timeSlot)
	if err != nil {
		return err
	}

	if !firstBookableTime.Equal(date) {
		return errors.New("not the right date")
	}

	if fid == "" || timeSlotKey == "" {
		return errors.New("fid or time slot not found")
	}

	_, err = bookingRequest(map[string]string{
		"fid":       fid,
		timeSlotKey: "buchen",
	})
	if err != nil {
		return err
	}

	node, err = bookingRequest(map[string]string{
		"fid":           fid,
		"Termin":        timeSlot,
		"pw_email":      email,
		"pw_pwd_" + fid: pw,
	})

	regData := map[string]string{
		"fid":        fid,
		"Termin":     timeSlot,
		"vorname":    getValue(node, xPathVorname),
		"sex":        getValue(node, xPathSex),
		"name":       getValue(node, xPathName),
		"strasse":    getValue(node, xPathStrasse),
		"ort":        getValue(node, xPathOrt),
		"email":      getValue(node, xPathEmail),
		"telefon":    getValue(node, xPathTelefon),
		"statusorig": getValue(node, xPathStatusorig),
		"tnbed":      "1",
	}
	for k, v := range regData {
		if v == "" {
			err = errors.Join(err, errors.New(fmt.Sprintf("%s was empty", k)))
		}
	}
	if err != nil {
		err = errors.Join(err, errors.New("maybe email and password are wrong"))
		return err
	}
	node, err = bookingRequest(regData)
	if err != nil {
		return err
	}

	regData["Phase"] = "final"
	regData["_formdata"] = getValue(node, xPathFormData)
	regData["preis_anz"] = getValue(node, xPathPreisAnz)
	regData["tnbed"] = getValue(node, xPathTnbed)
	for k, v := range regData {
		if v == "" {
			err = errors.Join(err, errors.New(fmt.Sprintf("%s was empty", k)))
		}
	}
	if err != nil {
		return err
	}
	node, err = bookingRequest(regData)
	if err != nil {
		return err
	}

	bookgingErrorMsgNode := htmlquery.QuerySelector(node, xPathBuchungsErrorMsg)
	if bookgingErrorMsgNode == nil || bookgingErrorMsgNode.Data != "" {
		// Booking error
		err = errors.New("booking unsuccessful")
		bookingLink := getAtrValue(node, xPathBuchungsLink, "href")
		if bookingLink != "" {
			// Already Registered, bookingLink contains the confirmation link
			err = errors.Join(err, errors.New(fmt.Sprintf("already registered: %s", bookingLink)))
		}
		return err
	}

	confirmationNode := htmlquery.QuerySelector(node, xPathConfirmation)
	if confirmationNode == nil || confirmationNode.Data == "" {
		return errors.New("no confirmation found. If no email arrived, you are probably not registered")
	}

	//_ = html.Render(os.Stdout, node)

	return nil
}
