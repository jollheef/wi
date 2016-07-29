/**
 * @file commands.go
 * @author Mikhail Klementyev jollheef<AT>riseup.net
 * @license GNU GPLv3
 * @date July, 2016
 * @brief Command line options ('wi (get|link|...)')
 */

package commands

import (
	"bytes"
	"database/sql"
	"strings"

	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/jollheef/wi/storage"

	"github.com/jaytaylor/html2text"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

func parseLink(db *sql.DB, oldPage, value string, lastUrl *url.URL) (htmlPage string, err error) {
	linkUrl, err := lastUrl.Parse(value)
	if err != nil {
		return
	}

	linkNo, err := storage.GetLinkID(db, linkUrl.String())
	if err != nil {
		linkNo, err = storage.AddLink(db, linkUrl.String())
		if err != nil {
			return
		}
	}

	htmlPage = oldPage

	for _, s := range []string{value, html.EscapeString(value)} {
		htmlPage = strings.Replace(htmlPage, "\""+s+"\"",
			"\""+fmt.Sprintf("%d", linkNo)+"\"", -1)
	}

	return
}

func parseLinks(db *sql.DB, body []byte, lastUrl *url.URL) (htmlPage string, err error) {
	htmlPage = string(body)

	z := html.NewTokenizer(bytes.NewReader(body))

	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		}

		for {
			key, value, moreAttr := z.TagAttr()

			if string(key) == "href" {
				htmlPage, err = parseLink(db, htmlPage, string(value), lastUrl)
				if err != nil {
					return
				}
			}

			if !moreAttr {
				break
			}
		}
	}

	return
}

func Get(db *sql.DB, linkUrl string) {
	client := &http.Client{}

	var lastUrl *url.URL

	client.CheckRedirect = func(r *http.Request, via []*http.Request) (err error) {
		lastUrl = r.URL
		return
	}

	if !strings.Contains(linkUrl, "://") {
		linkUrl = "https://" + linkUrl
	}

	// TODO Full url encoding
	req, err := http.NewRequest("GET", strings.Replace(linkUrl, " ", "%20", -1), nil)
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

	body, err := ioutil.ReadAll(utf8)
	if err != nil {
		log.Fatalln("IO error:", err)
	}

	htmlPage, err := parseLinks(db, body, lastUrl)
	if err != nil {
		log.Fatalln("Parse links error:", err)
	}

	text, err := html2text.FromString(htmlPage)
	if err != nil {
		log.Fatalln("Html to text error:", err)
	}
	text += ""

	fmt.Println(text)
}

func Link(db *sql.DB, linkID int64, fromHistory bool) {

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

	Get(db, linkUrl)
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
