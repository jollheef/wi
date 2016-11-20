/**
 * @file storage.go
 * @author Mikhail Klementyev jollheef<AT>riseup.net
 * @license GNU GPLv3
 * @date July, 2016
 * @brief Database functions
 */

package storage

import (
	"database/sql"
	"errors"
	"reflect"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func OpenDB(path string) (db *sql.DB, err error) {
	db, err = sql.Open("sqlite3", path)
	if err != nil {
		return
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS `links` " +
		"( `id` INTEGER PRIMARY KEY AUTOINCREMENT, `url` TEXT );")
	if err != nil {
		return
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS `history` " +
		"( `id` INTEGER PRIMARY KEY AUTOINCREMENT, `url` TEXT );")
	if err != nil {
		return
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS `fields` " +
		"( `id`      INTEGER PRIMARY KEY AUTOINCREMENT, " +
		"  `form_id` INTEGER, " +
		"  `hidden`  BOOLEAN, " +
		"  `value`   TEXT, " +
		"  `name`   TEXT );")
	if err != nil {
		return
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS `forms` " +
		"( `id`   INTEGER PRIMARY KEY AUTOINCREMENT, " +
		"  `post` BOOLEAN, " +
		"  `url`  TEXT );")

	return
}

type Field struct {
	Hidden bool
	Value  string
	Name   string
}

func getFields(db *sql.DB, formNo int64) (fields []Field, err error) {
	stmt, err := db.Prepare("SELECT `hidden`, `value`, `name` FROM `fields` WHERE form_id=$1;")
	if err != nil {
		return
	}
	defer stmt.Close()

	rows, err := stmt.Query(formNo)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var f Field
		err = rows.Scan(&f.Hidden, &f.Value, &f.Name)
		if err != nil {
			return
		}

		fields = append(fields, f)
	}

	return
}

func addField(db *sql.DB, name, value string, hidden bool, formNo int64) (err error) {
	stmt, err := db.Prepare("INSERT INTO `fields` " +
		"(`form_id`, `name`, `hidden`, `value`) VALUES ($1, $2, $3, $4);")
	if err != nil {
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(formNo, name, hidden, value)
	return
}

func AddForm(db *sql.DB, fields []Field, url, method string) (formNo int64, err error) {
	stmt, err := db.Prepare("INSERT INTO `forms` (`url`, `post`) VALUES ($1, $2);")
	if err != nil {
		return
	}
	defer stmt.Close()

	post := false // GET
	if strings.ToUpper(method) == "POST" {
		post = true
	}

	r, err := stmt.Exec(url, post)
	if err != nil {
		return
	}

	formNo, err = r.LastInsertId()
	if err != nil {
		return
	}

	for _, f := range fields {
		addField(db, f.Name, f.Value, f.Hidden, formNo)
	}

	return
}

func GetFormID(db *sql.DB, fields []Field, url, method string) (formNo int64, err error) {
	stmt, err := db.Prepare("SELECT `id` FROM `forms` WHERE url=$1 AND post=$2;")
	if err != nil {
		return
	}
	defer stmt.Close()

	post := false // GET
	if strings.ToUpper(method) == "POST" {
		post = true
	}

	err = stmt.QueryRow(url, post).Scan(&formNo)
	if err != nil {
		return
	}

	dbFields, err := getFields(db, formNo)
	if err != nil {
		return
	}

	if !reflect.DeepEqual(fields, dbFields) {
		err = errors.New("Fields not match")
		return
	}

	return
}

func GetForm(db *sql.DB, formID int64) (fields []Field, url string, post bool, err error) {
	stmt, err := db.Prepare("SELECT `post`, `url` FROM `forms` WHERE id=$1;")
	if err != nil {
		return
	}
	defer stmt.Close()

	err = stmt.QueryRow(formID).Scan(&post, &url)
	if err != nil {
		return
	}

	fields, err = getFields(db, formID)
	if err != nil {
		return
	}

	return
}

func AddLink(db *sql.DB, url string) (linkNo int64, err error) {
	stmt, err := db.Prepare("INSERT INTO `links` (`url`) VALUES ($1);")
	if err != nil {
		return
	}
	defer stmt.Close()

	r, err := stmt.Exec(url)
	if err != nil {
		return
	}

	linkNo, err = r.LastInsertId()

	return
}

func GetLink(db *sql.DB, linkID int64) (url string, err error) {
	stmt, err := db.Prepare("SELECT `url` FROM `links` WHERE id=$1;")
	if err != nil {
		return
	}
	defer stmt.Close()

	err = stmt.QueryRow(linkID).Scan(&url)

	return
}

func GetLinkID(db *sql.DB, url string) (linkID int64, err error) {
	stmt, err := db.Prepare("SELECT `id` FROM `links` WHERE url=$1;")
	if err != nil {
		return
	}
	defer stmt.Close()

	err = stmt.QueryRow(url).Scan(&linkID)

	return
}

func AddHistoryURL(db *sql.DB, url string) (err error) {
	stmt, err := db.Prepare("INSERT INTO `history` (`url`) VALUES ($1);")
	if err != nil {
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(url)

	return
}

type HistoryItem struct {
	ID  int64
	URL string
}

func GetHistory(db *sql.DB) (history []HistoryItem, err error) {
	rows, err := db.Query("SELECT `id`, `url` FROM `history`;")
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var h HistoryItem

		err = rows.Scan(&h.ID, &h.URL)
		if err != nil {
			return
		}

		history = append(history, h)
	}

	return
}

func GetHistoryUrl(db *sql.DB, historyID int64) (url string, err error) {
	stmt, err := db.Prepare("SELECT `url` FROM `history` WHERE id=$1;")
	if err != nil {
		return
	}
	defer stmt.Close()

	err = stmt.QueryRow(historyID).Scan(&url)

	return
}
