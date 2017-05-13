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
	"strconv"
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

		query := update.CallbackQuery
		if query != nil {
			handleCallbackQuery(bot, db, query)
		}

	}
}

func handleStart(bot *tgbotapi.BotAPI, db *sql.DB, msg *tgbotapi.Message) {
	quickReply(
		"欢迎使用「世界树」！\n" +
		"长夜漫漫，随便找个人，陪你聊到天亮。\n" +
		"\n" +
		"戳 /new 寻找今天的聊伴。\n" +
		"匿名聊天时请遵守道德和法律。",
		bot, msg)
}

func handleNewChat(bot *tgbotapi.BotAPI, db *sql.DB, msg *tgbotapi.Message) {
	user_a := msg.Chat.ID

	// Detect whether the user is already in a chat.
	user_b, err := queryUser(db, user_a)
	if err != nil {
		replyErr(err, bot, msg)
		return
	}
	if user_b != 0 {
		quickReply(
			"「世界树」\n" +
			"你正在一次会话中。\n" +
			"先戳 /leave 离开本次谈话，才能开始下一个会话。",
			bot, msg)
		return
	}

	if !IsOpenHour(time.Now()) {
		if !DEBUG_MODE {
			quickReply(
				"「世界树」\n" +
				CLOSED_MSG,
				bot, msg)
			return
		}
	}

	// Cancel if the user is already in a queue.
	err = cancelTopic(db, user_a)
	if err != nil {
		replyErr(err, bot, msg)
		return
	}

	// List topics for the user.
	topics, err := listTopics(db)
	if err != nil {
		replyErr(err, bot, msg)
		return
	}
	err = setChoosingStatus(db, user_a, true)
	if err != nil {
		replyErr(err, bot, msg)
		return
	}
	active_users, err := getActiveUsers(db)
	if err != nil {
		replyErr(err, bot, msg)
		return
	}
	if len(topics) == 0 {
		reply := tgbotapi.NewMessage(user_a,
			"「世界树」\n" +
			"当前已有 " + strconv.Itoa(active_users+1) + " 个用户连接到世界树。\n" +
			"\n" +
			"在开始之前，请输入你想讨论的话题。\n" +
			"其它人会看到你的话题并与你交谈。")
		reply.ReplyMarkup = tgbotapi.ForceReply {
			ForceReply: true,
			Selective: true,
		}
		bot.Send(reply)
	} else {
		reply := tgbotapi.NewMessage(user_a,
			"「世界树」\n" +
			"当前已有 " + strconv.Itoa(active_users+1) + " 个用户连接到世界树。\n" +
			"\n" +
			"以下是等待中的话题。\n" +
			"点击你感兴趣的话题与对方聊天。\n" +
			"\n" +
			"如果你想讨论别的话题，请直接输入。\n" +
			"其它人会看到你的话题并与你交谈。")
		keyboard := make([][]tgbotapi.InlineKeyboardButton, len(topics))
		for i := range topics {
			topic_data := "+" + topics[i];
			keyboard[i] = []tgbotapi.InlineKeyboardButton {
				tgbotapi.InlineKeyboardButton {
					Text: topics[i],
					CallbackData: &topic_data,
				},
			}
		}
		reply.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
		bot.Send(reply)
	}
}

func handleCallbackQuery(bot *tgbotapi.BotAPI, db *sql.DB, query *tgbotapi.CallbackQuery) {
	msg := query.Message
	if msg == nil || !msg.Chat.IsPrivate() {
		return
	}
	user_a := msg.Chat.ID

	// Detect whether the user is already in a chat.
	user_b, err := queryUser(db, user_a)
	if err != nil {
		replyErr(err, bot, msg)
		return
	}
	if user_b != 0 {
		quickReply(
			"「世界树」\n" +
			"你正在一次会话中。\n" +
			"先戳 /leave 离开本次谈话，才能开始下一个会话。",
			bot, msg)
		return
	}

	err = setChoosingStatus(db, user_a, false)
	if err != nil {
		replyErr(err, bot, msg)
	}

	will_repost := query.Data[:1]
	topic := query.Data[1:]
	user_b, err = popTopic(db, topic)
	if err != nil {
		replyErr(err, bot, msg)
		return
	}
	if user_b == 0 {
		// The topic has gone.
		if will_repost == "+" {
			err = pushTopic(db, user_a, topic)
			if err != nil {
				replyErr(err, bot, msg)
				return
			}
			quickReply(
				"「世界树」\n" +
				"你可能来晚了，匹配失败。\n" +
				"我们为你重新排队，请等待下一个志趣相投的人。\n" +
				"也可以戳 /new 重新选择话题。",
				bot, msg)
			broadcastNewTopic(bot, db, topic, user_a)
		} else {
			quickReply(
				"「世界树」\n" +
				"你可能来晚了，匹配失败。\n" +
				"也可以戳 /new 重新选择话题。",
				bot, msg)
		}
		return
	}

	err = cancelTopic(db, user_a)
	if err != nil {
		replyErr(err, bot, msg)
		return
	}

	err = connectUser(db, user_a, user_b)
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
	replyErr(err, bot, msg)
}

func handleLeaveChat(bot *tgbotapi.BotAPI, db *sql.DB, msg *tgbotapi.Message) {
	user_a := msg.Chat.ID

	// Detect choosing status.
	choosing, err := getChoosingStatus(db, user_a)
	if err != nil {
		replyErr(err, bot, msg)
		return
	}
	if choosing {
		setChoosingStatus(db, user_a, false)
		if err != nil {
			replyErr(err, bot, msg)
			return
		}
		quickReply(
			"「世界树」\n" +
			"已放弃排队。",
			bot, msg)
		return
	}

	// Detect whether the user is already in queue.
	ok, err := isUserInQueue(db, user_a)
	if err != nil {
		replyErr(err, bot, msg)
		return
	}
	if ok {
		err := cancelTopic(db, user_a)
		if err != nil {
			replyErr(err, bot, msg)
			return
		}
		quickReply(
			"「世界树」\n" +
			"已放弃排队。\n" +
			"戳 /new 重新选择话题。",
			bot, msg)
		return
	}

	// Detect whether the user is not in chat.
	user_b, err := queryUser(db, user_a)
	if user_b == 0 {
		quickReply(
			"「世界树」\n" +
			"你现在不在会话中。\n" +
			"要不要试试戳 /new 开始一段聊天？",
			bot, msg)
	} else {
		// Disconnect with the partnet
		err = disconnectUser(db, user_a)
		if err != nil {
			replyErr(err, bot, msg)
		}
		quickReply(
			"「世界树」\n" +
			"本次谈话已结束。\n" +
			"要不要试试戳 /new 换一个聊伴？\n" +
			"\n" +
			"我们希望你能玩得开心，并向朋友推荐 @WorldTreeBot 。",
			bot, msg)
		reply := tgbotapi.NewMessage(user_b,
			"「世界树」\n" +
			"对方结束了本次谈话。\n" +
			"要不要试试戳 /new 换一个聊伴？\n" +
			"\n" +
			"我们希望你能玩得开心，并向朋友推荐 @WorldTreeBot 。",
			)
		_, err = bot.Send(reply)
		replyErr(err, bot, msg)
	}
}

func handleMessage(bot *tgbotapi.BotAPI, db *sql.DB, msg *tgbotapi.Message) {
	user_a := msg.Chat.ID

	// Detect choosing status.
	choosing, err := getChoosingStatus(db, user_a)
	if err != nil {
		replyErr(err, bot, msg)
		return
	}
	if choosing {
		setChoosingStatus(db, user_a, false)
		if err != nil {
			replyErr(err, bot, msg)
			return
		}
		topic := strings.TrimSpace(msg.Text)
		user_b, err := popTopic(db, topic)
		if err != nil {
			replyErr(err, bot, msg)
			return
		}
		if user_b == 0 {
			// Push into topic queue
			err = pushTopic(db, user_a, topic)
			if err != nil {
				replyErr(err, bot, msg)
				return
			}
			quickReply(
				"「世界树」\n" +
				"我们已为你排队，请等待下一个志趣相投的人，有消息会通知你的。\n" +
				"在此期间，不妨去看看漫画吧： t.cn/RaomgYF\n" +
				"若要放弃，请戳 /leave 。",
				bot, msg)
			broadcastNewTopic(bot, db, topic, user_a)
		} else {
			// Connect 
			err = connectUser(db, user_a, user_b)
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
			replyErr(err, bot, msg)
		}
		return
	}

	// Detect whether the user is already in queue.
	ok, err := isUserInQueue(db, user_a)
	if err != nil {
		replyErr(err, bot, msg)
		return
	}
	if ok {
		quickReply(
			"「世界树」\n" +
			"我们已为你排队，请等待下一个志趣相投的人，有消息会通知你的。\n" +
			"在此期间，不妨去看看漫画吧： t.cn/RaomgYF\n" +
			"若要放弃，请戳 /leave 。",
			bot, msg)
		return
	}

	// Detect whether the user is not in chat.
	user_b, err := queryUser(db, user_a)
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
	} else {
		// Forward the message to the partner
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
	}
}

func broadcastNewTopic(bot *tgbotapi.BotAPI, db *sql.DB, topic string, exclude_user int64) {
	users, err := listPendingUsers(db)
	if err != nil {
		log.Println(err)
		return
	}
	topic_data := "-" + topic;
	reply_markup := tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton {
			tgbotapi.InlineKeyboardButton {
				Text: topic,
				CallbackData: &topic_data,
			},
		})
	for i := range users {
		if users[i] == exclude_user {
			continue
		}
		reply := tgbotapi.NewMessage(users[i],
			"「世界树」\n" +
			"有人新发布了以下话题。\n" +
			"你要放弃自己当前的话题，与对方聊天吗？\n" +
			"如果想与对方交谈，请点击按钮。")
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
