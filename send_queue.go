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

import (
    "container/list"
	"gopkg.in/telegram-bot-api.v4"
)

type sendQueueItem struct {
    priority    int,
    msg_config  []*tgbotapi.Chattable,
    msg_result  []*tgbotapi.Message,
    msg_errors  []error,
    msg_index   int,
    callback    func ([]*tgbotapi.Message, []error)
}

type sendQueue struct {
    channel     chan sendQueueItem,
    high        list.List,
    normal      list.List,
    low         list.List
}

func New(bot *tgbotapi.BotAPI) *sendQueue {
    q := new(sendQueue)
    q.channel := make(chan sendQueueItem)
    go func() {
        for {
            select {
            case item := <-q.channel:
            default:
            }
        }
    }()
    return q
}

func (sendQueue *q) Send(priority int, msg_config []*tgbotapi.Chattable, callback func ([]*tgbotapi.Message, []error)) {
    q.channel <- sendQueueItem {
        priority: priority,
        msg_config: msg_config,
        msg_result: make([]*tgbotapi.Message, len(msg_config)),
        msg_errors: make([]error, len(msg_config)),
        msg_index: 0,
        callback: callback,
    }
}
