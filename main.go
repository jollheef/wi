/**
 * @file main.go
 * @author Mikhail Klementyev jollheef<AT>riseup.net
 * @license GNU GPLv3
 * @date July, 2016
 * @brief Tiny non-interactive cli browser
 */

package main

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
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	get    = kingpin.Command("get", "Get url")
	getUrl = get.Arg("url", "Url").Required().String()

	link   = kingpin.Command("link", "Get link")
	linkNo = link.Arg("no", "Number").Required().Int64()

	historyList      = kingpin.Command("history", "List history")
	historyListItems = historyList.Arg("items", "Amount of items").Int64()
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

func cmd_url(db *sql.DB, url string) {
	client := &http.Client{}

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

func cmd_link(db *sql.DB, linkID int64) {
	url, err := storage.GetLink(db, linkID)
	if err != nil {
		panic(err)
	}

	cmd_url(db, url)
}

func main() {
	db, err := storage.OpenDB("/tmp/wi.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	switch kingpin.Parse() {
	case "get":
		cmd_url(db, *getUrl)
	case "link":
		cmd_link(db, *linkNo)
	case "history":
		fmt.Println("not implemented")
	}
}
