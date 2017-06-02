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
	"time"
	"strings"
	"gopkg.in/telegram-bot-api.v4"
	_ "github.com/mattn/go-sqlite3"
)

func (bot *Bot) handleStart(msg *tgbotapi.Message) {
	user_a := msg.Chat.ID

	// Detect whether the user is in chat.
	ok, err := bot.dbm.IsUserInChat(user_a)
	if err != nil {
		bot.replyError(err, msg)
		return
	}
	if ok {
		bot.quickReply(
			"「世界树」\n" +
			"\n" +
			"你正在一次会话中。\n" +
			"先戳 /leave 离开本次谈话，才能开始下一个会话。",
			msg)
		return
	}

	// Detect whether the user is not in lobby yet.
	ok, err = bot.dbm.IsUserInLobby(user_a)
	if err != nil {
		bot.replyError(err, msg)
		return
	}
	if ok {
		chat, lobby, err := bot.dbm.GetActiveUsers()
		if err != nil {
			bot.replyError(err, msg)
			return
		}
		bot.quickReply(fmt.Sprintf(
			"欢迎使用「世界树」！\n" +
			"——长夜漫漫，随便找个人，陪你聊到天亮。\n" +
			"\n" +
			"你已进入大厅。在这里，你可以匿名地发布聊天话题。\n" +
			"如果有人为你点赞，你们将开始一段匿名的私人聊天。\n" +
			"你至多可发布一条话题，或为一个人点赞，最后一次操作有效。\n" +
			"\n" +
			"当前有 %d 人连接到世界树，其中 %d 人在大厅。\n" +
			"若要离开世界树，请戳 /disconnect 。\n" +
			"请友善待人，遵守道德和法律。",
			chat + lobby, lobby), msg)
		if !IsOpenHour(time.Now()) && !DEBUG_MODE {
			bot.quickReply(
				"「世界树」\n" +
				"\n" +
				"\u274c " + CLOSED_MSG,
				msg)
			return
		}
		bot.sendTopicList(user_a,
			"「世界树」\n" +
			"\n" +
			"你对这些话题感兴趣吗？\n" +
			"点击感兴趣的话题即可立刻开始聊天：")
		return
	}

	err = bot.dbm.JoinLobby(user_a)
	if err != nil {
		bot.replyError(err, msg)
		return
	}
	chat, lobby, err := bot.dbm.GetActiveUsers()
	if err != nil {
		bot.replyError(err, msg)
		return
	}
	bot.quickReply(fmt.Sprintf(
		"欢迎使用「世界树」！\n" +
		"——长夜漫漫，随便找个人，陪你聊到天亮。\n" +
		"\n" +
		"你已进入大厅。在这里，你可以匿名地发布聊天话题。\n" +
		"如果有人为你点赞，你们将开始一段匿名的私人聊天。\n" +
		"你至多可发布一条话题，或为一个人点赞，最后一次操作有效。\n" +
		"\n" +
		"当前有 %d 人连接到世界树，其中 %d 人在大厅。\n" +
		"若要离开世界树，请戳 /disconnect 。\n" +
		"请友善待人，遵守道德和法律。",
		chat + lobby, lobby), msg)
	if !IsOpenHour(time.Now()) && !DEBUG_MODE {
		bot.quickReply(
			"「世界树」\n" +
			"\n" +
			"\u274c " + CLOSED_MSG,
			msg)
		return
	}
	bot.sendTopicList(user_a,
		"「世界树」\n" +
		"\n" +
		"你对这些话题感兴趣吗？\n" +
		"点击感兴趣的话题即可立刻开始聊天：")
}

func (bot *Bot) handleList(msg *tgbotapi.Message) {
	user_a := msg.Chat.ID

	chat, lobby, err := bot.dbm.GetActiveUsers()
	if err != nil {
		bot.replyError(err, msg)
		return
	}

	// Detect whether the user is in chat.
	ok, err := bot.dbm.IsUserInChat(user_a)
	if err != nil {
		bot.replyError(err, msg)
		return
	}
	if ok {
		count, err := bot.sendTopicList(user_a, fmt.Sprintf(
			"「世界树」\n" +
			"\n" +
			"当前有 %d 人连接到世界树，其中 %d 人在大厅。\n" +
			"以下是大厅内的话题清单：\n",
			chat + lobby, lobby))
		if err != nil {
			bot.replyError(err, msg)
			return
		}
		if count == 0 {
			bot.quickReply(fmt.Sprintf(
				"「世界树」\n" +
				"\n" +
				"当前有 %d 人连接到世界树，其中 %d 人在大厅。\n" +
				"当前大厅内没有话题。",
				chat + lobby, lobby), msg)
		}
		return
	}

	// Detect whether the user is in lobby.
	ok, err = bot.dbm.IsUserInLobby(user_a)
	if err != nil {
		bot.replyError(err, msg)
		return
	}
	if ok {
		count, err := bot.sendTopicList(user_a, fmt.Sprintf(
			"「世界树」\n" +
			"\n" +
			"以下是大厅内的话题清单，\n" +
			"当前有 %d 人连接到世界树，其中 %d 人在大厅。\n" +
			"点击感兴趣的话题即可立刻开始聊天：",
			chat + lobby, lobby))
		if err != nil {
			bot.replyError(err, msg)
			return
		}
		if count == 0 {
			bot.quickReply(fmt.Sprintf(
				"「世界树」\n" +
				"\n" +
				"当前有 %d 人连接到世界树，其中 %d 人在大厅。\n" +
				"当前大厅内没有话题。\n" +
				"何不发布一个呢？",
				chat + lobby, lobby), msg)
		}
		return
	}

	bot.quickReply(
		"「世界树」\n" +
		"——长夜漫漫，随便找个人，陪你聊到天亮。\n" +
		"\n" +
		"你尚未连接到世界树。\n" +
		"何不戳一下 /start 试试看？",
		msg)
}

func (bot *Bot) handleLeave(msg *tgbotapi.Message) {
	user_a := msg.Chat.ID

	// Detect whether the user is in chat.
	ok, err := bot.dbm.IsUserInChat(user_a)
	if err != nil {
		bot.replyError(err, msg)
		return
	}
	if ok {
		user_b, err := bot.dbm.QueryMatch(user_a)
		if err != nil {
			bot.replyError(err, msg)
			// Ignore the error
		}
		err = bot.dbm.DisconnectChat(user_a, user_b)
		if err != nil {
			bot.replyError(err, msg)
		}

		err = bot.dbm.JoinLobby(user_a)
		if err != nil {
			bot.replyError(err, msg)
			return
		}

		chat, lobby, err := bot.dbm.GetActiveUsers()
		if err != nil {
			bot.replyError(err, msg)
			return
		}
		bot.quickReply(fmt.Sprintf(
			"「世界树」\n" +
			"\n" +
			"本次谈话已结束，你已回到大厅。\n" +
			"何不试试发布下一个聊天话题？\n" +
			"如果喜欢的话，请推荐世界树 @WorldTreeBot 给朋友。人多才会好玩哩！\n" +
			"\n" +
			"当前有 %d 人连接到世界树，其中 %d 人在大厅。\n" +
			"若要离开世界树，请戳 /disconnect 。",
			chat + lobby, lobby), msg)
		bot.sendTopicList(user_a,
			"「世界树」\n" +
			"\n" +
			"你对这些话题感兴趣吗？\n" +
			"点击感兴趣的话题即可立刻开始聊天：")

		if user_b != 0 {
			reply := tgbotapi.NewMessage(user_b,
				"「世界树」\n" +
				"\n" +
				"对方结束了本次谈话。\n" +
				"戳 /leave 回到大厅。")
			_, err = bot.api.Send(reply)
			if err != nil {
				bot.replyError(err, msg)
				return
			}
		}
		return
	}

	// Detect whether the user is in queue.
	ok, err = bot.dbm.IsUserInQueue(user_a)
	if err != nil {
		bot.replyError(err, msg)
		return
	}
	if ok {
		err := bot.dbm.JoinLobby(user_a)
		if err != nil {
			bot.replyError(err, msg)
			return
		}
		bot.quickReply(
			"「世界树」\n" +
			"\n" +
			"已取消你最后发布的话题。\n" +
			"若要断开与世界树的连接，请戳 /disconnect 。",
			msg)
		return
	}

	// Detect whether the user is in lobby.
	ok, err = bot.dbm.IsUserInLobby(user_a)
	if err != nil {
		bot.replyError(err, msg)
		return
	}
	if ok {
		bot.quickReply(
			"「世界树」\n" +
			"\n" +
			"你当前没有发布任何话题。\n" +
			"若要断开与世界树的连接，请戳 /disconnect 。",
			msg)
		return
	}

	bot.quickReply(
		"「世界树」\n" +
		"——长夜漫漫，随便找个人，陪你聊到天亮。\n" +
		"\n" +
		"你尚未连接到世界树。\n" +
		"何不戳一下 /start 试试看？",
		msg)
}

func (bot *Bot) handleDisconnect(msg *tgbotapi.Message) {
	user_a := msg.Chat.ID

	// Detect whether the user is in chat.
	ok, err := bot.dbm.IsUserInChat(user_a)
	if err != nil {
		bot.replyError(err, msg)
		return
	}
	if ok {
		bot.quickReply(
			"「世界树」\n" +
			"\n" +
			"你正在一次会话中。\n" +
			"先戳 /leave 离开本次谈话。",
			msg)
		return
	}

	// Detect whether the user is in lobby.
	ok, err = bot.dbm.IsUserInLobby(user_a)
	if err != nil {
		bot.replyError(err, msg)
		return
	}
	if ok {
		err = bot.dbm.LeaveLobby(user_a)
		if err != nil {
			bot.replyError(err, msg)
			return
		}
		bot.quickReply(
			"「世界树」\n" +
			"\n" +
			"你已断开与世界树的连接。\n" +
			"世界树会想念你的，记得常回来逛～\n" +
			"不想聊了就去看看漫画吧： t.cn/RaomgYF\n" +
			"\n" +
			"戳 /start 重新开始。",
			msg)
		return
	}

	bot.quickReply(
		"「世界树」\n" +
		"——长夜漫漫，随便找个人，陪你聊到天亮。\n" +
		"\n" +
		"你尚未连接到世界树。\n" +
		"何不戳一下 /start 试试看？",
		msg)
}

func (bot *Bot) handleMessage(msg *tgbotapi.Message) {
	user_a := msg.Chat.ID

	// Detect whether the user is in chat.
	ok, err := bot.dbm.IsUserInChat(user_a)
	if err != nil {
		bot.replyError(err, msg)
		return
	}
	if ok {
		printLog(msg.From, msg.Text, true)

		user_b, err := bot.dbm.QueryMatch(user_a)
		if err != nil {
			bot.replyError(err, msg)
			return
		}
		if user_b == 0 {
			bot.quickReply(
				"「世界树」\n" +
				"\n" +
				"对方结束了本次谈话，你的消息未送达。\n" +
				"戳 /leave 回到大厅。",
				msg)
			return
		}

		// Forward the message to the partner
		if msg.ForwardFrom != nil || msg.ForwardFromChat != nil {
			fwd := tgbotapi.NewForward(user_b, msg.Chat.ID, msg.MessageID)
			_, err = bot.api.Send(fwd)
			if err != nil {
				bot.replyError(err, msg)
			}
			return
		}
		if msg.ReplyToMessage != nil {
			bot.quickReply(
				"「世界树」\n" +
				"\n" +
				"本服务不保留聊天记录，故无法追踪过去的消息。\n" +
				"由于这个限制，你无法使用定向回复功能。",
				msg)
		}
		if msg.Text != "" {
			fwd := tgbotapi.NewMessage(user_b, msg.Text)
			_, err = bot.api.Send(fwd)
			if err != nil {
				bot.replyError(err, msg)
			}
		}
		if msg.Audio != nil {
			fwd := tgbotapi.NewAudioShare(user_b, msg.Audio.FileID)
			fwd.Caption = msg.Caption
			fwd.Duration = msg.Audio.Duration
			fwd.Performer = msg.Audio.Performer
			fwd.Title = msg.Audio.Title
			_, err = bot.api.Send(fwd)
			if err != nil {
				bot.replyError(err, msg)
			}
		}
		if msg.Document != nil {
			fwd := tgbotapi.NewDocumentShare(user_b, msg.Document.FileID)
			_, err = bot.api.Send(fwd)
			if err != nil {
				bot.replyError(err, msg)
			}
		}
		if msg.Photo != nil {
			if len(*msg.Photo) != 0 {
				fwd := tgbotapi.NewPhotoShare(user_b, (*msg.Photo)[0].FileID)
				fwd.Caption = msg.Caption
				_, err = bot.api.Send(fwd)
				if err != nil {
					bot.replyError(err, msg)
				}
			}
		}
		if msg.Sticker != nil {
			fwd := tgbotapi.NewStickerShare(user_b, msg.Sticker.FileID)
			_, err = bot.api.Send(fwd)
			if err != nil {
				bot.replyError(err, msg)
			}
		}
		if msg.Video != nil {
			fwd := tgbotapi.NewVideoShare(user_b, msg.Video.FileID)
			fwd.Duration = msg.Video.Duration
			fwd.Caption = msg.Caption
			_, err = bot.api.Send(fwd)
			if err != nil {
				bot.replyError(err, msg)
			}
		}
		if msg.Voice != nil {
			fwd := tgbotapi.NewVoiceShare(user_b, msg.Voice.FileID)
			fwd.Caption = msg.Caption
			fwd.Duration = msg.Voice.Duration
			_, err = bot.api.Send(fwd)
			if err != nil {
				bot.replyError(err, msg)
			}
		}
		if msg.Contact != nil {
			fwd := tgbotapi.NewContact(user_b, msg.Contact.PhoneNumber, msg.Contact.FirstName)
			fwd.LastName = msg.Contact.LastName
			_, err = bot.api.Send(fwd)
			if err != nil {
				bot.replyError(err, msg)
			}
		}
		if msg.Location != nil {
			fwd := tgbotapi.NewLocation(user_b, msg.Location.Latitude, msg.Location.Longitude)
			_, err = bot.api.Send(fwd)
			if err != nil {
				bot.replyError(err, msg)
			}
		}
		if msg.Venue != nil {
			fwd := tgbotapi.NewVenue(user_b, msg.Venue.Title, msg.Venue.Address, msg.Venue.Location.Latitude, msg.Venue.Location.Longitude)
			fwd.FoursquareID = msg.Venue.FoursquareID
			_, err = bot.api.Send(fwd)
			if err != nil {
				bot.replyError(err, msg)
			}
		}
		return
	}

	printLog(msg.From, "(lobby) " + msg.Text, false)

	// Detect whether the user is in lobby.
	ok, err = bot.dbm.IsUserInLobby(user_a)
	if err != nil {
		bot.replyError(err, msg)
		return
	}
	if ok {
		topic := strings.TrimSpace(msg.Text)
		short_topic := bot.limitTopic(topic)
		if topic == "" {
			return
		}
		user_b, err := bot.dbm.QueryTopic(short_topic)
		if err != nil {
			bot.replyError(err, msg)
			return
		}
		if user_b == 0 || user_b == user_a {
			if !IsOpenHour(time.Now()) && !DEBUG_MODE {
				bot.quickReply(
					"「世界树」\n" +
					"——长夜漫漫，随便找个人，陪你聊到天亮。\n" +
					"\n" +
					"\u274c " + CLOSED_MSG,
					msg)
				return
			}

			err = bot.dbm.SetTopic(user_a, short_topic)
			if err != nil {
				bot.replyError(err, msg)
				return
			}
			bot.quickReply(
				"「世界树」\n" +
				"\n" +
				"你发布了：" + topic + "\n" +
				"\n" +
				"请等待世界树配对一个点赞的人。\n" +
				"或戳 /list 看看还有哪些话题。",
				msg)
			go bot.broadcastNewTopic(topic, short_topic, user_a)
		} else {
			// Found a match
			err = bot.dbm.LeaveLobby(user_a)
			if err != nil {
				bot.replyError(err, msg)
				return
			}
			err = bot.dbm.LeaveLobby(user_b)
			if err != nil {
				bot.replyError(err, msg)
				return
			}

			bot.quickReply(
				"「世界树」\n" +
				"\n" +
				"你发布了：" + topic + "\n" +
				"\n" +
				"请等待世界树配对一个点赞的人。",
				msg)

			err = bot.dbm.ConnectChat(user_a, user_b)
			if err != nil {
				bot.replyError(err, msg)
				return
			}

			match_ok := "「世界树」\n" +
				"\n" +
				"\U0001f495 会话已接通，祝你们聊天愉快。\n" +
				"话题：" + topic + "\n" +
				"戳 /leave 离开本次谈话。\n" +
				"\n"
			if DEBUG_MODE {
				match_ok += "注：当前程序运行在调试模式下，管理员可能会看到聊天记录。请友善待人，不要分享机密信息。"
			} else {
				match_ok += "注：接下来的聊天内容不会被记录，管理员无法读取，但请友善待人，不要分享机密信息。"
			}
			reply := tgbotapi.NewMessage(user_a, match_ok)
			bot.api.Send(reply)
			reply = tgbotapi.NewMessage(user_b, match_ok)
			_, err = bot.api.Send(reply)
			if err != nil {
				bot.replyError(err, msg)
				return
			}
		}
		return
	}

	bot.quickReply(
		"「世界树」\n" +
		"——长夜漫漫，随便找个人，陪你聊到天亮。\n" +
		"\n" +
		"你尚未连接到世界树。\n" +
		"何不戳一下 /start 试试看？",
		msg)
}

func (bot *Bot) handleWall(msg *tgbotapi.Message) {
	user_a := msg.Chat.ID

	// Detect whether the user is an admininistrator.
	ok, err := bot.dbm.IsUserAnAdmin(user_a)
	if err != nil {
		bot.replyError(err, msg)
		return
	}
	if ok {
		alert := strings.TrimSpace(msg.CommandArguments())
		if alert == "" {
			return
		}
		go func() {
			users, err := bot.dbm.ListAllUsers()
			if err != nil {
				bot.replyError(err, msg)
				return
			}
			for i := range users {
				reply := tgbotapi.NewMessage(users[i],
					"「世界树」\n" +
					"\n" +
					"系统公告：\n" +
					alert)
				bot.api.Send(reply)
			}
		}()
		return
	}

	bot.handleInvalid(msg)
}

func (bot *Bot) handleInvalid(msg *tgbotapi.Message) {
	user_a := msg.Chat.ID

	// Detect whether the user is in chat.
	ok, err := bot.dbm.IsUserInChat(user_a)
	if err != nil {
		bot.replyError(err, msg)
		return
	}
	if ok {
		bot.handleMessage(msg)
		return
	}

	// Detect whether the user is in lobby.
	ok, err = bot.dbm.IsUserInLobby(user_a)
	if err != nil {
		bot.replyError(err, msg)
		return
	}
	if ok {
		bot.quickReply(
			"「世界树」\n" +
			"\n" +
			"你输入了错误的指令。\n" +
			"何不戳一下 /start 试试看？",
			msg)
		return
	}

	bot.quickReply(
		"「世界树」\n" +
		"——长夜漫漫，随便找个人，陪你聊到天亮。\n" +
		"\n" +
		"你输入了错误的指令。\n" +
		"何不戳一下 /start 试试看？",
		msg)
}

func (bot *Bot) handleCallbackQuery(query *tgbotapi.CallbackQuery) {
	msg := query.Message
	if msg == nil || !msg.Chat.IsPrivate() {
		return
	}
	user_a := msg.Chat.ID

	printLog(query.From, "(menu) " + query.Data, false)

	// Detect whether the user is in chat.
	ok, err := bot.dbm.IsUserInChat(user_a)
	if err != nil {
		bot.replyError(err, msg)
		return
	}
	if ok {
		user_b, err := bot.dbm.QueryMatch(user_a)
		if err != nil {
			bot.replyError(err, msg)
			return
		}
		if user_b != 0 {
			bot.quickReply(
				"「世界树」\n" +
				"\n" +
				"你正在一次会话中。\n" +
				"先戳 /leave 离开本次谈话，才能开始下一个会话。",
				msg)
			bot.api.AnswerCallbackQuery(tgbotapi.CallbackConfig {
				CallbackQueryID: query.ID,
			})
			return
		} else {
			err := bot.dbm.DisconnectChat(user_a, 0)
			if err != nil {
				bot.replyError(err, msg)
				return
			}
			err = bot.dbm.JoinLobby(user_a)
			if err != nil {
				bot.replyError(err, msg)
				return
			}
		}
	}

	topic := query.Data
	if topic == "" {
		return
	}
	bot.api.AnswerCallbackQuery(tgbotapi.CallbackConfig {
		CallbackQueryID: query.ID,
		Text: "你 \u2764\ufe0f 了：" + topic,
	})
	user_b, err := bot.dbm.QueryTopic(topic)
	if err != nil {
		bot.replyError(err, msg)
		return
	}
	if user_b == 0 || user_b == user_a {
		// The topic has gone.
		if !IsOpenHour(time.Now()) && !DEBUG_MODE {
			bot.quickReply(
				"「世界树」\n" +
				"——长夜漫漫，随便找个人，陪你聊到天亮。\n" +
				"\n" +
				"\u274c " + CLOSED_MSG,
				msg)
			return
		}

		err = bot.dbm.SetTopic(user_a, topic)
		if err != nil {
			bot.replyError(err, msg)
			return
		}
		bot.quickReply(
			"「世界树」\n" +
			"\n" +
			"你 \u2764\ufe0f 了：" + topic + "\n" +
			"\n" +
			"请等待世界树配对另一个点赞的人。\n" +
			"或戳 /list 看看还有哪些话题。",
			msg)
		if user_b == 0 {
			go bot.broadcastNewTopic(topic, topic, user_a)
		}
	} else {
		err = bot.dbm.LeaveLobby(user_a)
		if err != nil {
			bot.replyError(err, msg)
			return
		}
		err = bot.dbm.LeaveLobby(user_b)
		if err != nil {
			bot.replyError(err, msg)
			return
		}

		bot.quickReply(
			"「世界树」\n" +
			"\n" +
			"你 \u2764\ufe0f 了：" + topic + "\n" +
			"\n" +
			"请等待世界树配对另一个点赞的人。",
			msg)

		err = bot.dbm.ConnectChat(user_a, user_b)
		if err != nil {
			bot.replyError(err, msg)
			return
		}

		match_ok := "「世界树」\n" +
			"\n" +
			"\U0001f495 会话已接通，祝你们聊天愉快。\n" +
			"话题：" + topic + "\n" +
			"戳 /leave 离开本次谈话。\n" +
			"\n"
		if DEBUG_MODE {
			match_ok += "注：当前程序运行在调试模式下，管理员可能会看到聊天记录。请友善待人，不要分享机密信息。"
		} else {
			match_ok += "注：接下来的聊天内容不会被记录，管理员无法读取，但请友善待人，不要分享机密信息。"
		}
		reply := tgbotapi.NewMessage(user_a, match_ok)
		bot.api.Send(reply)
		reply = tgbotapi.NewMessage(user_b, match_ok)
		_, err = bot.api.Send(reply)
		if err != nil {
			bot.replyError(err, msg)
			return
		}
	}
}
