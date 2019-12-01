package handlers

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestExtractSummary(t *testing.T) {
	pagePrologue := "<html><head>"
	pageEiplogue := "</head><body></body></html>"
	pageURL := "http://test.com/test.html"
	cases := []struct {
		name            string
		hint            string
		html            string
		expectedSummary *PageSummary
	}{
		{
			"Open Graph Type",
			`Make sure you are reading the <meta property="og:type" content="..."> element`,
			pagePrologue + `<meta property="og:type" content="test type">` + pageEiplogue,
			&PageSummary{
				Type: "test type",
			},
		},
		{
			"Open Graph URL",
			`Make sure you are reading the <meta property="og:url" content="..."> element`,
			pagePrologue + `<meta property="og:url" content="http://test.com">` + pageEiplogue,
			&PageSummary{
				URL: "http://test.com",
			},
		},
		{
			"Open Graph Title",
			`Make sure you are reading the <meta property="og:title" content="..."> element`,
			pagePrologue + `<meta property="og:title" content="test title">` + pageEiplogue,
			&PageSummary{
				Title: "test title",
			},
		},
		{
			"Open Graph Site name",
			`Make sure you are reading the <meta property="og:site_name" content="..."> element`,
			pagePrologue + `<meta property="og:site_name" content="test site name">` + pageEiplogue,
			&PageSummary{
				SiteName: "test site name",
			},
		},
		{
			"Open Graph Description",
			`Make sure you are reading the <meta property="og:description" content="..."> element`,
			pagePrologue + `<meta property="og:description" content="test description">` + pageEiplogue,
			&PageSummary{
				Description: "test description",
			},
		},
		{
			"Open Graph Image",
			`Make sure you are reading the <meta property="og:image" content="..."> element`,
			pagePrologue + `<meta property="og:image" content="http://test.com/test.png">` + pageEiplogue,
			&PageSummary{
				Images: []*PreviewImage{
					{
						URL: "http://test.com/test.png",
					},
				},
			},
		},
		{
			"Open Graph Structured Image",
			`Make sure you are handling the image structured properties, as described in http://ogp.me/#structured`,
			pagePrologue + `
			<meta property="og:image" content="http://test.com/test.png">
			<meta property="og:image:secure_url" content="https://test.com/test.png">
			<meta property="og:image:type" content="image/png">
			<meta property="og:image:width" content="300">
			<meta property="og:image:height" content="300">
			<meta property="og:image:alt" content="test alt">
			` + pageEiplogue,
			&PageSummary{
				Images: []*PreviewImage{
					{
						URL:       "http://test.com/test.png",
						SecureURL: "https://test.com/test.png",
						Type:      "image/png",
						Width:     300,
						Height:    300,
						Alt:       "test alt",
					},
				},
			},
		},
		{
			"Open Graph Multiple Images",
			`Make sure you are handling multiple images, as described in http://ogp.me/#array`,
			pagePrologue + `
			<meta property="og:image" content="http://test.com/test1.png">
			<meta property="og:image:width" content="100">
			<meta property="og:image:height" content="100">
			<meta property="og:image:alt" content="test alt 1">
			<meta property="og:image" content="http://test.com/test2.png">
			<meta property="og:image" content="http://test.com/test3.png">
			<meta property="og:image:alt" content="test alt 3">
			` + pageEiplogue,
			&PageSummary{
				Images: []*PreviewImage{
					{
						URL:    "http://test.com/test1.png",
						Width:  100,
						Height: 100,
						Alt:    "test alt 1",
					},
					{
						URL: "http://test.com/test2.png",
					},
					{
						URL: "http://test.com/test3.png",
						Alt: "test alt 3",
					},
				},
			},
		},
		{
			"All Open Graph Props",
			"Make sure you are handling all of the Open Graph properties listed in the assignment",
			pagePrologue + `
			<meta property="og:type" content="test type">
			<meta property="og:url" content="http://test.com">
			<meta property="og:title" content="test title">
			<meta property="og:site_name" content="test site name">
			<meta property="og:description" content="test description">
			<meta property="og:image" content="http://test.com/test.png">
			` + pageEiplogue,
			&PageSummary{
				Type:        "test type",
				URL:         "http://test.com",
				Title:       "test title",
				SiteName:    "test site name",
				Description: "test description",
				Images: []*PreviewImage{
					{
						URL: "http://test.com/test.png",
					},
				},
			},
		},
		{
			"HTML Title",
			`Make sure you get the page title from the <title> element if not Open Graph title property is available`,
			pagePrologue + `<title>HTML Page Title</title>` + pageEiplogue,
			&PageSummary{
				Title: "HTML Page Title",
			},
		},
		{
			"HTML Description",
			`Make sure you get the page description from the <meta name="author" content="..."> tag if no Open Graph description is available`,
			pagePrologue + `<meta name="description" content="test description">` + pageEiplogue,
			&PageSummary{
				Description: "test description",
			},
		},
		{
			"HTML Author",
			`Make sure you get the page author from the <meta name="author" content="..."> tag`,
			pagePrologue + `<meta name="author" content="test author">` + pageEiplogue,
			&PageSummary{
				Author: "test author",
			},
		},
		{
			"HTML Keywords With Spaces",
			`Make sure you get the page keywords from the <meta name="keywords" content="..."> tag`,
			pagePrologue + `<meta name="keywords" content="one, two, three">` + pageEiplogue,
			&PageSummary{
				Keywords: []string{"one", "two", "three"},
			},
		},
		{
			"HTML Keywords With No Spaces",
			`Make sure you get the page keywords from the <meta name="keywords" content="..."> tag`,
			pagePrologue + `<meta name="keywords" content="one,two,three">` + pageEiplogue,
			&PageSummary{
				Keywords: []string{"one", "two", "three"},
			},
		},
		{
			"HTML Icon",
			`Make sure you get the page icon from the <link rel="icon" href="..."> tag`,
			pagePrologue + `<link rel="icon" href="http://test.com/test.png">` + pageEiplogue,
			&PageSummary{
				Icon: &PreviewImage{
					URL: "http://test.com/test.png",
				},
			},
		},
		{
			"HTML Icon With Sizes",
			`Make sure you parse the 'sizes' attribute to get the icon height and width`,
			pagePrologue + `<link rel="icon" href="http://test.com/test.png" sizes="100x200">` + pageEiplogue,
			&PageSummary{
				Icon: &PreviewImage{
					URL:    "http://test.com/test.png",
					Height: 100,
					Width:  200,
				},
			},
		},
		{
			"HTML Icon With Size Any",
			`The sizes attribute of the <link rel="icon"> tag may have the value "any" to indicate no size preference`,
			pagePrologue + `<link rel="icon" href="http://test.com/test.png" sizes="any">` + pageEiplogue,
			&PageSummary{
				Icon: &PreviewImage{
					URL: "http://test.com/test.png",
				},
			},
		},
		{
			"HTML Icon With Type",
			`Make sure you read the 'type' attribute to get the icon type`,
			pagePrologue + `<link rel="icon" href="http://test.com/test.png" type="image/png">` + pageEiplogue,
			&PageSummary{
				Icon: &PreviewImage{
					URL:  "http://test.com/test.png",
					Type: "image/png",
				},
			},
		},
		{
			"Self-Closing Meta",
			"Make sure you are handling self-closing <meta ... /> tags",
			pagePrologue + `<meta property="og:title" content="Open Graph Title"/>` + pageEiplogue,
			&PageSummary{
				Title: "Open Graph Title",
			},
		},
		{
			"Attribute Order",
			"HTML elements and attributes can be in any order; don't assume a particular order",
			pagePrologue + `
			<meta content="test title" property="og:title">
			<meta content="test type" property="og:type">
			<meta content="http://test.com/test.png" property="og:image">
			<meta content="test site name" property="og:site_name">
			<meta content="test description" property="og:description">
			<meta content="http://test.com" property="og:url">
			` + pageEiplogue,
			&PageSummary{
				Type:        "test type",
				URL:         "http://test.com",
				Title:       "test title",
				SiteName:    "test site name",
				Description: "test description",
				Images: []*PreviewImage{
					{
						URL: "http://test.com/test.png",
					},
				},
			},
		},
		{
			"HTML and Open Graph Title",
			`Make sure the <meta property="og:title"> overrides the HTML <title> element`,
			pagePrologue + `
			<meta property="og:title" content="Open Graph Title"/>
			<title>HTML Page Title</title>` + pageEiplogue,
			&PageSummary{
				Title: "Open Graph Title",
			},
		},
		{
			"HTML and Open Graph Description",
			`Make sure the <meta property="og:description"> overrides the HTML <meta name="description"> element`,
			pagePrologue + `
			<meta property="og:description" content="og description"/>
			<meta name="description" content="html description">` + pageEiplogue,
			&PageSummary{
				Description: "og description",
			},
		},
		{
			"Relative Image URL",
			"Remember to resolve relative image URLs to absolute ones using the page URL as a base",
			pagePrologue + `<meta property="og:image" content="/test.png"/>` + pageEiplogue,
			&PageSummary{
				Images: []*PreviewImage{
					{
						URL: "http://test.com/test.png",
					},
				},
			},
		},
		{
			"Relative Icon URL",
			"Remember to resolve relative HTML Icon URLs to absolute ones using the page URL as a base",
			pagePrologue + `<link rel="icon" href="/test.png"/>` + pageEiplogue,
			&PageSummary{
				Icon: &PreviewImage{
					URL: "http://test.com/test.png",
				},
			},
		},
		{
			"Empty Input",
			"A URL might return an empty page",
			"",
			&PageSummary{},
		},
	}

	for _, c := range cases {
		summary, err := extractSummary(pageURL, ioutil.NopCloser(strings.NewReader(c.html)))
		if err != nil && err != io.EOF {
			t.Errorf("case %s: unexpected error %v\nHINT: %s\n", c.name, err, c.hint)
		}
		if summary == nil {
			t.Errorf("case: %s: returned summary struct is nil", c.name)
			continue
		}
		if !reflect.DeepEqual(summary, c.expectedSummary) {
			//reflect.DeepEqual considers a non-nil empty slice to be different
			//than a nill slice, so check for those cases first
			if c.expectedSummary.Images == nil && summary.Images != nil {
				t.Errorf("case %s: expected nil `Images` slice, but got a non-nill slice", c.name)
			} else if c.expectedSummary.Keywords == nil && summary.Keywords != nil {
				t.Errorf("case %s: expected nil `Keywords` slice, but got a non-nill slice", c.name)
			} else if c.expectedSummary.Icon == nil && summary.Icon != nil {
				t.Errorf("case %s: expected nil `Icon` pointer, but got a non-nill pointer", c.name)
			} else {
				expectedJSON, _ := json.MarshalIndent(c.expectedSummary, "", "  ")
				actualJSON, _ := json.MarshalIndent(summary, "", "  ")
				t.Errorf("case %s: incorrect result:\nEXPECTED: %s\nACTUAL: %s\nHINT: %s\n",
					c.name, string(expectedJSON), string(actualJSON), c.hint)
			}
		}
	}
}

func TestFetchHTML(t *testing.T) {
	cases := []struct {
		name        string
		hint        string
		URL         string
		expectError bool
	}{
		{
			"Valid URL",
			"This is a valid HTML page, so this should work",
			"https://info344-a17.github.io/tests/ogall.html",
			false,
		},
		{
			"Not Found URL",
			"Remember to check the response status code",
			"https://info344-a17.github.io/tests/not-found.html",
			true,
		},
		{
			"Non-HTML URL",
			"Remember to check the response content-type to ensure it's an HTML page",
			"https://info344-a17.github.io/tests/test.png",
			true,
		},
	}

	for _, c := range cases {
		stream, err := fetchHTML(c.URL)

		if err != nil && !c.expectError {
			t.Errorf("case %s: unexpected error %v\nHINT: %s", c.name, err, c.hint)
		}
		if c.expectError && err == nil {
			t.Errorf("case %s: expected error but didn't get one\nHINT: %s", c.name, c.hint)
		}

		if stream != nil {
			stream.Close()
		}
	}
}

func TestSummaryHandler(t *testing.T) {
	//verify that response has
	// - correct response status code
	// - correct Content-Type header
	resp := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/summary?url=http://ogp.me", nil)
	SummaryHandler(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("incorrect response status code: expected %d but got %d", http.StatusOK, resp.Code)
	}
	expectedctype := "application/json"
	ctype := resp.Header().Get("Content-Type")
	if len(ctype) == 0 {
		t.Errorf("No `Content-Type` header found in the response: must be there start with `%s`", expectedctype)
	} else if !strings.HasPrefix(ctype, expectedctype) {
		t.Errorf("incorrect `Content-Type` header value: expected it to start with `%s` but got `%s`", expectedctype, ctype)
	}
}
