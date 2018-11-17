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
	"strings"
	"time"

	// "gopkg.in/telegram-bot-api.v4"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func (bot *Bot) handleStart(msg *tgbotapi.Message) {
	user_a := msg.Chat.ID

	// Detect whether the user is typing topic.
	ok, err := bot.dbm.IsUserTypingTopic(user_a)
	if err != nil {
		bot.replyError(err, msg, true)
	}
	if ok {
		err = bot.dbm.RemoveInvitation(user_a)
		if err != nil {
			bot.replyError(err, msg, false)
		}
		// fall-through
	}

	// Detect whether the user is in chat.
	ok, err = bot.dbm.IsUserInChat(user_a)
	if err != nil {
		bot.replyError(err, msg, true)
	}
	if ok {
		bot.quickReply(
			"「世界树」\n"+
				"\n"+
				"你正在一对一私聊中。\n"+
				"要继续操作的话，请戳 /leave 回到大厅。",
			msg)
		return
	}

	// Detect whether the user is in lobby.
	ok, err = bot.dbm.IsUserInLobby(user_a)
	if err != nil {
		bot.replyError(err, msg, true)
	}
	if ok {
		chat, lobby, err := bot.dbm.GetActiveUsers()
		if err != nil {
			bot.replyError(err, msg, true)
		}
		user_hash := bot.hashIdentification(msg.Chat)
		bot.quickReply(fmt.Sprintf(
			"欢迎使用「世界树」！\n"+
				"——长夜漫漫，随便找个人，陪你聊到天亮。\n"+
				"\n"+
				"世界树有两种聊天模式：大厅群聊、一对一私聊。\n"+
				"现在正在大厅群聊，你今天的 ID 是 [%s]。\n"+
				"要建立一对一的私聊，请戳 /new 。\n"+
				"\n"+
				"当前有 %d 人连接到世界树，其中 %d 人在大厅。\n"+
				"若要彻底离开世界树，请戳 /disconnect 。\n"+
				"请友善待人，遵守道德和法律。",
			user_hash, chat+lobby, lobby), msg)
		if !IsOpenHour(time.Now()) && !DEBUG_MODE {
			bot.quickReply(
				"「世界树」\n"+
					"\n"+
					"\u274c "+CLOSED_MSG,
				msg)
			return
		}
		bot.sendTopicList(user_a,
			"「世界树」\n"+
				"\n"+
				"以下这些是私聊邀请。\n"+
				"点击感兴趣的话题即可立刻开始私聊：")
		return
	}

	err = bot.dbm.JoinLobby(user_a, 0) // TODO: more lobbies
	if err != nil {
		bot.replyError(err, msg, true)
	}
	chat, lobby, err := bot.dbm.GetActiveUsers()
	if err != nil {
		bot.replyError(err, msg, true)
	}
	user_hash := bot.hashIdentification(msg.Chat)
	bot.quickReply(fmt.Sprintf(
		"欢迎使用「世界树」！\n"+
			"——长夜漫漫，随便找个人，陪你聊到天亮。\n"+
			"\n"+
			"世界树有两种聊天模式：大厅群聊、一对一私聊。\n"+
			"现在正在大厅群聊，你今天的 ID 是 [%s]。\n"+
			"要建立一对一的私聊，请戳 /new 。\n"+
			"\n"+
			"当前有 %d 人连接到世界树，其中 %d 人在大厅。\n"+
			"若要彻底离开世界树，请戳 /disconnect 。\n"+
			"请友善待人，遵守道德和法律。",
		user_hash, chat+lobby, lobby), msg)
	if !IsOpenHour(time.Now()) && !DEBUG_MODE {
		bot.quickReply(
			"「世界树」\n"+
				"\n"+
				"\u274c "+CLOSED_MSG,
			msg)
		return
	}
	bot.sendTopicList(user_a,
		"「世界树」\n"+
			"\n"+
			"以下这些是私聊邀请。\n"+
			"点击感兴趣的话题即可立刻开始私聊：")
}

func (bot *Bot) handleNew(msg *tgbotapi.Message) {
	user_a := msg.Chat.ID
	user_a_nick := bot.hashIdentification(msg.Chat)

	// Detect whether the user is typing topic.
	ok, err := bot.dbm.IsUserTypingTopic(user_a)
	if err != nil {
		bot.replyError(err, msg, true)
	}
	if ok {
		err = bot.dbm.RemoveInvitation(user_a)
		if err != nil {
			bot.replyError(err, msg, false)
		}
		// fall-through
	}

	// Detect whether the user is in chat.
	ok, err = bot.dbm.IsUserInChat(user_a)
	if err != nil {
		bot.replyError(err, msg, true)
	}
	if ok {
		bot.quickReply(
			"「世界树」\n"+
				"\n"+
				"你正在一对一私聊中。\n"+
				"要继续操作的话，请戳 /leave 回到大厅。",
			msg)
		return
	}

	// Detect whether the user is in lobby.
	ok, err = bot.dbm.IsUserInLobby(user_a)
	if err != nil {
		bot.replyError(err, msg, true)
	}
	if ok {
		topic := strings.TrimSpace(msg.CommandArguments())
		if topic != "" {
			short_topic := bot.limitTopic(topic)
			bot.respondTopic(topic, short_topic, user_a, user_a_nick,
				"「世界树」\n"+
					"\n"+
					"你发布了：%s",
				"「世界树」\n"+
					"\n"+
					"你发布了：%s\n"+
					"\n"+
					"请等待有人回应你。\n"+
					"或戳 /list 看看还有哪些别的话题。",
				msg)
		} else {
			if !IsOpenHour(time.Now()) && !DEBUG_MODE {
				bot.quickReply(
					"「世界树」\n"+
						"\n"+
						"\u274c "+CLOSED_MSG,
					msg)
				return
			}
			err = bot.dbm.NewPendingInvitation(user_a)
			if err != nil {
				bot.replyError(err, msg, true)
			}
			bot.askReply(
				"「世界树」\n"+
					"\n"+
					"接下来，请输入一句话题，等待有人回应：",
				msg)
		}
		return
	}

	bot.quickReply(
		"「世界树」\n"+
			"——长夜漫漫，随便找个人，陪你聊到天亮。\n"+
			"\n"+
			"你尚未连接到世界树。\n"+
			"何不戳一下 /start 试试看？",
		msg)
}

func (bot *Bot) handleNick(msg *tgbotapi.Message) {
	user_a := msg.Chat.ID

	// Detect whether the user is typing topic.
	ok, err := bot.dbm.IsUserTypingTopic(user_a)
	if err != nil {
		bot.replyError(err, msg, true)
	}
	if ok {
		err = bot.dbm.RemoveInvitation(user_a)
		if err != nil {
			bot.replyError(err, msg, false)
		}
		// fall-through
	}

	// Detect whether the user is in chat.
	ok, err = bot.dbm.IsUserInChat(user_a)
	if err != nil {
		bot.replyError(err, msg, true)
	}
	if ok {
		user_hash := bot.hashIdentification(msg.Chat)
		bot.quickReply(fmt.Sprintf(
			"「世界树」\n"+
				"\n"+
				"你今天在大厅的 ID 是 [%s]\n"+
				"北京时间 3:00 会自动更新。",
			user_hash), msg)
		return
	}

	// Detect whether the user is in lobby.
	ok, err = bot.dbm.IsUserInLobby(user_a)
	if err != nil {
		bot.replyError(err, msg, true)
	}
	if ok {
		user_hash := bot.hashIdentification(msg.Chat)
		bot.quickReply(fmt.Sprintf(
			"「世界树」\n"+
				"\n"+
				"你今天在大厅的 ID 是 [%s]\n"+
				"北京时间 3:00 会自动更新。",
			user_hash), msg)
		return
	}

	bot.quickReply(
		"「世界树」\n"+
			"——长夜漫漫，随便找个人，陪你聊到天亮。\n"+
			"\n"+
			"你尚未连接到世界树。\n"+
			"何不戳一下 /start 试试看？",
		msg)
}

func (bot *Bot) handleList(msg *tgbotapi.Message) {
	user_a := msg.Chat.ID

	// Detect whether the user is typing topic.
	ok, err := bot.dbm.IsUserTypingTopic(user_a)
	if err != nil {
		bot.replyError(err, msg, true)
	}
	if ok {
		err = bot.dbm.RemoveInvitation(user_a)
		if err != nil {
			bot.replyError(err, msg, false)
		}
		// fall-through
	}

	// Detect whether the user is in chat.
	ok, err = bot.dbm.IsUserInChat(user_a)
	if err != nil {
		bot.replyError(err, msg, true)
	}
	if ok {
		chat, lobby, err := bot.dbm.GetActiveUsers()
		if err != nil {
			bot.replyError(err, msg, true)
		}
		num_topics, err := bot.sendTopicList(user_a, fmt.Sprintf(
			"「世界树」\n"+
				"\n"+
				"当前有 %d 人连接到世界树，其中 %d 人在大厅。\n"+
				"以下这些是私聊邀请：",
			chat+lobby, lobby))
		if err != nil {
			bot.replyError(err, msg, true)
		}
		if num_topics == 0 {
			bot.quickReply(fmt.Sprintf(
				"「世界树」\n"+
					"\n"+
					"当前有 %d 人连接到世界树，其中 %d 人在大厅。\n"+
					"当前没有私聊邀请。",
				chat+lobby, lobby), msg)
		}
		return
	}

	// Detect whether the user is in lobby.
	ok, err = bot.dbm.IsUserInLobby(user_a)
	if err != nil {
		bot.replyError(err, msg, true)
	}
	if ok {
		chat, lobby, err := bot.dbm.GetActiveUsers()
		if err != nil {
			bot.replyError(err, msg, true)
		}
		num_topics, err := bot.sendTopicList(user_a, fmt.Sprintf(
			"「世界树」\n"+
				"\n"+
				"当前有 %d 人连接到世界树，其中 %d 人在大厅。\n"+
				"以下这些是私聊邀请，\n"+
				"点击感兴趣的话题即可立刻开始私聊。\n"+
				"\n"+
				"你也可以输入 /new 来自定义话题。",
			chat+lobby, lobby))
		if err != nil {
			bot.replyError(err, msg, true)
		}
		if num_topics == 0 {
			bot.quickReply(fmt.Sprintf(
				"「世界树」\n"+
					"\n"+
					"当前有 %d 人连接到世界树，其中 %d 人在大厅。\n"+
					"当前没有私聊邀请。\n"+
					"何不输入 /new 来发布一个呢？",
				chat+lobby, lobby), msg)
		}
		return
	}

	bot.quickReply(
		"「世界树」\n"+
			"——长夜漫漫，随便找个人，陪你聊到天亮。\n"+
			"\n"+
			"你尚未连接到世界树。\n"+
			"何不戳一下 /start 试试看？",
		msg)
}

func (bot *Bot) handleLeave(msg *tgbotapi.Message) {
	user_a := msg.Chat.ID

	// Detect whether the user is typing topic.
	ok, err := bot.dbm.IsUserTypingTopic(user_a)
	if err != nil {
		bot.replyError(err, msg, true)
	}
	if ok {
		err = bot.dbm.RemoveInvitation(user_a)
		if err != nil {
			bot.replyError(err, msg, false)
		}
		bot.quickReply(
			"「世界树」\n"+
				"\n"+
				"已撤销你发布的私聊邀请，\n"+
				"并回到了大厅。",
			msg)
		return
	}

	// Detect whether the user is in chat.
	ok, err = bot.dbm.IsUserInChat(user_a)
	if err != nil {
		bot.replyError(err, msg, true)
	}
	if ok {
		user_b, err := bot.dbm.QueryChat(user_a)
		if err != nil {
			bot.replyError(err, msg, false)
		}
		err = bot.dbm.DisconnectChat(user_a, user_b)
		if err != nil {
			bot.replyError(err, msg, false)
		}

		err = bot.dbm.JoinLobby(user_a, 0) // TODO: more lobbies
		if err != nil {
			bot.replyError(err, msg, true)
		}

		chat, lobby, err := bot.dbm.GetActiveUsers()
		if err != nil {
			bot.replyError(err, msg, true)
		}
		bot.quickReply(fmt.Sprintf(
			"「世界树」\n"+
				"\n"+
				"本次私聊已结束，你已回到大厅。\n"+
				"如果喜欢的话，请推荐世界树 @WorldTreeBot 给朋友。人多才会好玩哩！\n"+
				"\n"+
				"当前有 %d 人连接到世界树，其中 %d 人在大厅。",
			chat+lobby, lobby), msg)
		bot.sendTopicList(user_a,
			"「世界树」\n"+
				"\n"+
				"以下这些是其它私聊邀请。\n"+
				"点击感兴趣的话题即可立刻开始私聊：")

		if user_b != 0 {
			reply := tgbotapi.NewMessage(user_b,
				"「世界树」\n"+
					"\n"+
					"对方结束了本次私聊。\n"+
					"戳 /leave 回到大厅。")
			_, err = bot.api.Send(reply)
			if err != nil {
				bot.replyError(err, msg, true)
			}
		}
		return
	}

	// Detect whether the user is in queue.
	ok, err = bot.dbm.IsUserInQueue(user_a)
	if err != nil {
		bot.replyError(err, msg, true)
	}
	if ok {
		err := bot.dbm.RemoveInvitation(user_a)
		if err != nil {
			bot.replyError(err, msg, true)
		}
		bot.quickReply(
			"「世界树」\n"+
				"\n"+
				"已撤销你发布的私聊邀请，\n"+
				"并回到了大厅。",
			msg)
		return
	}

	// Detect whether the user is in lobby.
	ok, err = bot.dbm.IsUserInLobby(user_a)
	if err != nil {
		bot.replyError(err, msg, true)
	}
	if ok {
		bot.quickReply(
			"「世界树」\n"+
				"\n"+
				"你已经在大厅了,和大家一起聊天呗。\n"+
				"不过，若要彻底离开世界树，请戳 /disconnect 。",
			msg)
		return
	}

	bot.quickReply(
		"「世界树」\n"+
			"——长夜漫漫，随便找个人，陪你聊到天亮。\n"+
			"\n"+
			"你尚未连接到世界树。\n"+
			"何不戳一下 /start 试试看？",
		msg)
}

func (bot *Bot) handleDisconnect(msg *tgbotapi.Message) {
	user_a := msg.Chat.ID

	// Detect whether the user is typing topic.
	ok, err := bot.dbm.IsUserTypingTopic(user_a)
	if err != nil {
		bot.replyError(err, msg, true)
	}
	if ok {
		err = bot.dbm.RemoveInvitation(user_a)
		if err != nil {
			bot.replyError(err, msg, false)
		}
		// fall-through
	}

	// Detect whether the user is in chat.
	ok, err = bot.dbm.IsUserInChat(user_a)
	if err != nil {
		bot.replyError(err, msg, true)
	}
	if ok {
		bot.quickReply(
			"「世界树」\n"+
				"\n"+
				"你正在一对一私聊中。\n"+
				"要继续操作的话，请戳 /leave 回到大厅。",
			msg)
		return
	}

	// Detect whether the user is in lobby.
	ok, err = bot.dbm.IsUserInLobby(user_a)
	if err != nil {
		bot.replyError(err, msg, true)
	}
	if ok {
		err = bot.dbm.RemoveInvitation(user_a)
		if err != nil {
			bot.replyError(err, msg, false)
		}
		err = bot.dbm.LeaveLobby(user_a)
		if err != nil {
			bot.replyError(err, msg, false)
		}
		bot.quickReply(
			"「世界树」\n"+
				"\n"+
				"你已断开与世界树的连接。\n"+
				"世界树会想念你的，记得常回来逛～\n"+
				"不想聊了就去看看漫画吧： t.cn/RaomgYF\n"+
				"\n"+
				"戳 /start 重新开始。",
			msg)
		return
	}

	bot.quickReply(
		"「世界树」\n"+
			"——长夜漫漫，随便找个人，陪你聊到天亮。\n"+
			"\n"+
			"你尚未连接到世界树。\n"+
			"何不戳一下 /start 试试看？",
		msg)
}

func (bot *Bot) handleMessage(msg *tgbotapi.Message) {
	user_a := msg.Chat.ID
	user_a_nick := bot.hashIdentification(msg.Chat)

	// Detect whether the user is typing topic.
	ok, err := bot.dbm.IsUserTypingTopic(user_a)
	if err != nil {
		bot.replyError(err, msg, true)
	}
	if ok {
		printLog(msg.From, "(topic) "+msg.Text, false)
		topic := strings.TrimSpace(msg.Text)
		if topic == "" {
			bot.askReply(
				"「世界树」\n"+
					"\n"+
					"接下来，请发布一句话题，等待有人回应：",
				msg)
			return
		}
		err = bot.dbm.RemoveInvitation(user_a)
		if err != nil {
			bot.replyError(err, msg, true)
		}
		short_topic := bot.limitTopic(topic)
		bot.respondTopic(topic, short_topic, user_a, user_a_nick,
			"「世界树」\n"+
				"\n"+
				"你发布了：%s",
			"「世界树」\n"+
				"\n"+
				"你发布了：%s\n"+
				"\n"+
				"请等待有人回应你。\n"+
				"或戳 /list 看看还有哪些别的话题。",
			msg)
		return
	}

	// Detect whether the user is in chat.
	ok, err = bot.dbm.IsUserInChat(user_a)
	if err != nil {
		bot.replyError(err, msg, true)
	}
	if ok {
		printLog(msg.From, msg.Text, true)

		user_b, err := bot.dbm.QueryChat(user_a)
		if err != nil {
			bot.replyError(err, msg, true)
		}
		if user_b == 0 {
			bot.quickReply(
				"「世界树」\n"+
					"\n"+
					"对方提前结束了本次私聊，你的消息未送达。\n"+
					"要继续操作的话，请戳 /leave 回到大厅。",
				msg)
			return
		}

		if msg.ReplyToMessage != nil && msg.ForwardFrom == nil && msg.ForwardFromChat == nil {
			bot.quickReply(
				"「世界树」\n"+
					"\n"+
					"本服务不保留聊天记录，故无法追踪过去的消息。\n"+
					"由于这个限制，你无法使用定向回复功能。十分抱歉。",
				msg)
		}

		// Forward the message to the partner
		replies := make([]tgbotapi.Chattable, 0, 2)
		replies = bot.generateForwardMessage(replies, user_b, "", msg, false)
		bot.queue.Send(QUEUE_PRIORITY_NORMAL, replies, func(msg_result []*tgbotapi.Message, msg_errors []error) {
			bot.replyError(msg_errors[0], msg, false)
		})
		return
	}

	// Detect whether the user is in lobby.
	ok, err = bot.dbm.IsUserInLobby(user_a)
	if err != nil {
		bot.replyError(err, msg, true)
	}
	if ok {
		printLog(msg.From, "(lobby) "+msg.Text, false)

		if !IsOpenHour(time.Now()) && !DEBUG_MODE {
			bot.quickReply(
				"「世界树」\n"+
					"——长夜漫漫，随便找个人，陪你聊到天亮。\n"+
					"\n"+
					"\u274c "+CLOSED_MSG,
				msg)
			return
		}

		room, err := bot.dbm.QueryLobby(user_a)
		if err != nil {
			bot.replyError(err, msg, true)
		}

		users, err := bot.dbm.ListUsersInLobby(room)
		if err != nil {
			bot.replyError(err, msg, true)
		}

		if msg.ReplyToMessage != nil && msg.ForwardFrom == nil && msg.ForwardFromChat == nil {
			bot.quickReply(
				"「世界树」\n"+
					"\n"+
					"本服务不保留聊天记录，故无法追踪过去的消息。\n"+
					"由于这个限制，你无法使用定向回复功能。十分抱歉。",
				msg)
		}

		// Forward the message to all users in the lobby
		replies := make([]tgbotapi.Chattable, 0, len(users)*2)
		for i := range users {
			if users[i] == user_a {
				continue
			}
			replies = bot.generateForwardMessage(replies, users[i], user_a_nick, msg, true)
		}
		bot.queue.Send(QUEUE_PRIORITY_LOW, replies, func(msg_result []*tgbotapi.Message, msg_errors []error) {
			bot.logBroadcastResult(msg_errors, msg)
		})
		return
	}

	printLog(msg.From, "(disconnected) "+msg.Text, false)

	bot.quickReply(
		"「世界树」\n"+
			"——长夜漫漫，随便找个人，陪你聊到天亮。\n"+
			"\n"+
			"你尚未连接到世界树。\n"+
			"何不戳一下 /start 试试看？",
		msg)
}

func (bot *Bot) handleWall(msg *tgbotapi.Message) {
	user_a := msg.Chat.ID

	// Detect whether the user is typing topic.
	ok, err := bot.dbm.IsUserTypingTopic(user_a)
	if err != nil {
		bot.replyError(err, msg, true)
	}
	if ok {
		err = bot.dbm.RemoveInvitation(user_a)
		if err != nil {
			bot.replyError(err, msg, false)
		}
		// fall-through
	}

	// Detect whether the user is an admininistrator.
	ok, err = bot.dbm.IsUserAnAdmin(user_a)
	if err != nil {
		bot.replyError(err, msg, true)
	}
	if ok {
		alert := strings.TrimSpace(msg.CommandArguments())
		if alert == "" {
			return
		}
		users, err := bot.dbm.ListAllUsers()
		if err != nil {
			bot.replyError(err, msg, true)
		}
		replies := make([]tgbotapi.Chattable, 0, len(users))
		for i := range users {
			reply := tgbotapi.NewMessage(users[i],
				"【系统公告】\n"+
					"\n"+
					alert)
			replies = append(replies, reply)
		}
		bot.queue.Send(QUEUE_PRIORITY_HIGH, replies, func(msg_result []*tgbotapi.Message, msg_errors []error) {
			bot.sendBroadcastResult(msg_errors, msg)
		})
		return
	}

	bot.handleInvalid(msg)
}

func (bot *Bot) handleInvalid(msg *tgbotapi.Message) {
	user_a := msg.Chat.ID

	// Detect whether the user is in chat.
	ok, err := bot.dbm.IsUserInChat(user_a)
	if err != nil {
		bot.replyError(err, msg, true)
	}
	if ok {
		bot.handleMessage(msg)
		return
	}

	// Detect whether the user is in lobby.
	ok, err = bot.dbm.IsUserInLobby(user_a)
	if err != nil {
		bot.replyError(err, msg, true)
	}
	if ok {
		bot.quickReply(
			"「世界树」\n"+
				"\n"+
				"你输入了错误的指令。\n"+
				"何不戳一下 /start 试试看？",
			msg)
		return
	}

	bot.quickReply(
		"「世界树」\n"+
			"——长夜漫漫，随便找个人，陪你聊到天亮。\n"+
			"\n"+
			"你输入了错误的指令。\n"+
			"何不戳一下 /start 试试看？",
		msg)
}

func (bot *Bot) handleCallbackQuery(query *tgbotapi.CallbackQuery) {
	msg := query.Message
	if msg == nil || !msg.Chat.IsPrivate() {
		return
	}
	user_a := msg.Chat.ID
	user_a_nick := bot.hashIdentification(msg.Chat)

	printLog(query.From, "(menu) "+query.Data, false)

	topic := query.Data
	if topic == "" {
		return
	}
	bot.api.AnswerCallbackQuery(tgbotapi.CallbackConfig{
		CallbackQueryID: query.ID,
		Text:            "正在加入：" + topic,
	})

	// Detect whether the user is typing topic.
	ok, err := bot.dbm.IsUserTypingTopic(user_a)
	if err != nil {
		bot.replyError(err, msg, true)
	}
	if ok {
		err = bot.dbm.RemoveInvitation(user_a)
		if err != nil {
			bot.replyError(err, msg, false)
		}
		// fall-through
	}

	// Detect whether the user is in chat.
	ok, err = bot.dbm.IsUserInChat(user_a)
	if err != nil {
		bot.replyError(err, msg, true)
	}
	if ok {
		user_b, err := bot.dbm.QueryChat(user_a)
		if err != nil {
			bot.replyError(err, msg, true)
		}
		if user_b != 0 {
			bot.quickReply(
				"「世界树」\n"+
					"\n"+
					"你正在一对一私聊中。\n"+
					"要继续操作的话，请戳 /leave 回到大厅。",
				msg)
			return
		} else {
			err := bot.dbm.DisconnectChat(user_a, 0)
			if err != nil {
				bot.replyError(err, msg, true)
			}
			err = bot.dbm.JoinLobby(user_a, 0) // TODO: more lobbies
			if err != nil {
				bot.replyError(err, msg, true)
			}
			// fall-through
		}
	}

	bot.respondTopic(topic, topic, user_a, user_a_nick,
		"「世界树」\n"+
			"\n"+
			"正在加入话题：%s",
		"「世界树」\n"+
			"\n"+
			"正在加入话题：%s\n"+
			"\n"+
			"请等待有人回应你。\n"+
			"或戳 /list 看看还有哪些别的话题。",
		msg)
}
