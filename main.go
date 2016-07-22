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

	"./storage"

	"github.com/jaytaylor/html2text"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	arg_url  = kingpin.Flag("url", "Url").String()
	arg_link = kingpin.Flag("link", "Link").Int()
)

func cmd_url(db *sql.DB, url string) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
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

	htmlPage := string(body)

	z := html.NewTokenizer(bytes.NewReader(body))

	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		}

		for {
			key, value, moreAttr := z.TagAttr()

			if string(key) == "href" {

				url, err := req.URL.Parse(string(value))
				if err != nil {
					panic(err)
				}

				linkNo, err := storage.AddLink(db, url.String())
				if err != nil {
					panic(err)
				}

				for _, s := range []string{string(value), html.EscapeString(string(value))} {
					htmlPage = strings.Replace(htmlPage, "\""+s+"\"",
						"\""+fmt.Sprintf("%d", linkNo)+"\"", -1)
				}
			}

			if !moreAttr {
				break
			}
		}
	}

	text, err := html2text.FromString(htmlPage)
	if err != nil {
		panic(err)
	}
	text += ""

	fmt.Println(text)
}

func cmd_link(db *sql.DB, linkID int) {
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

	kingpin.Parse()

	if *arg_url != "" {
		cmd_url(db, *arg_url)
	} else if *arg_link != 0 {
		cmd_link(db, *arg_link)
	}
}
