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
	"fmt"
	"log"

	// "gopkg.in/telegram-bot-api.v4"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if DEBUG_MODE {
		log.Println("The program is running in debug mode,")
		log.Println("  the time limit will be disabled,")
		log.Println("  all chat logs will be print.")
	}

	db, err := sql.Open("sqlite3", "./bot.db")
	checkError(err)

	dbm := NewDBManager(db)

	err = dbm.CreateTables()
	checkError(err)

	log.Println("Database initialized.")

	api, err := tgbotapi.NewBotAPI(SECRET)
	checkError(err)

	log.Println("Bot API connected.")

	bot, err := NewBot(api, dbm)
	checkError(err)

	log.Println("Controller initialized.")

	bot.Run()
}

func printLog(user *tgbotapi.User, text string, scramble bool) {
	if !DEBUG_MODE && scramble {
		text = "(scrambled)"
	}
	var user_repr string
	user_repr += fmt.Sprintf("#%d ", user.ID)
	if user.UserName != "" {
		user_repr += fmt.Sprintf("@%s ", user.UserName)
	}
	user_repr += user.FirstName
	if user.LastName != "" {
		user_repr += " " + user.LastName
	}
	log.Printf("[%s]: %s\n", user_repr, text)
}

func checkError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
