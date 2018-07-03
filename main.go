package atmovies

import (
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	baseSite = "http://search.atmovies.com.tw"
)

type MovieDetail struct {
	URL     string
	Title   string
	Poster  string
	Desc    string
	Runtime string
	OnDate  string
}

func getMovieURL(movieName string) (string, bool) {
	client := &http.Client{}
	qstr := strings.NewReader("fr=search-page&enc=UTF-8&type=all&search_term=" + movieName)
	req, err := http.NewRequest("POST", baseSite+"/search/", qstr)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", baseSite)
	res, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	s := doc.Find("blockquote header").First()
	return s.Find("a").Attr("href")
}
func getMovieDetail(movieURL string) MovieDetail {
	res, err := http.Get(baseSite + movieURL)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	title := doc.Find("div.filmTitle").First().Text()
	poster, _ := doc.Find("#filmTagBlock .image.Poster img").First().Attr("src")
	runtime := doc.Find("#filmTagBlock span ul.runtime").First().Text()
	desc := doc.Find("#filmTagBlock span~span").Children().Remove().End().Text()

	var validRuntime = regexp.MustCompile(`片長：(\d*)分 上映日期：(\d\d\d\d/\d\d/\d\d)`)
	rtResult := validRuntime.FindStringSubmatch(runtime)

	detail := MovieDetail{
		URL:     movieURL,
		Title:   strings.TrimSpace(title),
		Poster:  poster,
		Desc:    strings.TrimSpace(desc),
		Runtime: rtResult[1],
		OnDate:  rtResult[2],
	}
	return detail
}
func parseMovieName(name string) string {
	var validName = regexp.MustCompile(`^(.*)\.20\d\d`)

	result := validName.FindStringSubmatch(name)
	if len(result) < 1 {
		return name
	}
	return result[1]
}

func Query(str string) (*MovieDetail, bool) {
	name := parseMovieName(str)
	movieURL, exist := getMovieURL(name)
	if !exist {
		return nil, false
	}
	detail := getMovieDetail(movieURL)

	return &detail, true
}
