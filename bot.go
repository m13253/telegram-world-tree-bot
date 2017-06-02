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
	"strings"
	"gopkg.in/telegram-bot-api.v4"
	_ "github.com/mattn/go-sqlite3"
)

type Bot struct {
	api     *tgbotapi.BotAPI
	dbm     *dbManager
	queue   *sendQueue
	updates <-chan tgbotapi.Update
}

func NewBot(api *tgbotapi.BotAPI, dbm *dbManager) (bot *Bot, err error) {
	bot = &Bot {
		api:        api,
		dbm:        dbm,
		queue:      NewSendQueue(api),
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	bot.updates, err = api.GetUpdatesChan(u)
	if err != nil { return }

	return
}

func (bot *Bot) Run() {
	for update := range bot.updates {
		msg := update.Message
		if msg != nil && msg.Chat.IsPrivate() {

			if strings.HasPrefix(msg.Text, "/") {
				printLog(msg.From, msg.Text, false)
			}

			cmd := msg.Command()
			if cmd == "" {
				bot.handleMessage(msg)
			} else if cmd == "start" {
				bot.handleStart(msg)
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
				"「世界树」\n" +
				"\n" +
				"本服务不保留聊天记录，故无法追踪消息编辑状态。\n" +
				"由于这个限制，你无法使用消息编辑功能。",
				edit_msg)
		}

		query := update.CallbackQuery
		if query != nil {
			bot.handleCallbackQuery(query)
		}
	}
}

func (bot *Bot) sendTopicList(user int64, caption string) (count int, err error) {
	topics, err := bot.dbm.ListTopics()
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
	_, err = bot.api.Send(reply)
	return
}

func (bot *Bot) broadcastNewTopic(topic string, short_topic string, exclude_user int64) {
	users, err := bot.dbm.ListPendingUsers()
	if err != nil {
		log.Println(err)
		return
	}
	reply_markup := tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton {
			tgbotapi.InlineKeyboardButton {
				Text: "\u2764\ufe0f 赞",
				CallbackData: &short_topic,
			},
		})
	for i := range users {
		if users[i] == exclude_user {
			continue
		}
		reply := tgbotapi.NewMessage(users[i],
			"【新话题】\n" +
			"\n" +
			topic)
		reply.ReplyMarkup = reply_markup
		reply.DisableNotification = true
		bot.api.Send(reply)
	}
}

func (bot *Bot) limitTopic(topic string) string {
	if len(topic) > 64 {
		last_i := 0
		for i, _ := range topic {
			if i > 60 {
				return topic[:last_i] + "…"
			}
			last_i = i
		}
	}
	return topic
}

func (bot *Bot) quickReply(text string, msg *tgbotapi.Message) {
	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	reply.ReplyToMessageID = msg.MessageID
	reply.DisableWebPagePreview = true
	bot.queue.Send(QUEUE_PRIORITY_HIGH, []tgbotapi.Chattable { reply }, nil)
}

func (bot *Bot) replyError(err error, msg *tgbotapi.Message) {
	if err != nil {
		log.Println(err)
		bot.quickReply(
			"「世界树」\n" +
			"\n" +
			"程序发生了错误，刚刚的消息可能没有送达。",
			msg)
	}
}
