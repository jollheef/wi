/**
 * @file main.go
 * @author Mikhail Klementyev jollheef<AT>riseup.net
 * @license GNU GPLv3
 * @date July, 2016
 * @brief Tiny non-interactive cli browser
 */

package main

import (
	"github.com/jollheef/wi/commands"
	"github.com/jollheef/wi/storage"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	get    = kingpin.Command("get", "Get url")
	getUrl = get.Arg("url", "Url").Required().String()

	link   = kingpin.Command("link", "Get link")
	linkNo = link.Arg("no", "Number").Required().Int64()

	historyList      = kingpin.Command("history", "List history")
	historyListItems = historyList.Arg("items", "Amount of items").Int64()
	historyListAll   = historyList.Flag("all", "Show all items").Bool()
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
	case "link":
		commands.Link(db, *linkNo)
	case "history":
		commands.History(db, *historyListItems, 20, *historyListAll)
	}
}
