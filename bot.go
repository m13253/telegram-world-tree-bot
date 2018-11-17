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
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"time"

	// "gopkg.in/telegram-bot-api.v4"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type Bot struct {
	api     *tgbotapi.BotAPI
	dbm     *dbManager
	queue   *sendQueue
	updates <-chan tgbotapi.Update
}

func NewBot(api *tgbotapi.BotAPI, dbm *dbManager) (bot *Bot, err error) {
	bot = &Bot{
		api:   api,
		dbm:   dbm,
		queue: NewSendQueue(api, dbm),
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	bot.updates, err = api.GetUpdatesChan(u)
	if err != nil {
		return
	}

	return
}

func (bot *Bot) Run() {
	for update := range bot.updates {
		bot.processUpdate(&update)
	}
}

func (bot *Bot) processUpdate(update *tgbotapi.Update) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Fatal: %+v\n", r)
		}
	}()

	msg := update.Message
	if msg != nil && msg.Chat.IsPrivate() {

		ok, err := bot.dbm.IsUserInBanList(msg.Chat.ID)
		if err != nil {
			bot.replyError(err, msg, true)
		}
		if ok {
			bot.quickReply(`你已被拉黑`, msg)
			return
		}

		if strings.HasPrefix(msg.Text, "/") {
			printLog(msg.From, msg.Text, false)
		}

		cmd := msg.Command()
		if cmd == "" {
			bot.handleMessage(msg)
		} else if cmd == "start" {
			bot.handleStart(msg)
		} else if cmd == "new" {
			bot.handleNew(msg)
		} else if cmd == "nick" {
			bot.handleNick(msg)
		} else if cmd == "list" {
			bot.handleList(msg)
		} else if cmd == "leave" {
			bot.handleLeave(msg)
		} else if cmd == "disconnect" {
			bot.handleDisconnect(msg)
		} else if cmd == "wall" {
			bot.handleWall(msg)
		} else {
			bot.handleInvalid(msg)
		}

	}

	edit_msg := update.EditedMessage
	if edit_msg != nil && edit_msg.Chat.IsPrivate() {
		bot.quickReply(
			"「世界树」\n"+
				"\n"+
				"本服务不保留聊天记录，故无法追踪消息编辑状态。\n"+
				"由于这个限制，你无法使用消息编辑功能。十分抱歉。",
			edit_msg)
	}

	query := update.CallbackQuery
	if query != nil {
		bot.handleCallbackQuery(query)
	}
}

func (bot *Bot) quickReply(text string, msg *tgbotapi.Message) {
	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	if msg != nil {
		reply.ReplyToMessageID = msg.MessageID
	}
	reply.ReplyMarkup = tgbotapi.ForceReply{
		ForceReply: false,
	}
	reply.DisableWebPagePreview = true
	bot.queue.Send(QUEUE_PRIORITY_HIGH, []tgbotapi.Chattable{reply}, nil)
}

func (bot *Bot) askReply(text string, msg *tgbotapi.Message) {
	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	if msg != nil {
		reply.ReplyToMessageID = msg.MessageID
	}
	reply.ReplyMarkup = tgbotapi.ForceReply{
		ForceReply: true,
		Selective:  true,
	}
	reply.DisableWebPagePreview = true
	bot.queue.Send(QUEUE_PRIORITY_HIGH, []tgbotapi.Chattable{reply}, nil)
}

func (bot *Bot) generateForwardMessage(existing_replies []tgbotapi.Chattable, dest int64, nick string, msg *tgbotapi.Message, disable_notification bool) []tgbotapi.Chattable {
	num_existing_replies := len(existing_replies)
	has_nick := nick != ""
	if msg.ForwardFrom != nil || msg.ForwardFromChat != nil {
		if has_nick {
			fwd_nick := tgbotapi.NewMessage(dest, "["+nick+"]")
			fwd_nick.DisableNotification = disable_notification
			existing_replies = append(existing_replies, fwd_nick)
		}
		fwd := tgbotapi.NewForward(dest, msg.Chat.ID, msg.MessageID)
		fwd.DisableNotification = disable_notification
		existing_replies = append(existing_replies, fwd)
		return existing_replies
	}
	if msg.Text != "" {
		var text string
		if has_nick {
			text = "[" + nick + "] " + msg.Text
		} else {
			text = msg.Text
		}
		fwd := tgbotapi.NewMessage(dest, text)
		fwd.DisableNotification = disable_notification
		existing_replies = append(existing_replies, fwd)
	}
	if msg.Audio != nil {
		if has_nick {
			fwd_nick := tgbotapi.NewMessage(dest, "["+nick+"]")
			fwd_nick.DisableNotification = disable_notification
			existing_replies = append(existing_replies, fwd_nick)
		}
		fwd := tgbotapi.NewAudioShare(dest, msg.Audio.FileID)
		fwd.DisableNotification = disable_notification
		fwd.Caption = msg.Caption
		fwd.Duration = msg.Audio.Duration
		fwd.Performer = msg.Audio.Performer
		fwd.Title = msg.Audio.Title
		existing_replies = append(existing_replies, fwd)
	}
	if msg.Document != nil {
		var text string
		if has_nick {
			text = "[" + nick + "] " + msg.Caption
		} else {
			text = msg.Caption
		}
		fwd := tgbotapi.NewDocumentShare(dest, msg.Document.FileID)
		fwd.DisableNotification = disable_notification
		fwd.Caption = text
		existing_replies = append(existing_replies, fwd)
	}
	if msg.Photo != nil {
		var text string
		if has_nick {
			text = "[" + nick + "] " + msg.Caption
		} else {
			text = msg.Caption
		}
		if len(*msg.Photo) != 0 {
			fwd := tgbotapi.NewPhotoShare(dest, (*msg.Photo)[0].FileID)
			fwd.DisableNotification = disable_notification
			fwd.Caption = text
			existing_replies = append(existing_replies, fwd)
		}
	}
	if msg.Sticker != nil {
		if has_nick {
			fwd_nick := tgbotapi.NewMessage(dest, "["+nick+"]")
			fwd_nick.DisableNotification = disable_notification
			existing_replies = append(existing_replies, fwd_nick)
		}
		fwd := tgbotapi.NewStickerShare(dest, msg.Sticker.FileID)
		fwd.DisableNotification = disable_notification
		existing_replies = append(existing_replies, fwd)
	}
	if msg.Video != nil {
		var text string
		if has_nick {
			text = "[" + nick + "] " + msg.Caption
		} else {
			text = msg.Caption
		}
		fwd := tgbotapi.NewVideoShare(dest, msg.Video.FileID)
		fwd.DisableNotification = disable_notification
		fwd.Duration = msg.Video.Duration
		fwd.Caption = text
		existing_replies = append(existing_replies, fwd)
	}
	if msg.VideoNote != nil {
		if has_nick {
			fwd_nick := tgbotapi.NewMessage(dest, "["+nick+"]")
			fwd_nick.DisableNotification = disable_notification
			existing_replies = append(existing_replies, fwd_nick)
		}
		fwd := tgbotapi.NewVideoNoteShare(dest, msg.VideoNote.Length, msg.VideoNote.FileID)
		fwd.DisableNotification = disable_notification
		fwd.Duration = msg.VideoNote.Duration
		existing_replies = append(existing_replies, fwd)
	}
	if msg.Voice != nil {
		var text string
		if has_nick {
			text = "[" + nick + "] " + msg.Caption
		} else {
			text = msg.Caption
		}
		fwd := tgbotapi.NewVoiceShare(dest, msg.Voice.FileID)
		fwd.DisableNotification = disable_notification
		fwd.Caption = text
		fwd.Duration = msg.Voice.Duration
		existing_replies = append(existing_replies, fwd)
	}
	if msg.Contact != nil {
		if has_nick {
			fwd_nick := tgbotapi.NewMessage(dest, "["+nick+"]")
			fwd_nick.DisableNotification = disable_notification
			existing_replies = append(existing_replies, fwd_nick)
		}
		fwd := tgbotapi.NewContact(dest, msg.Contact.PhoneNumber, msg.Contact.FirstName)
		fwd.DisableNotification = disable_notification
		fwd.LastName = msg.Contact.LastName
		existing_replies = append(existing_replies, fwd)
	}
	if msg.Location != nil {
		if has_nick {
			fwd_nick := tgbotapi.NewMessage(dest, "["+nick+"]")
			fwd_nick.DisableNotification = disable_notification
			existing_replies = append(existing_replies, fwd_nick)
		}
		fwd := tgbotapi.NewLocation(dest, msg.Location.Latitude, msg.Location.Longitude)
		fwd.DisableNotification = disable_notification
		existing_replies = append(existing_replies, fwd)
	}
	if msg.Venue != nil {
		var text string
		if has_nick {
			text = "[" + nick + "] " + msg.Venue.Title
		} else {
			text = msg.Venue.Title
		}
		fwd := tgbotapi.NewVenue(dest, text, msg.Venue.Address, msg.Venue.Location.Latitude, msg.Venue.Location.Longitude)
		fwd.DisableNotification = disable_notification
		fwd.FoursquareID = msg.Venue.FoursquareID
		existing_replies = append(existing_replies, fwd)
	}
	if num_existing_replies == len(existing_replies) {
		var text string
		if has_nick {
			text = "[" + nick + "] [不支持的消息]"
		} else {
			text = "[不支持的消息]"
		}
		fwd := tgbotapi.NewMessage(dest, text)
		fwd.DisableNotification = disable_notification
		existing_replies = append(existing_replies, fwd)
	}
	return existing_replies
}

func (bot *Bot) sendBroadcastResult(msg_errors []error, msg *tgbotapi.Message) {
	success := 0
	failure := 0
	for i := range msg_errors {
		if msg_errors[i] == nil {
			success++
		} else {
			failure++
		}
	}
	log.Printf("Sent / failed: %d / %d", success, failure)
	var text string
	if failure == 0 {
		text = fmt.Sprintf("\u2705送达：%d", success)
	} else {
		text = fmt.Sprintf("\u2705送达：%d，\u2716失败：%d", success, failure)
	}
	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	reply.DisableNotification = true
	reply.ReplyToMessageID = msg.MessageID
	bot.queue.Send(QUEUE_PRIORITY_HIGH, []tgbotapi.Chattable{reply}, nil)
}

func (bot *Bot) logBroadcastResult(msg_errors []error, msg *tgbotapi.Message) {
	success := 0
	failure := 0
	for i := range msg_errors {
		if msg_errors[i] == nil {
			success++
		} else {
			failure++
		}
	}
	log.Printf("Sent / failed: %d / %d", success, failure)
}

func (bot *Bot) sendTopicList(user int64, caption string) (count int, err error) {
	topics, err := bot.dbm.ListInvites()
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
		keyboard[i] = []tgbotapi.InlineKeyboardButton{
			tgbotapi.InlineKeyboardButton{
				Text:         topics[i],
				CallbackData: &topics[i],
			},
		}
	}
	reply.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
	bot.queue.Send(QUEUE_PRIORITY_HIGH, []tgbotapi.Chattable{reply}, nil)
	return
}

func (bot *Bot) respondTopic(topic string, short_topic string, user_a int64, user_a_nick string, success_text string, wait_text string, msg *tgbotapi.Message) {
	user_b, err := bot.dbm.QueryInvitation(short_topic)
	if err != nil {
		bot.replyError(err, msg, true)
	}
	if user_b == 0 || user_b == user_a {
		// The topic has gone.
		if !IsOpenHour(time.Now()) && !DEBUG_MODE {
			bot.quickReply(
				"「世界树」\n"+
					"——长夜漫漫，随便找个人，陪你聊到天亮。\n"+
					"\n"+
					"\u274c "+CLOSED_MSG,
				msg)
			return
		}

		err = bot.dbm.NewInvitation(user_a, short_topic)
		if err != nil {
			bot.replyError(err, msg, true)
		}
		bot.quickReply(fmt.Sprintf(wait_text, topic), msg)
		if user_b == 0 {
			err = bot.broadcastInvitation(topic, topic, user_a, user_a_nick)
			if err != nil {
				bot.replyError(err, msg, true)
			}
		}
	} else {
		err = bot.dbm.RemoveInvitationByTopic(topic)
		if err != nil {
			bot.replyError(err, msg, true)
		}
		err = bot.dbm.LeaveLobby(user_a)
		if err != nil {
			bot.replyError(err, msg, true)
		}
		err = bot.dbm.LeaveLobby(user_b)
		if err != nil {
			bot.replyError(err, msg, true)
		}

		bot.quickReply(fmt.Sprintf(success_text, topic), msg)

		err = bot.dbm.ConnectChat(user_a, user_b)
		if err != nil {
			bot.replyError(err, msg, true)
		}

		text := "「世界树」\n" +
			"\n" +
			"\U0001f495 会话已接通，祝你们聊天愉快。\n" +
			"\n" +
			"话题：" + topic + "\n" +
			"戳 /leave 离开本次谈话。\n" +
			"\n"
		if DEBUG_MODE {
			text += "注：当前程序运行在调试模式下，管理员可能会看到聊天记录。请友善待人，不要分享机密信息。"
		} else {
			text += "注：接下来的聊天内容不会被记录，管理员无法读取，但请友善待人，不要分享机密信息。"
		}
		bot.queue.Send(QUEUE_PRIORITY_HIGH, []tgbotapi.Chattable{
			tgbotapi.NewMessage(user_a, text),
			tgbotapi.NewMessage(user_b, text),
		}, nil)

		err = bot.broadcastMatch(topic, user_a, user_b, user_a_nick)
		if err != nil {
			bot.replyError(err, msg, true)
		}
	}
}

func (bot *Bot) broadcastInvitation(topic string, short_topic string, exclude_user int64, nick string) error {
	users, err := bot.dbm.ListUnmatchedUsers()
	if err != nil {
		return err
	}
	reply_markup := tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{
			tgbotapi.InlineKeyboardButton{
				Text:         "\u2764\ufe0f 加入",
				CallbackData: &short_topic,
			},
		})
	replies := make([]tgbotapi.Chattable, 0, len(users))
	for i := range users {
		if users[i] == exclude_user {
			continue
		}
		reply := tgbotapi.NewMessage(users[i],
			"【新私聊邀请】\n"+
				"\n"+
				"["+nick+"] "+topic)
		reply.ReplyMarkup = reply_markup
		reply.DisableNotification = true
		replies = append(replies, reply)
	}
	bot.queue.Send(QUEUE_PRIORITY_LOW, replies, nil)
	return nil
}

func (bot *Bot) broadcastMatch(topic string, exclude_user_a int64, exclude_user_b int64, nick string) error {
	users, err := bot.dbm.ListUnmatchedUsers()
	if err != nil {
		return err
	}
	replies := make([]tgbotapi.Chattable, 0, len(users))
	for i := range users {
		if users[i] == exclude_user_a || users[i] == exclude_user_b {
			continue
		}
		reply := tgbotapi.NewMessage(users[i],
			"【私聊已配对】\n"+
				"\n"+
				"["+nick+"] "+topic)
		reply.DisableNotification = true
		replies = append(replies, reply)
	}
	bot.queue.Send(QUEUE_PRIORITY_LOW, replies, nil)
	return nil
}

func (bot *Bot) replyError(err error, msg *tgbotapi.Message, fatal bool) {
	if err != nil {
		bot.quickReply(
			"「世界树」\n"+
				"\n"+
				"程序发生了错误，刚刚的消息可能没有送达。",
			msg)
		if fatal {
			panic(err)
		} else {
			log.Printf("Error: %+v\n", err)
		}
	}
}

func (bot *Bot) hashIdentification(chat *tgbotapi.Chat) string {
	date_seed := (time.Now().Unix() + 5*3600) / 86400
	hash_sum := sha1.Sum([]byte(fmt.Sprintf("%s %x %x %s %x %s %x", SECRET, chat.ID, len(chat.FirstName), chat.FirstName, len(chat.LastName), chat.LastName, date_seed)))
	return base64.RawURLEncoding.EncodeToString(hash_sum[:6])
}

func (bot *Bot) limitTopic(topic string) string {
	if len(topic) > 64 {
		last_i := 0
		for i := range topic {
			if i > 60 {
				return topic[:last_i] + "…"
			}
			last_i = i
		}
	}
	return topic
}
