package hspscraper

import (
	"errors"
	"github.com/antchfx/xpath"
	"strings"
)

const bookingUrl = "https://buchung.hochschulsport-hamburg.de/cgi/anmeldung.fcgi"

var (
	xPathSex        = xpath.MustCompile("//input[@name='sex' and @checked]")
	xPathVorname    = xpath.MustCompile("//input[@name='vorname']")
	xPathName       = xpath.MustCompile("//input[@name='name']")
	xPathStrasse    = xpath.MustCompile("//input[@name='strasse']")
	xPathOrt        = xpath.MustCompile("//input[@name='ort']")
	xPathEmail      = xpath.MustCompile("//input[@name='email']")
	xPathTelefon    = xpath.MustCompile("//input[@name='telefon']")
	xPathStatusorig = xpath.MustCompile("//select[@name='statusorig']/option[@selected]")
	xPathTimeSlot   = xpath.MustCompile("//input[@value='buchen']")
	xPathFid        = xpath.MustCompile("//input[@name='fid']")
	xPathFormData   = xpath.MustCompile("//input[@name='_formdata']")
	xPathPreisAnz   = xpath.MustCompile("//input[@name='preis_anz']")
	xPathTnbed      = xpath.MustCompile("//input[@name='tnbed']")
)

func Register(course *Course, email string, pw string) error {

	if !course.CourseOpen {
		return errors.New("course not open")
	}

	node, err := bookingRequestWithReferer(map[string]string{
		course.courseID: courseStateOpen,
	}, getHspSportUrl(course.Sport))
	if err != nil {
		return err
	}

	fid := getValue(node, xPathFid)
	timeSlotKey := getAtrValue(node, xPathTimeSlot, "name")
	timeSlot := strings.TrimPrefix(timeSlotKey, "BS_Termin_")

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
		"sex":        getValue(node, xPathSex),
		"vorname":    getValue(node, xPathVorname),
		"name":       getValue(node, xPathName),
		"strasse":    getValue(node, xPathStrasse),
		"ort":        getValue(node, xPathOrt),
		"email":      getValue(node, xPathEmail),
		"telefon":    getValue(node, xPathTelefon),
		"statusorig": getValue(node, xPathStatusorig),
		"tnbed":      "1",
	}
	node, err = bookingRequest(regData)
	if err != nil {
		return err
	}

	regData["Phase"] = "final"
	regData["_formdata"] = getValue(node, xPathFormData)
	regData["preis_anz"] = getValue(node, xPathPreisAnz)
	regData["tnbed"] = getValue(node, xPathTnbed)
	node, err = bookingRequest(regData)
	if err != nil {
		return err
	}

	// TODO check last response, weather the registration was successful

	return nil
}
