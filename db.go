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

type dbManager struct {
	db *sql.DB
}

func NewDBManager(db *sql.DB) *dbManager {
	return &dbManager{
		db: db,
	}
}

func (dbm *dbManager) CreateTables() (err error) {
	_, err = dbm.db.Exec("CREATE TABLE IF NOT EXISTS admin (user INTEGER PRIMARY KEY)")
	if err != nil {
		return
	}
	_, err = dbm.db.Exec("CREATE TABLE IF NOT EXISTS invite (user INTEGER PRIMARY KEY, topic TEXT)")
	if err != nil {
		return
	}
	_, err = dbm.db.Exec("CREATE TABLE IF NOT EXISTS lobby (user INTEGER PRIMARY KEY, room INTEGER)")
	if err != nil {
		return
	}
	_, err = dbm.db.Exec("CREATE TABLE IF NOT EXISTS chat (user_a INTEGER PRIMARY KEY, user_b INTEGER)")
	if err != nil {
		return
	}
	_, err = dbm.db.Exec("CREATE TABLE IF NOT EXISTS banlist (user INTEGER PRIMARY KEY)")
	return
}

// Users

func (dbm *dbManager) GetActiveUsers() (chat int, lobby int, err error) {
	err = dbm.db.QueryRow("SELECT count(*) FROM chat").Scan(&chat)
	if err != nil {
		return
	}
	err = dbm.db.QueryRow("SELECT count(*) FROM lobby").Scan(&lobby)
	return
}

func (dbm *dbManager) ListAllUsers() (users []int64, err error) {
	rows, err := dbm.db.Query("SELECT user_a FROM chat ORDER BY random()")
	if err != nil {
		return
	}
	{
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
		if err != nil {
			return
		}
	}
	rows, err = dbm.db.Query("SELECT user FROM lobby ORDER BY random()")
	if err != nil {
		return
	}
	{
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
		if err != nil {
			return
		}
	}
	return
}

// Chats

func (dbm *dbManager) ConnectChat(user_a int64, user_b int64) (err error) {
	tx, err := dbm.db.Begin()
	if err != nil {
		return
	}

	_, err = tx.Exec("INSERT OR REPLACE INTO chat VALUES (?, ?)", user_a, user_b)
	if err != nil {
		tx.Rollback()
		return
	}

	_, err = tx.Exec("INSERT OR REPLACE INTO chat VALUES (?, ?)", user_b, user_a)
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

func (dbm *dbManager) DisconnectChat(user_a int64, user_b int64) (err error) {
	if user_b != 0 {
		_, err = dbm.db.Exec("UPDATE chat SET user_b = 0 WHERE user_a = ?", user_b)
	}
	_, err1 := dbm.db.Exec("DELETE FROM chat WHERE user_a = ?", user_a)
	if err != nil {
		return
	}
	return err1
}

func (dbm *dbManager) QueryChat(user_a int64) (user_b int64, err error) {
	err = dbm.db.QueryRow("SELECT user_b FROM chat WHERE user_a = ? LIMIT 1", user_a).Scan(&user_b)
	return
}

func (dbm *dbManager) ListUnmatchedUsers() (users []int64, err error) {
	rows, err := dbm.db.Query("SELECT user FROM lobby ORDER BY random()")
	if err != nil {
		return
	}
	{
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
		if err != nil {
			return
		}
	}
	rows, err = dbm.db.Query("SELECT user_a FROM chat WHERE user_b = 0 ORDER BY random()")
	if err != nil {
		return
	}
	{
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
		if err != nil {
			return
		}
	}
	return
}

// Invitations

func (dbm *dbManager) NewInvitation(user int64, topic string) (err error) {
	_, err = dbm.db.Exec("INSERT OR REPLACE INTO invite VALUES (?, ?)", user, topic)
	return
}

func (dbm *dbManager) NewPendingInvitation(user int64) (err error) {
	_, err = dbm.db.Exec("INSERT OR REPLACE INTO invite VALUES (?, NULL)", user)
	return
}

func (dbm *dbManager) RemoveInvitation(user int64) (err error) {
	_, err = dbm.db.Exec("DELETE FROM invite WHERE user = ?", user)
	return
}

func (dbm *dbManager) RemoveInvitationByTopic(topic string) (err error) {
	_, err = dbm.db.Exec("DELETE FROM invite WHERE topic = ?", topic)
	return
}

func (dbm *dbManager) QueryInvitation(topic string) (user int64, err error) {
	err = dbm.db.QueryRow("SELECT user FROM invite WHERE topic = ? LIMIT 1", topic).Scan(&user)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return
}

func (dbm *dbManager) ListInvites() (topics []string, err error) {
	rows, err := dbm.db.Query("SELECT topic FROM invite WHERE topic IS NOT NULL ORDER BY random()")
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

// Lobbies

func (dbm *dbManager) JoinLobby(user int64, room int64) (err error) {
	_, err = dbm.db.Exec("INSERT OR REPLACE INTO lobby VALUES (?, ?)", user, room)
	return
}

func (dbm *dbManager) LeaveLobby(user int64) (err error) {
	_, err = dbm.db.Exec("DELETE FROM lobby WHERE user = ?", user)
	return
}

func (dbm *dbManager) QueryLobby(user int64) (room int64, err error) {
	err = dbm.db.QueryRow("SELECT room FROM lobby WHERE user = ?", user).Scan(&room)
	return
}

func (dbm *dbManager) ListUsersInLobby(room int64) (users []int64, err error) {
	rows, err := dbm.db.Query("SELECT user FROM lobby WHERE room = ? ORDER BY random()", room)
	if err != nil {
		return
	}
	{
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
		if err != nil {
			return
		}
	}
	return
}

// Status

func (dbm *dbManager) IsUserInChat(user int64) (ok bool, err error) {
	var count int
	err = dbm.db.QueryRow("SELECT count(*) FROM chat WHERE user_a = ?", user).Scan(&count)
	if err != nil {
		return false, err
	}
	return count != 0, nil
}

func (dbm *dbManager) IsUserInLobby(user int64) (ok bool, err error) {
	var count int
	err = dbm.db.QueryRow("SELECT count(*) FROM lobby WHERE user = ?", user).Scan(&count)
	if err != nil {
		return false, err
	}
	return count != 0, nil
}

func (dbm *dbManager) IsUserTypingTopic(user int64) (ok bool, err error) {
	var count int
	err = dbm.db.QueryRow("SELECT count(*) FROM invite WHERE user = ? AND topic IS NULL", user).Scan(&count)
	if err != nil {
		return false, err
	}
	return count != 0, nil
}

func (dbm *dbManager) IsUserInQueue(user int64) (ok bool, err error) {
	var count int
	err = dbm.db.QueryRow("SELECT count(*) FROM invite WHERE user = ?", user).Scan(&count)
	if err != nil {
		return false, err
	}
	return count != 0, nil
}

func (dbm *dbManager) IsUserAnAdmin(user int64) (ok bool, err error) {
	var count int
	err = dbm.db.QueryRow("SELECT count(*) FROM admin WHERE user = ?", user).Scan(&count)
	if err != nil {
		return false, err
	}
	return count != 0, nil
}

func (dbm *dbManager) IsUserInBanList(user int64) (ok bool, err error) {
	var count int
	err = dbm.db.QueryRow("SELECT count(*) FROM banlist WHERE user = ?", user).Scan(&count)
	if err != nil {
		return false, err
	}
	return count != 0, nil
}
