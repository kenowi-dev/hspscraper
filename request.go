package hspscraper

import (
	"errors"
	"github.com/antchfx/htmlquery"
	"github.com/antchfx/xpath"
	"golang.org/x/net/html"
	"net/http"
	"net/url"
	"strings"
)

func bookingRequest(formData map[string]string) (*html.Node, error) {
	return bookingRequestWithReferer(formData, bookingUrl)
}

func bookingRequestWithReferer(bookingData map[string]string, referer string) (node *html.Node, err error) {
	formData := make(url.Values, len(bookingData))
	for k, v := range bookingData {
		formData.Set(k, v)
	}
	request, err := http.NewRequest(http.MethodPost, bookingUrl, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}

	/*
		out, _ := httputil.DumpRequestOut(request, true)
		fmt.Printf("%s\n\n", out)
	*/
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Referer", referer)

	form, e := http.DefaultClient.Do(request)
	if e != nil {
		return nil, e
	}
	defer func() {
		if c := form.Body.Close(); err == nil {
			err = c
		}
	}()

	if form.StatusCode == 302 {
		location := form.Header.Get("Location")
		if location == "" {
			return nil, errors.New("redirect with no location")
		}
		redirectRequest, err := http.NewRequest(http.MethodGet, location, nil)
		if err != nil {
			return nil, err
		}
		request.Header.Set("Referer", referer)
		form, e = http.DefaultClient.Do(redirectRequest)
		if e != nil {
			return nil, e
		}
		defer func() {
			if c := form.Body.Close(); err == nil {
				err = c
			}
		}()
	}

	n, e := html.Parse(form.Body)
	if e != nil {
		return nil, e
	}
	return n, nil
}

func getValue(node *html.Node, expr *xpath.Expr) string {
	return getAtrValue(node, expr, "value")
}

func getAtrValue(node *html.Node, expr *xpath.Expr, atr string) string {
	return htmlquery.SelectAttr(htmlquery.QuerySelector(node, expr), atr)
}
