package objects

import (
	"encoding/json"
	"log"

	"hmc_server.com/hmc_server/src/game_net"
	"hmc_server.com/hmc_server/src/transform"
)

type NPC struct {
	Id        string
	Position  transform.Vector3
	MoveSpeed float32
}

func (p *NPC) UpdatePosition(pos *transform.Vector3, MoveSpeed float32) {
	p.Position = *pos
	p.MoveSpeed = MoveSpeed

	packet := game_net.New_FRecvPacket_NPCMove()

	packet.Id = p.Id
	packet.Destination = p.Position
	packet.MoveSpeed = p.MoveSpeed

	e, err := json.Marshal(packet)

	if err != nil {
		log.Fatal("Parse Error")
		return
	}

	packet_str := string(e)
	ObjectManagerInst().Players.Range(func(key, value interface{}) bool {
		game_net.NetManagerInst().SendString(value.(*Player).context, packet_str)
		return true
	})
}
