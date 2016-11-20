/**
 * @file commands.go
 * @author Mikhail Klementyev jollheef<AT>riseup.net
 * @license GNU GPLv3
 * @date July, 2016
 * @brief Command line options ('wi (get|link|...)')
 */

package commands

import (
	"database/sql"
	"strings"

	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/jollheef/wi/storage"

	"github.com/PuerkitoBio/goquery"
	"github.com/jaytaylor/html2text"
	cookiejar "github.com/juju/persistent-cookiejar"
	"golang.org/x/net/html/charset"
)

func fixLinks(db *sql.DB, doc *goquery.Document, pageUrl *url.URL) (err error) {

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		url, exists := s.Attr("href")
		if !exists {
			return
		}

		linkUrl, err := pageUrl.Parse(url)
		if err != nil {
			return
		}

		linkNo, err := storage.GetLinkID(db, linkUrl.String())
		if err != nil {
			linkNo, err = storage.AddLink(db, linkUrl.String())
			if err != nil {
				log.Fatalln("Add link:", err)
			}
		}

		s.SetAttr("href", fmt.Sprintf("%d", linkNo))
	})

	return
}

func fixForms(db *sql.DB, doc *goquery.Document, pageUrl *url.URL) (err error) {

	doc.Find("form").Each(func(i int, s *goquery.Selection) {
		var fields []storage.Field
		s.Find("input").Map(
			func(i int, s *goquery.Selection) (str string) {
				f := storage.Field{}
				var exists bool
				f.Name, exists = s.Attr("name")
				if !exists {
					return
				}
				f.Value, _ = s.Attr("value")
				hidden, _ := s.Attr("type")
				if hidden == "hidden" {
					f.Hidden = true
				}
				fields = append(fields, f)
				return
			})

		action, _ := s.Attr("action")

		actionUrl, err := pageUrl.Parse(action)
		if err != nil {
			return
		}

		method, _ := s.Attr("method")

		formNo, err := storage.GetFormID(db, fields, actionUrl.String(), method)
		if err != nil {
			formNo, err = storage.AddForm(db, fields, actionUrl.String(), method)
			if err != nil {
				log.Fatalln(err)
			}
		}

		s.AppendHtml(fmt.Sprintf("(%d %s)", formNo, strings.ToUpper(method)))
	})

	return
}

func Get(db *sql.DB, jar *cookiejar.Jar, linkUrl string) {
	client := &http.Client{Jar: jar}

	var lastUrl *url.URL

	client.CheckRedirect = func(r *http.Request, via []*http.Request) (err error) {
		lastUrl = r.URL
		return
	}

	if !strings.Contains(linkUrl, "://") {
		linkUrl = "https://" + linkUrl
	}

	u, err := url.Parse(linkUrl)
	if err != nil {
		log.Fatalln(err)
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Set("User-Agent", "Wi 0.0")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	if lastUrl == nil {
		lastUrl = req.URL
	}

	storage.AddHistoryURL(db, linkUrl)

	defer resp.Body.Close()

	utf8, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		log.Fatalln("Encoding error:", err)
	}

	doc, err := goquery.NewDocumentFromReader(utf8)
	if err != nil {
		log.Fatalln("Create document error:", err)
	}

	err = fixLinks(db, doc, lastUrl)
	if err != nil {
		log.Fatalln("Fix links error:", err)
	}

	err = fixForms(db, doc, lastUrl)
	if err != nil {
		log.Fatalln("Fix forms error", err)
	}

	htmlPage, err := doc.Html()
	if err != nil {
		log.Fatalln("Convert to html error:", err)
	}

	text, err := html2text.FromString(htmlPage)
	if err != nil {
		log.Fatalln("Html to text error:", err)
	}
	text += ""

	fmt.Println(text)
}

func Form(db *sql.DB, jar *cookiejar.Jar, formID int64, formArgs []string) {
	fields, formUrl, post, err := storage.GetForm(db, formID)
	if err != nil {
		log.Fatalln("Get form:", err)
	}

	if len(formArgs) == 0 {
		if post {
			fmt.Print("POST ")
		}

		fmt.Println(formUrl)

		fmt.Print("Values: ")
		for i, f := range fields {
			if i != 0 {
				fmt.Print("\n\t")
			}
			fmt.Printf(`%s="%s"`, f.Name, f.Value)
		}
		fmt.Println()

		return
	}

	urlData := url.Values{}
	for _, f := range fields {
		urlData.Set(f.Name, f.Value)
	}

	for _, fa := range formArgs {
		nameAndValue := strings.Split(fa, "=")
		if len(nameAndValue) != 2 {
			continue
		}
		name := nameAndValue[0]
		value := nameAndValue[1]
		urlData.Set(name, value)
	}

	client := &http.Client{Jar: jar}

	var lastUrl *url.URL

	client.CheckRedirect = func(r *http.Request, via []*http.Request) (err error) {
		lastUrl = r.URL
		return
	}

	resp, err := client.PostForm(formUrl, urlData)
	if err != nil {
		fmt.Println(err)
	}

	if lastUrl == nil {
		lastUrl, _ = resp.Location()
	}

	log.Println(resp.Status)

	var status int64
	fmt.Sscanf(resp.Status, "%d", &status)

	if status >= 300 && status < 400 {
		Get(db, jar, lastUrl.String())
	}
}

func Link(db *sql.DB, jar *cookiejar.Jar, linkID int64, fromHistory bool) {

	var linkUrl string
	var err error

	if fromHistory {
		linkUrl, err = storage.GetHistoryUrl(db, linkID)
	} else {
		linkUrl, err = storage.GetLink(db, linkID)
	}

	if err != nil {
		log.Fatalln("Get link/history url error:", err)
	}

	Get(db, jar, linkUrl)
}

func History(db *sql.DB, argAmount, defaultAmount int64, all bool) {
	history, err := storage.GetHistory(db)
	if err != nil {
		log.Fatalln("Get history error:", err)
	}

	var amount int64

	if all {
		amount = int64(len(history))
	} else if argAmount == 0 {
		if int64(len(history)) < defaultAmount {
			amount = int64(len(history))
		} else {
			amount = defaultAmount
		}
	} else {
		if amount > int64(len(history)) {
			amount = int64(len(history))
		} else {
			amount = argAmount
		}
	}

	for _, h := range history[int64(len(history))-amount:] {
		fmt.Println(h.ID, h.URL)
	}
}
