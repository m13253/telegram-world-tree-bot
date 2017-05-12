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
	"log"
	"time"
	"database/sql"
	"gopkg.in/telegram-bot-api.v4"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "./bot.db")
	checkErr(err)

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS match (a INTEGER UNIQUE, b INTEGER UNIQUE)")
	checkErr(err)

	log.Println("Database initialized.")

	bot, err := tgbotapi.NewBotAPI(SECRET)
	checkErr(err)

	log.Println("Bot API connected.")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		msg := update.Message
		if msg != nil && msg.Chat.IsPrivate() {

			if msg.Chat.LastName == "" {
				log.Printf("[%s]: %s\n", msg.Chat.FirstName, msg.Text)
			} else {
				log.Printf("[%s %s]: %s\n", msg.Chat.FirstName, msg.Chat.LastName, msg.Text)
			}
			cmd := msg.Command()
			if cmd == "start" {
				handleStart(bot, db, msg)
			} else if cmd == "new" {
				handleNewChat(bot, db, msg)
			} else if cmd == "leave" {
				handleLeaveChat(bot, db, msg)
			} else {
				handleMessage(bot, db, msg)
			}

		}

		edit_msg := update.EditedMessage
		if edit_msg != nil && edit_msg.Chat.IsPrivate() {
			quickReply(
				"「世界树」\n" +
				"本服务原则上不保留聊天记录，故无法追踪消息编辑状态。\n" +
				"由于这个限制，你无法使用消息编辑功能。",
				bot, edit_msg)
		}
	}
}

func handleStart(bot *tgbotapi.BotAPI, db *sql.DB, msg *tgbotapi.Message) {
	quickReply(
		"欢迎使用「世界树」！\n" +
		"长夜漫漫，随便找个人，陪你聊到天亮。\n" +
		"戳 /new 寻找今天的聊伴。\n" +
		"匿名聊天时请遵守道德和法律。",
		bot, msg)
}

func handleNewChat(bot *tgbotapi.BotAPI, db *sql.DB, msg *tgbotapi.Message) {
	user_a := msg.Chat.ID
	// Detect whether the user is already in queue or in a chat.
	user_b, err := queryUserB(db, user_a)
	if err != nil {
		replyErr(err, bot, msg)
		return
	}
	if user_b == -1 {
		quickReply(
			"「世界树」\n" +
			"现在还找不到聊伴，已为你排队。\n" +
			"在此期间，不妨去听听音乐？\n" +
			"戳 /leave 放弃排队。",
			bot, msg)
		return
	} else if user_b != 0 {
		quickReply(
			"「世界树」\n" +
			"你正在一次会话中。\n" +
			"先戳 /leave 离开本次谈话，才能开始下一个会话。",
			bot, msg)
		return
	}
	// Then, try to find anyone not paired.
	user_b, err = queryUserA(db, -1)
	if err != nil {
		replyErr(err, bot, msg)
		return
	}
	if user_b == 0 {
		if !IsOpenHour(time.Now()) {
			quickReply(
				"「世界树」\n" +
				CLOSED_MSG,
				bot, msg)
			return
		}
		// Queue this user
		_, err = db.Exec("INSERT INTO match VALUES (?, -1)", user_a)
		if err != nil {
			replyErr(err, bot, msg)
			return
		}
		quickReply(
			"「世界树」\n" +
			"现在还找不到聊伴，已为你排队。\n" +
			"在此期间，不妨去听听音乐？\n" +
			"戳 /leave 放弃排队。",
			bot, msg)
	} else {
		// Found a pending user_a
		tx, err := db.Begin()
		if err != nil {
			replyErr(err, bot, msg)
			return
		}
		_, err = tx.Exec("INSERT OR REPLACE INTO match VALUES (?, ?)", user_a, user_b)
		if err != nil {
			tx.Rollback()
			replyErr(err, bot, msg)
			return
		}
		_, err = tx.Exec("INSERT OR REPLACE INTO match VALUES (?, ?)", user_b, user_a)
		if err != nil {
			tx.Rollback()
			replyErr(err, bot, msg)
			return
		}
		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			replyErr(err, bot, msg)
			return
		}
		const match_ok = "「世界树」\n" +
			"已匹配到另一个聊伴，祝你们聊天愉快。\n" +
			"戳 /leave 离开本次谈话。\n" +
			"\n" +
			"注：本服务不保证密码学等级的防窃听，但原则上不保留聊天记录。"
		reply := tgbotapi.NewMessage(user_a, match_ok)
		bot.Send(reply)
		reply = tgbotapi.NewMessage(user_b, match_ok)
		_, err = bot.Send(reply)
		replyErr(err, bot, msg)
	}
}

func handleLeaveChat(bot *tgbotapi.BotAPI, db *sql.DB, msg *tgbotapi.Message) {
	user_a := msg.Chat.ID
	// Detect whether the user is already in queue or in a chat.
	user_b, err := queryUserB(db, user_a)
	if err != nil {
		replyErr(err, bot, msg)
		return
	}
	if user_b == 0 {
		quickReply(
			"「世界树」\n" +
			"你现在不在会话中。\n" +
			"要不要试试戳 /new 开始一段聊天？",
			bot, msg)
	} else if user_b == -1 {
		_, err = db.Exec("DELETE FROM match WHERE a = ? OR b = ?", user_a, user_a)
		replyErr(err, bot, msg)
		quickReply(
			"「世界树」\n" +
			"已放弃排队。",
			bot, msg)
	} else {
		// Terminate this dialog.
		_, err = db.Exec("DELETE FROM match WHERE a = ? OR b = ?", user_a, user_a)
		replyErr(err, bot, msg)
		quickReply(
			"「世界树」\n" +
			"本次谈话已结束。\n" +
			"要不要试试戳 /new 换一个聊伴？",
			bot, msg)
		reply := tgbotapi.NewMessage(user_b,
			"「世界树」\n" +
			"对方结束了本次谈话。\n" +
			"要不要试试戳 /new 换一个聊伴？")
		_, err = bot.Send(reply)
		replyErr(err, bot, msg)
	}
}

func handleMessage(bot *tgbotapi.BotAPI, db *sql.DB, msg *tgbotapi.Message) {
	user_a := msg.Chat.ID
	user_b, err := queryUserB(db, user_a)
	if err != nil {
		replyErr(err, bot, msg)
		return
	}
	if user_b == 0 {
		quickReply(
			"「世界树」\n" +
			"你现在不在会话中。\n" +
			"要不要试试戳 /new 开始一段聊天？",
			bot, msg)
	} else if user_b == -1 {
		quickReply(
			"「世界树」\n" +
			"现在还找不到聊伴，已为你排队。\n" +
			"在此期间，不妨去听听音乐？\n" +
			"戳 /leave 放弃排队。",
			bot, msg)
	} else {
		if msg.ForwardFrom != nil || msg.ForwardFromChat != nil {
			fwd := tgbotapi.NewForward(user_b, msg.Chat.ID, msg.MessageID)
			_, err = bot.Send(fwd)
			if err != nil {
				replyErr(err, bot, msg)
			}
			return
		}
		if msg.ReplyToMessage != nil {
			quickReply(
				"「世界树」\n" +
				"本服务原则上不保留聊天记录，故无法追踪过去的消息。\n" +
				"由于这个限制，你无法使用定向回复功能。",
				bot, msg)
		}
		if msg.Text != "" {
			fwd := tgbotapi.NewMessage(user_b, msg.Text)
			_, err = bot.Send(fwd)
			if err != nil {
				replyErr(err, bot, msg)
			}
		}
		if msg.Audio != nil {
			fwd := tgbotapi.NewAudioShare(user_b, msg.Audio.FileID)
			_, err = bot.Send(fwd)
			if err != nil {
				replyErr(err, bot, msg)
			}
		}
		if msg.Document != nil {
			fwd := tgbotapi.NewDocumentShare(user_b, msg.Document.FileID)
			_, err = bot.Send(fwd)
			if err != nil {
				replyErr(err, bot, msg)
			}
		}
		if msg.Photo != nil {
			if len(*msg.Photo) != 0 {
				fwd := tgbotapi.NewPhotoShare(user_b, (*msg.Photo)[0].FileID)
				_, err = bot.Send(fwd)
				if err != nil {
					replyErr(err, bot, msg)
				}
			}
		}
		if msg.Sticker != nil {
			fwd := tgbotapi.NewStickerShare(user_b, msg.Sticker.FileID)
			_, err = bot.Send(fwd)
			if err != nil {
				replyErr(err, bot, msg)
			}
		}
		if msg.Video != nil {
			fwd := tgbotapi.NewVideoShare(user_b, msg.Video.FileID)
			_, err = bot.Send(fwd)
			if err != nil {
				replyErr(err, bot, msg)
			}
		}
		if msg.Voice != nil {
			fwd := tgbotapi.NewVoiceShare(user_b, msg.Voice.FileID)
			_, err = bot.Send(fwd)
			if err != nil {
				replyErr(err, bot, msg)
			}
		}
		if msg.Contact != nil {
			fwd := tgbotapi.NewContact(user_b, msg.Contact.PhoneNumber, msg.Contact.FirstName)
			fwd.LastName = msg.Contact.LastName
			_, err = bot.Send(fwd)
			if err != nil {
				replyErr(err, bot, msg)
			}
		}
		if msg.Location != nil {
			fwd := tgbotapi.NewLocation(user_b, msg.Location.Latitude, msg.Location.Longitude)
			_, err = bot.Send(fwd)
			if err != nil {
				replyErr(err, bot, msg)
			}
		}
		if msg.Venue != nil {
			fwd := tgbotapi.NewVenue(user_b, msg.Venue.Title, msg.Venue.Address, msg.Venue.Location.Latitude, msg.Venue.Location.Longitude)
			fwd.FoursquareID = msg.Venue.FoursquareID
			_, err = bot.Send(fwd)
			if err != nil {
				replyErr(err, bot, msg)
			}
		}
	}
}

func queryUserA(db *sql.DB, user_b int64) (user_a int64, err error) {
	err = db.QueryRow("SELECT a FROM match WHERE b = ?", user_b).Scan(&user_a)
	if err == sql.ErrNoRows {
		return 0, nil
	} else if err != nil {
		return 0, err
	} else {
		return
	}
}

func queryUserB(db *sql.DB, user_a int64) (user_b int64, err error) {
	err = db.QueryRow("SELECT b FROM match WHERE a = ?", user_a).Scan(&user_b)
	if err == sql.ErrNoRows {
		return 0, nil
	} else if err != nil {
		return 0, err
	} else {
		return
	}
}

func quickReply(text string, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) (err error) {
	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	reply.ReplyToMessageID = msg.MessageID
	_, err = bot.Send(reply)
	return
}

func replyErr(err error, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	if err != nil {
		log.Println(err)
		quickReply(
			"「世界树」\n" +
			"程序发生了错误，刚刚的消息可能没有送达。",
			bot, msg)
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
