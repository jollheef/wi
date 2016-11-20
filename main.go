/**
 * @file main.go
 * @author Mikhail Klementyev jollheef<AT>riseup.net
 * @license GNU GPLv3
 * @date July, 2016
 * @brief Tiny non-interactive cli browser
 */

package main

import (
	"os"
	"strings"

	"github.com/jollheef/wi/commands"
	"github.com/jollheef/wi/storage"

	cookiejar "github.com/juju/persistent-cookiejar"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type searchList []string

func (l *searchList) Set(value string) (err error) {
	*l = append(*l, value)
	return
}

func (l *searchList) String() (s string) {
	return ""
}

func (l *searchList) IsCumulative() bool {
	return true
}

func SearchList(settings kingpin.Settings) (target *[]string) {
	target = new([]string)
	settings.SetValue((*searchList)(target))
	return
}

var (
	get    = kingpin.Command("get", "Get url")
	getUrl = get.Arg("url", "Url").Required().String()

	form     = kingpin.Command("form", "Fill form")
	formID   = form.Arg("id", "Form ID").Required().Int64()
	formArgs = SearchList(form.Arg("args", "Form arguments"))

	link            = kingpin.Command("link", "Get link")
	linkNo          = link.Arg("no", "Number").Required().Int64()
	linkFromHistory = link.Flag("history", "Item from history").Bool()

	historyList      = kingpin.Command("history", "List history")
	historyListItems = historyList.Arg("items", "Amount of items").Int64()
	historyListAll   = historyList.Flag("all", "Show all items").Bool()

	search     = kingpin.Command("search", "Search engine (google by default)")
	searchArgs = SearchList(search.Arg("string", "String for search"))
)

func main() {
	db, err := storage.OpenDB("/tmp/wi.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	os.Setenv("GOCOOKIES", "/tmp/wi.jar")

	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}

	defer jar.Save()

	switch kingpin.Parse() {
	case "get":
		commands.Get(db, jar, *getUrl)
	case "form":
		commands.Form(db, jar, *formID, *formArgs)
	case "link":
		commands.Link(db, jar, *linkNo, *linkFromHistory)
	case "history":
		commands.History(db, *historyListItems, 20, *historyListAll)
	case "search":
		// FIXME: currenlty supports only Google
		commands.Get(db, jar, "https://google.com/search?q="+strings.Join(*searchArgs, "+"))
	}
}
