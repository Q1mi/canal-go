// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package protocol

import (
	"fmt"

	pbe "github.com/Q1mi/canal-go/protocol/entry"
	pbp "github.com/Q1mi/canal-go/protocol/packet"
	"google.golang.org/protobuf/proto"
)

type Message struct {
	Id         int64
	Entries    []*pbe.Entry
	Raw        bool
	RawEntries interface{}
}

func NewMessage(id int64) *Message {
	message := &Message{Id: id, Entries: nil, Raw: false, RawEntries: nil}
	return message
}

func Decode(data []byte, lazyParseEntry bool) (*Message, error) {
	p := new(pbp.Packet)
	err := proto.Unmarshal(data, p)
	if err != nil {
		return nil, err
	}

	message := new(Message)
	switch p.Type {
	case pbp.PacketType_MESSAGES:
		if !(p.GetCompression() == pbp.Compression_NONE) &&
			!(p.GetCompression() == pbp.Compression_COMPRESSIONCOMPATIBLEPROTO2) { // NONE和兼容pb2的处理方式相同
			panic("compression is not supported in this connector")
		}
		messages := new(pbp.Messages)
		items := make([]*pbe.Entry, 0, len(messages.Messages))
		err := proto.Unmarshal(p.Body, messages)
		if err != nil {
			return nil, err
		}
		if lazyParseEntry {
			message.RawEntries = messages.Messages
			message.Raw = true
		} else {
			for _, value := range messages.Messages {
				entry := new(pbe.Entry)
				err := proto.Unmarshal(value, entry)
				if err != nil {
					return nil, err
				}
				items = append(items, entry)
			}
		}
		message.Entries = items
		message.Id = messages.GetBatchId()
		return message, nil

	case pbp.PacketType_ACK:
		ack := new(pbp.Ack)
		err := proto.Unmarshal(p.Body, ack)
		if err != nil {
			return nil, err
		}
		panic(fmt.Errorf("something goes wrong with reason:%s", ack.GetErrorMessage()))
	default:
		panic(fmt.Errorf("unexpected packet type:%s", p.Type))
	}
}
