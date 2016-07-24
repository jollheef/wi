/**
 * @file storage.go
 * @author Mikhail Klementyev jollheef<AT>riseup.net
 * @license GNU GPLv3
 * @date July, 2016
 */

package storage

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func OpenDB(path string) (db *sql.DB, err error) {
	db, err = sql.Open("sqlite3", path)
	if err != nil {
		return
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS `links` " +
		"( `id` INTEGER PRIMARY KEY AUTOINCREMENT, `url` TEXT );")

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
