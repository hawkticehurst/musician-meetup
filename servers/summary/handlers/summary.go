package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// PreviewImage represents a preview image for a page
type PreviewImage struct {
	URL       string `json:"url,omitempty"`
	SecureURL string `json:"secureURL,omitempty"`
	Type      string `json:"type,omitempty"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
	Alt       string `json:"alt,omitempty"`
}

// PageSummary represents summary properties for a web page
type PageSummary struct {
	Type        string          `json:"type,omitempty"`
	URL         string          `json:"url,omitempty"`
	Title       string          `json:"title,omitempty"`
	SiteName    string          `json:"siteName,omitempty"`
	Description string          `json:"description,omitempty"`
	Author      string          `json:"author,omitempty"`
	Keywords    []string        `json:"keywords,omitempty"`
	Icon        *PreviewImage   `json:"icon,omitempty"`
	Images      []*PreviewImage `json:"images,omitempty"`
}

// SummaryHandler handles requests for the page summary API.
// This API expects one query string parameter named `url`,
// which should contain a URL to a web page. It responds with
// a JSON-encoded PageSummary struct containing the page summary
// meta-data.
func SummaryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	queryValues := r.URL.Query()
	URL := queryValues.Get("url")
	if URL == "" {
		http.Error(w, "Error: Please pass a URL with request.", http.StatusBadRequest)
		return
	}

	htmlStream, err := fetchHTML(URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pageSummary, err := extractSummary(URL, htmlStream)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer htmlStream.Close()

	summaryJSON, err := json.Marshal(pageSummary)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(summaryJSON)
}

// fetchHTML fetches `pageURL` and returns the body stream or an error.
// Errors are returned if the response status code is an error (>=400),
// or if the content type indicates the URL is not an HTML page.
func fetchHTML(pageURL string) (io.ReadCloser, error) {
	res, err := http.Get(pageURL)
	if err != nil {
		return nil, fmt.Errorf("Error fetching given URL: %v", err)
	}

	if res.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("Given URL responded with status code: %d", res.StatusCode)
	}

	ctype := res.Header.Get("Content-Type")
	if !strings.HasPrefix(ctype, "text/html") {
		return nil, fmt.Errorf("Given URL response content type '%s' was not 'text/html'", ctype)
	}

	return res.Body, nil
}

// extractSummary tokenizes the `htmlStream` and populates a PageSummary
// struct with the page's summary meta-data.
func extractSummary(pageURL string, htmlStream io.ReadCloser) (*PageSummary, error) {
	tokenizer := html.NewTokenizer(htmlStream)
	summary := &PageSummary{}
	currImgIndex := -1

	for {
		tokenType := tokenizer.Next()

		// -------- Check for Error Tokens --------
		// If an error token is encountered, the end of the file was reached or
		// the HTML was malformed
		if tokenType == html.ErrorToken {
			err := tokenizer.Err()
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("Error tokenizing HTML: %v", err)
		}

		// -------- Check for Start & Self Closing Tokens --------
		// Process start and self closing tokens. Checking for the various meta tags
		// that should be retrieved and add them to the summary struct if found.
		// In this case the only tags we are looking for are: meta, title, and link
		if tokenType == html.StartTagToken || tokenType == html.SelfClosingTagToken {
			token := tokenizer.Token()
			if token.Data == "meta" {
				metaTagType := ""
				metaTagValue := ""

				for _, attr := range token.Attr {
					if attr.Key == "property" {
						metaTagType = attr.Val
					}
					if attr.Key == "name" {
						metaTagType = attr.Val
					}
					if attr.Key == "content" {
						metaTagValue = attr.Val
					}
				}

				switch metaTagType {
				case "og:type":
					summary.Type = metaTagValue
				case "og:url":
					summary.URL = metaTagValue
				case "og:title":
					summary.Title = metaTagValue
				case "og:site_name":
					summary.SiteName = metaTagValue
				case "og:description":
					summary.Description = metaTagValue
				case "description":
					if summary.Description == "" {
						summary.Description = metaTagValue
					}
				case "author":
					summary.Author = metaTagValue
				case "keywords":
					keywords := strings.Split(metaTagValue, ",")
					for i := range keywords {
						keywords[i] = strings.TrimSpace(keywords[i])
					}
					summary.Keywords = keywords
				case "og:image":
					image := &PreviewImage{}
					summary.Images = append(summary.Images, image)
					image.URL = getAbsoluteURL(metaTagValue, pageURL)
					currImgIndex++
				case "og:image:secure_url":
					image := summary.Images[currImgIndex]
					image.SecureURL = metaTagValue
				case "og:image:type":
					image := summary.Images[currImgIndex]
					image.Type = metaTagValue
				case "og:image:width":
					image := summary.Images[currImgIndex]
					num, _ := strconv.Atoi(metaTagValue)
					image.Width = num
				case "og:image:height":
					image := summary.Images[currImgIndex]
					num, _ := strconv.Atoi(metaTagValue)
					image.Height = num
				case "og:image:alt":
					image := summary.Images[currImgIndex]
					image.Alt = metaTagValue
				}
			} else if token.Data == "title" {
				tokenType = tokenizer.Next()
				if tokenType == html.TextToken {
					if summary.Title == "" {
						summary.Title = tokenizer.Token().Data
					}
				}
			} else if token.Data == "link" {
				for _, attr := range token.Attr {
					if attr.Key == "rel" && (attr.Val == "icon" || attr.Val == "shortcut icon") {
						icon := &PreviewImage{}
						for _, attr := range token.Attr {
							switch attr.Key {
							case "href":
								icon.URL = getAbsoluteURL(attr.Val, pageURL)
							case "type":
								icon.Type = attr.Val
							case "sizes":
								if attr.Val != "any" {
									if strings.Contains(attr.Val, "x") {
										iconDimensions := strings.Split(attr.Val, "x")
										iconHeight, _ := strconv.Atoi(iconDimensions[0])
										iconWidth, _ := strconv.Atoi(iconDimensions[1])
										icon.Height = iconHeight
										icon.Width = iconWidth
									} else {
										iconDimensions := strings.Split(attr.Val, "X")
										iconHeight, _ := strconv.Atoi(iconDimensions[0])
										iconWidth, _ := strconv.Atoi(iconDimensions[1])
										icon.Height = iconHeight
										icon.Width = iconWidth
									}
								}
							case "alt":
								icon.Alt = attr.Val
							}
						}
						summary.Icon = icon
					}
				}
			}
		}

		// -------- Check for End Tokens --------
		// Process end tokens. Check if we've reached the closing
		// head tag, marking the end of the meta data we can retrieve
		if tokenType == html.EndTagToken {
			token := tokenizer.Token()
			if token.Data == "head" {
				break
			}
		}
	}

	return summary, nil
}

// getAbsoluteURL helper function that takes a reference (often relative) URI
// and an absolute URL of the page where the reference URI originates and
// resolve the reference URI into an absolute URL
func getAbsoluteURL(refURI string, pageURL string) string {
	referenceURI, _ := url.Parse(refURI)
	baseURI, _ := url.Parse(pageURL)

	return baseURI.ResolveReference(referenceURI).String()
}
