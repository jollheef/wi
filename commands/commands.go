/**
 * @file commands.go
 * @author Mikhail Klementyev jollheef<AT>riseup.net
 * @license GNU GPLv3
 * @date July, 2016
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

	"github.com/jollheef/wi/storage"

	"github.com/jaytaylor/html2text"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

func parseLink(db *sql.DB, oldPage, value string, req *http.Request) (htmlPage string, err error) {
	url, err := req.URL.Parse(value)
	if err != nil {
		return
	}

	linkNo, err := storage.GetLinkID(db, url.String())
	if err != nil {
		linkNo, err = storage.AddLink(db, url.String())
		if err != nil {
			return
		}
	}

	for _, s := range []string{value, html.EscapeString(value)} {
		htmlPage = strings.Replace(oldPage, "\""+s+"\"",
			"\""+fmt.Sprintf("%d", linkNo)+"\"", -1)
	}

	return
}

func parseLinks(db *sql.DB, body []byte, req *http.Request) (htmlPage string, err error) {
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
				htmlPage, err = parseLink(db, htmlPage, string(value), req)
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

func Get(db *sql.DB, url string) {
	client := &http.Client{}

	if !strings.Contains(url, "://") {
		url = "http://" + url
	}

	// TODO Full url encoding
	req, err := http.NewRequest("GET", strings.Replace(url, " ", "%20", -1), nil)
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Set("User-Agent", "Wi 0.0")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	storage.AddHistoryURL(db, url)

	defer resp.Body.Close()

	utf8, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		fmt.Println("Encoding error:", err)
		return
	}

	body, err := ioutil.ReadAll(utf8)
	if err != nil {
		fmt.Println("IO error:", err)
		return
	}

	htmlPage, err := parseLinks(db, body, req)
	if err != nil {
		panic(err)
	}

	text, err := html2text.FromString(htmlPage)
	if err != nil {
		panic(err)
	}
	text += ""

	fmt.Println(text)
}

func Link(db *sql.DB, linkID int64, fromHistory bool) {

	var url string
	var err error

	if fromHistory {
		url, err = storage.GetHistoryUrl(db, linkID)
	} else {
		url, err = storage.GetLink(db, linkID)
	}

	if err != nil {
		panic(err)
	}

	Get(db, url)
}

func History(db *sql.DB, argAmount, defaultAmount int64, all bool) {
	history, err := storage.GetHistory(db)
	if err != nil {
		panic(err)
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
