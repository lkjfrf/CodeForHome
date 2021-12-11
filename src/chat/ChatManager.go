package chat

import (
	"encoding/json"
	"log"
	"sync"

	"hmc_server.com/hmc_server/src/game_net"
	"hmc_server.com/hmc_server/src/helper"
	"hmc_server.com/hmc_server/src/objects"
)

type ChatManager struct {
}

var Instance *ChatManager
var once sync.Once

func ChatManagerInst() *ChatManager {
	once.Do(func() {
		Instance = &ChatManager{}
	})
	return Instance
}

func (cm *ChatManager) Init() {

	game_net.NetManagerInst().Callbacks["FSendPacket_GlobalChat"] = func(v interface{}) {
		type Message struct {
			Id       string
			UserName string
			Message  string
		}
		msg := Message{}
		helper.FillStruct_Interface(v, &msg)

		if p, ok := objects.ObjectManagerInst().Players.Load(msg.Id); ok {
			p.(*objects.Player).SendGlobalChat(msg.Message, msg.UserName, objects.ObjectManagerInst().Players)
		}
	}

	game_net.NetManagerInst().Callbacks["FSendPacket_PrivateChat"] = func(v interface{}) {
		type Message struct {
			Id       string
			TargetId string
			Message  string
			IsOnline bool
			UserName string
		}
		msg := Message{}
		helper.FillStruct_Interface(v, &msg)
		packet := game_net.New_FRecvPacket_PrivateChat()
		if p, ok := objects.ObjectManagerInst().Players.Load(msg.Id); ok {
			packet.IsOnline = true
			packet.Message = msg.Message
			packet.TargetId = msg.TargetId
			packet.UserName = msg.UserName
			e, err := json.Marshal(packet)
			if err != nil {
				log.Fatal("Parse Error")
				return
			}
			objects.ObjectManager().SendPacketToTarget
		} else {
			packet.IsOnline = false
			e, err := json.Marshal(packet)
			if err != nil {
				log.Fatal("Parse Error")
				return
			}
			game_net.NetManagerInst().SendString(Players[packet.Id].GetTCPContext(), string(e))
		}
	}
}
