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
	"fmt"
	"log"
	"time"
	"strings"
	"database/sql"
	"gopkg.in/telegram-bot-api.v4"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "./bot.db")
	checkErr(err)

	err = createTables(db)
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

			log_text := msg.Text
			if !DEBUG_MODE && !strings.HasPrefix(msg.Text, "/") {
				log_text = "(scrambled)"
			}
			if msg.Chat.LastName == "" {
				log.Printf("[%s]: %s\n", msg.Chat.FirstName, log_text)
			} else {
				log.Printf("[%s %s]: %s\n", msg.Chat.FirstName, msg.Chat.LastName, log_text)
			}

			cmd := msg.Command()
			if cmd == "start" {
				handleStart(bot, db, msg)
			} else if cmd == "list" {
				handleList(bot, db, msg)
			} else if cmd == "leave" {
				handleLeave(bot, db, msg)
			} else if cmd == "disconnect" {
				handleDisconnect(bot, db, msg)
			} else {
				handleMessage(bot, db, msg)
			}

		}

		edit_msg := update.EditedMessage
		if edit_msg != nil && edit_msg.Chat.IsPrivate() {
			quickReply(
				"「世界树」\n" +
				"\n" +
				"本服务原则上不保留聊天记录，故无法追踪消息编辑状态。\n" +
				"由于这个限制，你无法使用消息编辑功能。",
				bot, edit_msg)
		}

		query := update.CallbackQuery
		if query != nil {
			handleCallbackQuery(bot, db, query)
		}

	}
}

func handleStart(bot *tgbotapi.BotAPI, db *sql.DB, msg *tgbotapi.Message) {
	user_a := msg.Chat.ID

	// Detect whether the user is in chat.
	ok, err := isUserInChat(db, user_a)
	if err != nil {
		replyErr(err, bot, msg)
		return
	}
	if ok {
		quickReply(
			"「世界树」\n" +
			"\n" +
			"你正在一次会话中。\n" +
			"先戳 /leave 离开本次谈话，才能开始下一个会话。",
			bot, msg)
		return
	}

	// Detect whether the user is not in lobby yet.
	ok, err = isUserInLobby(db, user_a)
	if err != nil {
		replyErr(err, bot, msg)
		return
	}
	if !ok {
		if !IsOpenHour(time.Now()) {
			if !DEBUG_MODE {
				quickReply(
					"「世界树」\n" +
					CLOSED_MSG,
					bot, msg)
				return
			}
		}

		err = joinLobby(db, user_a)
		if err != nil {
			replyErr(err, bot, msg)
			return
		}
		// Fall through
	}

	// The user should be in lobby.
	chat, lobby, err := getActiveUsers(db)
	if err != nil {
		replyErr(err, bot, msg)
		return
	}
	quickReply(fmt.Sprintf(
		"欢迎使用「世界树」！\n" +
		"——长夜漫漫，随便找个人，陪你聊到天亮。\n" +
		"\n" +
		"你已进入大厅。在这里，你可以匿名地发布聊天话题。\n" +
		"如果有人为你点赞，你们将开始一段匿名的私人聊天。\n" +
		"你至多可发布一条话题，或为一个人点赞，最后一次操作有效。\n" +
		"\n" +
		"当前有 %d 人连接到世界树，其中 %d 人在大厅。\n" +
		"若要离开世界树，请戳 /disconnect 。\n" +
		"请善待他人，遵守道德和法律。",
		chat + lobby, lobby), bot, msg)
	sendTopicList(bot, db, user_a,
		"「世界树」\n" +
		"\n" +
		"你对这些话题感兴趣吗？\n" +
		"点击感兴趣的话题即可立刻开始聊天：")
}

func handleList(bot *tgbotapi.BotAPI, db *sql.DB, msg *tgbotapi.Message) {
	user_a := msg.Chat.ID

	// Detect whether the user is in chat.
	ok, err := isUserInChat(db, user_a)
	if err != nil {
		replyErr(err, bot, msg)
		return
	}
	if ok {
		count, err := sendTopicList(bot, db, user_a,
			"「世界树」\n" +
			"\n" +
			"以下是大厅内的话题清单：\n")
		if err != nil {
			replyErr(err, bot, msg)
			return
		}
		if count == 0 {
			quickReply(
				"「世界树」\n" +
				"\n" +
				"当前大厅内没有话题。",
				bot, msg)
		}
		return
	}

	// Detect whether the user is in lobby.
	ok, err = isUserInLobby(db, user_a)
	if err != nil {
		replyErr(err, bot, msg)
		return
	}
	if ok {
		count, err := sendTopicList(bot, db, user_a,
			"「世界树」\n" +
			"\n" +
			"以下是大厅内的话题清单，\n" +
			"点击感兴趣的话题即可立刻开始聊天：")
		if err != nil {
			replyErr(err, bot, msg)
			return
		}
		if count == 0 {
			quickReply(
				"「世界树」\n" +
				"\n" +
				"当前大厅内没有话题。\n" +
				"何不发布一个呢？",
				bot, msg)
		}
		return
	}

	quickReply(
		"「世界树」！\n" +
		"——长夜漫漫，随便找个人，陪你聊到天亮。\n" +
		"\n" +
		"你尚未连接到世界树。\n" +
		"何不戳一下 /start 试试看？",
		bot, msg)
}

func handleLeave(bot *tgbotapi.BotAPI, db *sql.DB, msg *tgbotapi.Message) {
	user_a := msg.Chat.ID

	// Detect whether the user is in chat.
	ok, err := isUserInChat(db, user_a)
	if err != nil {
		replyErr(err, bot, msg)
		return
	}
	if ok {
		user_b, err := queryMatch(db, user_a)
		if err != nil {
			replyErr(err, bot, msg)
			// Ignore the error
		}
		err = disconnectChat(db, user_a)
		if err != nil {
			replyErr(err, bot, msg)
		}

		err = joinLobby(db, user_a)
		if err != nil {
			replyErr(err, bot, msg)
			return
		}
		err = joinLobby(db, user_b)
		if err != nil {
			replyErr(err, bot, msg)
			return
		}

		chat, lobby, err := getActiveUsers(db)
		if err != nil {
			replyErr(err, bot, msg)
			return
		}
		quickReply(fmt.Sprintf(
			"「世界树」\n" +
			"本次谈话已结束，你已回到大厅。\n" +
			"何不试试发布下一个聊天话题？\n" +
			"不想聊了就去看看漫画吧： t.cn/RaomgYF\n" +
			"\n" +
			"当前有 %d 人连接到世界树，其中 %d 人在大厅。\n" +
			"若要离开世界树，请戳 /disconnect 。",
			chat + lobby, lobby), bot, msg)
		reply := tgbotapi.NewMessage(user_b, fmt.Sprintf(
			"「世界树」\n" +
			"对方结束了本次谈话，你已回到大厅。\n" +
			"何不试试发布下一个聊天话题？\n" +
			"不想聊了就去看看漫画吧： t.cn/RaomgYF\n" +
			"\n" +
			"当前有 %d 人连接到世界树，其中 %d 人在大厅。\n" +
			"若要离开世界树，请戳 /disconnect 。",
			chat + lobby, lobby))
		reply.DisableWebPagePreview = true
		_, err = bot.Send(reply)
		if err != nil {
			replyErr(err, bot, msg)
			return
		}
		sendTopicList(bot, db, user_a,
			"「世界树」\n" +
			"\n" +
			"你对这些话题感兴趣吗？\n" +
			"点击感兴趣的话题即可立刻开始聊天：")
		sendTopicList(bot, db, user_b,
			"「世界树」\n" +
			"\n" +
			"你对这些话题感兴趣吗？\n" +
			"点击感兴趣的话题即可立刻开始聊天：")
		return
	}

	// Detect whether the user is in queue.
	ok, err = isUserInQueue(db, user_a)
	if err != nil {
		replyErr(err, bot, msg)
		return
	}
	if ok {
		err := joinLobby(db, user_a)
		if err != nil {
			replyErr(err, bot, msg)
			return
		}
		quickReply(
			"「世界树」\n" +
			"\n" +
			"已取消你最后发布的话题。\n" +
			"若要断开与世界树的连接，请戳 /disconnect 。",
			bot, msg)
		return
	}

	// Detect whether the user is in lobby.
	ok, err = isUserInLobby(db, user_a)
	if err != nil {
		replyErr(err, bot, msg)
		return
	}
	if ok {
		quickReply(
			"「世界树」\n" +
			"\n" +
			"你当前没有发布任何话题。\n" +
			"若要断开与世界树的连接，请戳 /disconnect 。",
			bot, msg)
		return
	}

	quickReply(
		"「世界树」！\n" +
		"——长夜漫漫，随便找个人，陪你聊到天亮。\n" +
		"\n" +
		"你尚未连接到世界树。\n" +
		"何不戳一下 /start 试试看？",
		bot, msg)
}

func handleDisconnect(bot *tgbotapi.BotAPI, db *sql.DB, msg *tgbotapi.Message) {
	user_a := msg.Chat.ID

	// Detect whether the user is in chat.
	ok, err := isUserInChat(db, user_a)
	if err != nil {
		replyErr(err, bot, msg)
		return
	}
	if ok {
		quickReply(
			"「世界树」\n" +
			"\n" +
			"你正在一次会话中。\n" +
			"先戳 /leave 离开本次谈话。",
			bot, msg)
		return
	}

	// Detect whether the user is in lobby.
	ok, err = isUserInLobby(db, user_a)
	if err != nil {
		replyErr(err, bot, msg)
		return
	}
	if ok {
		err = leaveLobby(db, user_a)
		if err != nil {
			replyErr(err, bot, msg)
			return
		}
		quickReply(
			"「世界树」\n" +
			"\n" +
			"你已断开与世界树的连接。\n" +
			"世界树会想念你的，记得常回来看看～\n" +
			"如果喜欢的话，请推荐世界树 @WorldTreeBot 给朋友。\n" +
			"\n" +
			"戳 /start 重新开始。",
			bot, msg)
		return
	}

	quickReply(
		"「世界树」！\n" +
		"——长夜漫漫，随便找个人，陪你聊到天亮。\n" +
		"\n" +
		"你尚未连接到世界树。\n" +
		"何不戳一下 /start 试试看？",
		bot, msg)
}

func handleMessage(bot *tgbotapi.BotAPI, db *sql.DB, msg *tgbotapi.Message) {
	user_a := msg.Chat.ID

	// Detect whether the user is in chat.
	ok, err := isUserInChat(db, user_a)
	if err != nil {
		replyErr(err, bot, msg)
		return
	}
	if ok {
		// Forward the message to the partner
		user_b, err := queryMatch(db, user_a)
		if err != nil {
			replyErr(err, bot, msg)
			return
		}
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
				"\n" +
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
			fwd.Caption = msg.Caption
			fwd.Duration = msg.Audio.Duration
			fwd.Performer = msg.Audio.Performer
			fwd.Title = msg.Audio.Title
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
				fwd.Caption = msg.Caption
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
			fwd.Duration = msg.Video.Duration
			fwd.Caption = msg.Caption
			_, err = bot.Send(fwd)
			if err != nil {
				replyErr(err, bot, msg)
			}
		}
		if msg.Voice != nil {
			fwd := tgbotapi.NewVoiceShare(user_b, msg.Voice.FileID)
			fwd.Caption = msg.Caption
			fwd.Duration = msg.Voice.Duration
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
		return
	}

	// Detect whether the user is in lobby.
	ok, err = isUserInLobby(db, user_a)
	if err != nil {
		replyErr(err, bot, msg)
		return
	}
	if ok {
		topic := strings.TrimSpace(msg.Text)
		if topic == "" {
			return
		}
		user_b, err := queryTopic(db, topic)
		if err != nil {
			replyErr(err, bot, msg)
			return
		}
		if user_b == 0 || user_b == user_a {
			if !IsOpenHour(time.Now()) {
				if !DEBUG_MODE {
					quickReply(
						"「世界树」\n" +
						CLOSED_MSG,
						bot, msg)
					return
				}
			}

			err = setTopic(db, user_a, topic)
			if err != nil {
				replyErr(err, bot, msg)
				return
			}
			quickReply(
				"「世界树」\n" +
				"\n" +
				"你发布了：" + topic + "\n" +
				"\n" +
				"请等待世界树配对一个点赞的人。\n" +
				"或戳 /list 看看还有哪些话题。",
				bot, msg)
			broadcastNewTopic(bot, db, topic, user_a)
		} else {
			// Found a match
			err = leaveLobby(db, user_a)
			if err != nil {
				replyErr(err, bot, msg)
				return
			}
			err = leaveLobby(db, user_b)
			if err != nil {
				replyErr(err, bot, msg)
				return
			}

			quickReply(
				"「世界树」\n" +
				"\n" +
				"你发布了：" + topic + "\n" +
				"\n" +
				"请等待世界树配对一个点赞的人。",
				bot, msg)

			err = connectChat(db, user_a, user_b)
			if err != nil {
				replyErr(err, bot, msg)
				return
			}

			match_ok := "「世界树」\n" +
				"会话已接通，祝你们聊天愉快。\n" +
				"话题：" + topic + "\n" +
				"戳 /leave 离开本次谈话。\n" +
				"\n"
			if DEBUG_MODE {
				match_ok += "注：当前程序运行在调试模式下，管理员可能会看到聊天记录。"
			} else {
				match_ok += "注：本服务不保证密码学等级的防窃听，但原则上不保留聊天记录。"
			}
			reply := tgbotapi.NewMessage(user_a, match_ok)
			bot.Send(reply)
			reply = tgbotapi.NewMessage(user_b, match_ok)
			_, err = bot.Send(reply)
			if err != nil {
				replyErr(err, bot, msg)
				return
			}
		}
		return
	}

	quickReply(
		"「世界树」！\n" +
		"——长夜漫漫，随便找个人，陪你聊到天亮。\n" +
		"\n" +
		"你尚未连接到世界树。\n" +
		"何不戳一下 /start 试试看？",
		bot, msg)
}

func handleCallbackQuery(bot *tgbotapi.BotAPI, db *sql.DB, query *tgbotapi.CallbackQuery) {
	msg := query.Message
	if msg == nil || !msg.Chat.IsPrivate() {
		return
	}
	user_a := msg.Chat.ID

	// Detect whether the user is in chat.
	ok, err := isUserInChat(db, user_a)
	if err != nil {
		replyErr(err, bot, msg)
		return
	}
	if ok {
		quickReply(
			"「世界树」\n" +
			"\n" +
			"你正在一次会话中。\n" +
			"先戳 /leave 离开本次谈话，才能开始下一个会话。",
			bot, msg)
		bot.AnswerCallbackQuery(tgbotapi.CallbackConfig {
			CallbackQueryID: query.ID,
		})
		return
	}

	topic := query.Data
	if topic == "" {
		return
	}
	bot.AnswerCallbackQuery(tgbotapi.CallbackConfig {
		CallbackQueryID: query.ID,
		Text: "你 \u2764\ufe0f 了：" + topic,
	})
	user_b, err := queryTopic(db, topic)
	if err != nil {
		replyErr(err, bot, msg)
		return
	}
	if user_b == 0 || user_b == user_a {
		// The topic has gone.
		if !IsOpenHour(time.Now()) {
			if !DEBUG_MODE {
				quickReply(
					"「世界树」\n" +
					CLOSED_MSG,
					bot, msg)
				return
			}
		}

		err = setTopic(db, user_a, topic)
		if err != nil {
			replyErr(err, bot, msg)
			return
		}
		quickReply(
			"「世界树」\n" +
			"\n" +
			"你 \u2764\ufe0f 了：" + topic + "\n" +
			"\n" +
			"请等待世界树配对另一个点赞的人。\n" +
			"或戳 /list 看看还有哪些话题。",
			bot, msg)
		broadcastNewTopic(bot, db, topic, user_a)
	} else {
		err = leaveLobby(db, user_a)
		if err != nil {
			replyErr(err, bot, msg)
			return
		}
		err = leaveLobby(db, user_b)
		if err != nil {
			replyErr(err, bot, msg)
			return
		}

		quickReply(
			"「世界树」\n" +
			"\n" +
			"你 \u2764\ufe0f 了：" + topic + "\n" +
			"\n" +
			"请等待世界树配对另一个点赞的人。",
			bot, msg)

		err = connectChat(db, user_a, user_b)
		if err != nil {
			replyErr(err, bot, msg)
			return
		}

		match_ok := "「世界树」\n" +
			"会话已接通，祝你们聊天愉快。\n" +
			"话题：" + topic + "\n" +
			"戳 /leave 离开本次谈话。\n" +
			"\n"
		if DEBUG_MODE {
			match_ok += "注：当前程序运行在调试模式下，管理员可能会看到聊天记录。"
		} else {
			match_ok += "注：本服务不保证密码学等级的防窃听，但原则上不保留聊天记录。"
		}
		reply := tgbotapi.NewMessage(user_a, match_ok)
		bot.Send(reply)
		reply = tgbotapi.NewMessage(user_b, match_ok)
		_, err = bot.Send(reply)
		if err != nil {
			replyErr(err, bot, msg)
			return
		}
	}
}

func sendTopicList(bot *tgbotapi.BotAPI, db *sql.DB, user int64, caption string) (count int, err error) {
	topics, err := listTopics(db)
	if err != nil {
		return
	}
	count = len(topics)
	if count == 0 {
		return
	}
	if count > 10 {
		count = 10
	}
	reply := tgbotapi.NewMessage(user, caption)
	keyboard := make([][]tgbotapi.InlineKeyboardButton, count)
	for i := 0; i < count; i++ {
		keyboard[i] = []tgbotapi.InlineKeyboardButton {
			tgbotapi.InlineKeyboardButton {
				Text: topics[i],
				CallbackData: &topics[i],
			},
		}
	}
	reply.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
	_, err = bot.Send(reply)
	return
}

func broadcastNewTopic(bot *tgbotapi.BotAPI, db *sql.DB, topic string, exclude_user int64) {
	users, err := listPendingUsers(db)
	if err != nil {
		log.Println(err)
		return
	}
	reply_markup := tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton {
			tgbotapi.InlineKeyboardButton {
				Text: "\u2764\ufe0f 赞",
				CallbackData: &topic,
			},
		})
	for i := range users {
		if users[i] == exclude_user {
			continue
		}
		reply := tgbotapi.NewMessage(users[i],
			"「世界树」\n" +
			"有人新发布了以下话题：\n" +
			"\n" +
			topic)
		reply.ReplyMarkup = reply_markup
		bot.Send(reply)
	}
}

func quickReply(text string, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) (err error) {
	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	reply.ReplyToMessageID = msg.MessageID
	reply.DisableWebPagePreview = true
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
