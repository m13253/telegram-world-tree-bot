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
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS match (a INTEGER UNIQUE, b INTEGER)")
	if err != nil {
		return
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS lobby (user INTEGER UNIQUE, topic STRING)")
	return
}

func getActiveUsers(db *sql.DB) (chat int, lobby int, err error) {
	err = db.QueryRow("SELECT count(*) FROM match").Scan(&chat)
	if err != nil {
		return
	}
	err = db.QueryRow("SELECT count(*) FROM lobby").Scan(&lobby)
	return
}

func queryMatch(db *sql.DB, user_a int64) (user_b int64, err error) {
	err = db.QueryRow("SELECT b FROM match WHERE a = ? LIMIT 1", user_a).Scan(&user_b)
	return
}

func connectChat(db *sql.DB, user_a int64, user_b int64) (err error) {
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

func disconnectChat(db *sql.DB, user_a int64, user_b int64) (err error) {
	if user_b != 0 {
		_, err = db.Exec("UPDATE match SET b = 0 WHERE a = ?", user_b)
	}
	_, err1 := db.Exec("DELETE FROM match WHERE a = ?", user_a)
	if err != nil {
		return
	}
	return err1
}

func listTopics(db *sql.DB) (topics []string, err error) {
	rows, err := db.Query("SELECT topic FROM lobby WHERE topic IS NOT NULL ORDER BY random()")
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

func listPendingUsers(db *sql.DB) (users []int64, err error) {
	rows, err := db.Query("SELECT user FROM lobby")
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var user int64
		err = rows.Scan(&user)
		if err != nil {
			return
		}
		users = append(users, user)
	}
	err = rows.Err()
	return
}

func setTopic(db *sql.DB, user int64, topic string) (err error) {
	_, err = db.Exec("INSERT OR REPLACE INTO lobby VALUES (?, ?)", user, topic)
	return
}

func queryTopic(db *sql.DB, topic string) (user int64, err error) {
	err = db.QueryRow("SELECT user FROM lobby WHERE topic = ? LIMIT 1", topic).Scan(&user)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return
}

func joinLobby(db *sql.DB, user int64) (err error) {
	_, err = db.Exec("INSERT OR REPLACE INTO lobby VALUES (?, NULL)", user)
	return
}

func leaveLobby(db *sql.DB, user int64) (err error) {
	_, err = db.Exec("DELETE FROM lobby WHERE user = ?", user)
	return
}

func isUserInChat(db *sql.DB, user int64) (ok bool, err error) {
	var count int
	err = db.QueryRow("SELECT count(*) FROM match WHERE a = ?", user).Scan(&count)
	if err != nil {
		return false, err
	}
	return count != 0, nil
}

func isUserInLobby(db *sql.DB, user int64) (ok bool, err error) {
	var count int
	err = db.QueryRow("SELECT count(*) FROM lobby WHERE user = ?", user).Scan(&count)
	if err != nil {
		return false, err
	}
	return count != 0, nil
}

func isUserInQueue(db *sql.DB, user int64) (ok bool, err error) {
	var count int
	err = db.QueryRow("SELECT count(*) FROM lobby WHERE user = ? AND topic IS NOT NULL", user).Scan(&count)
	if err != nil {
		return false, err
	}
	return count != 0, nil
}

