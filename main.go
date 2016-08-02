/**
 * @file main.go
 * @author Mikhail Klementyev jollheef<AT>riseup.net
 * @license GNU GPLv3
 * @date July, 2016
 * @brief Tiny non-interactive cli browser
 */

package main

import (
	"strings"

	"github.com/jollheef/wi/commands"
	"github.com/jollheef/wi/storage"

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

	post     = kingpin.Command("post", "Fill post form")
	postID   = post.Arg("id", "Form ID").Required().Int64()
	postArgs = SearchList(post.Arg("args", "Post form arguments"))

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

	switch kingpin.Parse() {
	case "get":
		commands.Get(db, *getUrl)
	case "post":
		commands.Post(db, *postID, *postArgs)
	case "link":
		commands.Link(db, *linkNo, *linkFromHistory)
	case "history":
		commands.History(db, *historyListItems, 20, *historyListAll)
	case "search":
		// FIXME: currenlty supports only Google
		commands.Get(db, "https://google.com/search?q="+strings.Join(*searchArgs, "+"))
	}
}
