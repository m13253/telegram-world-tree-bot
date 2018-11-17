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
	"container/list"
	"log"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	// "gopkg.in/telegram-bot-api.v4"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type sendQueueItem struct {
	priority   int
	msg_config []tgbotapi.Chattable
	msg_result []*tgbotapi.Message
	msg_errors []error
	msg_index  int
	msg_finish uintptr
	callback   func([]*tgbotapi.Message, []error)
}

const (
	QUEUE_PRIORITY_LOW    = 0
	QUEUE_PRIORITY_NORMAL = 1
	QUEUE_PRIORITY_HIGH   = 2
)

type sendQueue struct {
	bot    *tgbotapi.BotAPI
	dbm    *dbManager
	lock   *sync.Mutex
	cv     *sync.Cond
	low    *list.List
	normal *list.List
	high   *list.List
}

func NewSendQueue(bot *tgbotapi.BotAPI, dbm *dbManager) *sendQueue {
	q := &sendQueue{
		bot:    bot,
		dbm:    dbm,
		lock:   new(sync.Mutex),
		cv:     sync.NewCond(new(sync.Mutex)),
		low:    list.New(),
		normal: list.New(),
		high:   list.New(),
	}
	go q.dispatchMessages()
	return q
}

func (q *sendQueue) Send(priority int, msg_config []tgbotapi.Chattable, callback func([]*tgbotapi.Message, []error)) {
	item := &sendQueueItem{
		priority:   priority,
		msg_config: msg_config,
		msg_result: make([]*tgbotapi.Message, len(msg_config)),
		msg_errors: make([]error, len(msg_config)),
		msg_index:  0,
		msg_finish: 0,
		callback:   callback,
	}
	var msg_list *list.List
	switch priority {
	case QUEUE_PRIORITY_LOW:
		msg_list = q.low
	case QUEUE_PRIORITY_NORMAL:
		msg_list = q.normal
	case QUEUE_PRIORITY_HIGH:
		msg_list = q.high
	default:
		panic("Unknown priority")
	}
	q.lock.Lock()
	msg_list.PushBack(item)
	q.lock.Unlock()
	q.cv.Signal()
}

func (q *sendQueue) dispatchMessages() {
	for {
		q.lock.Lock()
		if el := q.high.Front(); el != nil {
			item := el.Value.(*sendQueueItem)
			if item.msg_index == len(item.msg_config) {
				q.high.Remove(el)
				q.lock.Unlock()
			} else {
				q.lock.Unlock()
				q.dispatchMessage(item)
			}
		} else if el := q.normal.Front(); el != nil {
			item := el.Value.(*sendQueueItem)
			if item.msg_index == len(item.msg_config) {
				q.normal.Remove(el)
				q.lock.Unlock()
			} else {
				q.lock.Unlock()
				q.dispatchMessage(item)
			}
		} else if el := q.low.Front(); el != nil {
			item := el.Value.(*sendQueueItem)
			if item.msg_index == len(item.msg_config) {
				q.low.Remove(el)
				q.lock.Unlock()
			} else {
				q.lock.Unlock()
				q.dispatchMessage(item)
			}
		} else {
			q.cv.L.Lock()
			q.lock.Unlock()
			q.cv.Wait()
			q.cv.L.Unlock()
		}
	}
}

func (q *sendQueue) dispatchMessage(item *sendQueueItem) {
	i := item.msg_index
	item.msg_index = i + 1

	delay := time.After(40 * time.Millisecond)

	go func() {
		result := new(tgbotapi.Message)
		var err error
		*result, err = q.bot.Send(item.msg_config[i])
		item.msg_result[i], item.msg_errors[i] = result, err

		if err != nil {
			reflect_msg := reflect.ValueOf(item.msg_config[i])
			chat_id := reflect_msg.FieldByName("ChatID").Interface()
			log.Printf("Send to #%+v failed: %+v\n", chat_id, err)

			if err.Error() == "Forbidden: bot was blocked by the user" || err.Error() == "Forbidden: user is deactivated" {
				if chat_id_int64, ok := chat_id.(int64); ok {
					log.Printf("Removing #%+v from list\n", chat_id)
					q.kickUser(chat_id_int64)
				}
			}
		}

		if int(atomic.AddUintptr(&item.msg_finish, 1)) == len(item.msg_config) {
			if item.callback != nil {
				item.callback(item.msg_result, item.msg_errors)
			}
		}
	}()

	<-delay
}

func (q *sendQueue) kickUser(user_a int64) {
	if user_a == 0 {
		log.Println("kickUser: user_a == 0")
		return
	}
	err := q.dbm.RemoveInvitation(user_a)
	if err != nil {
		log.Println(err)
	}
	err = q.dbm.DisconnectChat(user_a, 0)
	if err != nil {
		log.Println(err)
	}
	err = q.dbm.LeaveLobby(user_a)
	if err != nil {
		log.Println(err)
	}
}
