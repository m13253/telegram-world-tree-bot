/*
	Telegram WorldTreeBot
	Copyright (C) 2017 StarBrilliant <m13253@hotmail.com>

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published
	by the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func createTables(db *sql.DB) (err error) {
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS match (a INTEGER UNIQUE, b INTEGER UNIQUE)")
	if err != nil {
		return
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS choosing (user INTEGER UNIQUE)")
	if err != nil {
		return
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS queue (user INTEGER UNIQUE, topic STRING UNIQUE)")
	return
}

func queryUser(db *sql.DB, user_a int64) (user_b int64, err error) {
	err = db.QueryRow("SELECT b FROM match WHERE a = ?", user_a).Scan(&user_b)
	if err == sql.ErrNoRows {
		return 0, nil
	} else if err != nil {
		return 0, err
	} else {
		return
	}
}

func connectUser(db *sql.DB, user_a int64, user_b int64) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return
	}

	_, err = tx.Exec("INSERT OR REPLACE INTO match VALUES (?, ?)", user_a, user_b)
	if err != nil {
		tx.Rollback()
		return
	}

	_, err = tx.Exec("INSERT OR REPLACE INTO match VALUES (?, ?)", user_b, user_a)
	if err != nil {
		tx.Rollback()
		return
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
	}

	return
}

func disconnectUser(db *sql.DB, user int64) (err error) {
	_, err = db.Exec("DELETE FROM match WHERE a = ? OR b = ?", user, user)
	return
}

func listTopics(db *sql.DB) (topics []string, err error) {
	rows, err := db.Query("SELECT topic FROM queue ORDER BY random()")
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var topic string
		err = rows.Scan(&topic)
		if err != nil {
			return
		}
		topics = append(topics, topic)
	}
	err = rows.Err()
	return
}

func pushTopic(db *sql.DB, user int64, topic string) (err error) {
	_, err = db.Exec("INSERT OR REPLACE INTO queue VALUES (?, ?)", user, topic)
	return
}

func popTopic(db *sql.DB, topic string) (user int64, err error) {
	err = db.QueryRow("SELECT user FROM queue WHERE topic = ?", topic).Scan(&user)
	if err == sql.ErrNoRows {
		return 0, nil
	} else if err != nil {
		return 0, err
	}
	err = cancelTopic(db, user)
	return
}

func cancelTopic(db *sql.DB, user int64) (err error) {
	_, err = db.Exec("DELETE FROM queue WHERE user = ?", user)
	return
}

func isUserInQueue(db *sql.DB, user int64) (ok bool, err error) {
	var user_ int64
	err = db.QueryRow("SELECT user FROM queue WHERE user = ?", user).Scan(&user_)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	} else {
		return true, nil
	}
}

func setChoosingStatus(db *sql.DB, user int64, choosing bool) (err error) {
	if choosing {
		_, err = db.Exec("INSERT OR REPLACE INTO choosing VALUES (?)", user)
	} else {
		_, err = db.Exec("DELETE FROM choosing WHERE user = ?", user)
	}
	return
}

func getChoosingStatus(db *sql.DB, user int64) (choosing bool, err error) {
	var user_ int64
	err = db.QueryRow("SELECT user FROM choosing WHERE user = ?", user).Scan(&user_)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err == nil {
		return true, nil
	} else {
		return false, err
	}
}

