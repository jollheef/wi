/**
 * @file main.go
 * @author Mikhail Klementyev jollheef<AT>riseup.net
 * @license GNU GPLv3
 * @date July, 2016
 * @brief Tiny non-interactive cli browser
 */

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/jaytaylor/html2text"
	"golang.org/x/net/html/charset"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	url = kingpin.Flag("url", "Url").String()
)

func main() {

	kingpin.Parse()

	client := &http.Client{}

	req, err := http.NewRequest("GET", *url, nil)
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

	text, err := html2text.FromString(string(body))
	if err != nil {
		panic(err)
	}

	fmt.Println(text)
}
