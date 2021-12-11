package objects

import (
	"encoding/json"
	"log"
	"net"
	"sync"

	"hmc_server.com/hmc_server/src/game_net"
	"hmc_server.com/hmc_server/src/transform"
)

type SquidGamePlayer struct {
	Id              string
	Position        transform.Vector3
	Rotation        transform.Vector3
	TransformUpdate bool
	MoveSpeed       float32
	RotateSpeed     float32
	context         *net.TCPConn
	NearPlayers     sync.Map

	RequestRemoveNearPlayers sync.Map

	IsMan          bool
	SkinIndex      int32
	TopIndex       int32
	BottomIndex    int32
	HairIndex      int32
	ShoesIndex     int32
	HairColorIndex int32
	FaceIndex      int32
	AccessoryIndex int32

	FirstName  string
	LastName   string
	BirthDay   int32
	BirthMonth int32
	DealerType string
	Country    string

	LevelName     string
	RoomIndex     int32
	SquidGameMode bool
	Wearable      bool

	RecevieHearBeat bool
}

func (p *SquidGamePlayer) GetTCPContext() *net.TCPConn {
	return p.context
}

func (p *SquidGamePlayer) GetNearPlayers() sync.Map {
	return p.NearPlayers
}

func (p *SquidGamePlayer) UpdateMovement(pos *transform.Vector3, rot *transform.Vector3, MoveSpeed float32, RotateSpeed float32) {
	p.TransformUpdate = true
	p.Position = *pos
	p.Rotation = *rot
	p.MoveSpeed = MoveSpeed
	p.RotateSpeed = RotateSpeed
}

func (p *SquidGamePlayer) PlayerActionEvent(Id string, ActionId string) {
	packet := game_net.New_FRecvPacket_PlayerActionEvent()

	packet.Id = Id
	packet.ActionId = ActionId

	e, err := json.Marshal(packet)

	if err != nil {
		log.Fatal("Parse Error")
		return
	}

	game_net.NetManagerInst().SendString(p.context, string(e))
}

// func (p *SquidGamePlayer) Init() {
// 	p.IsWearableWear = false
// }

func (p *SquidGamePlayer) Destroy() {
	//game_net.NetManagerInst().RemoveContextRecevier(p.GetTCPContext())
	//ObjectManagerInst().LeaveVoiceGroup(p.Id)

	packet := game_net.New_FRecvPacket_OtherPlayerDestroyInfo()

	packet.Id = p.Id
	p.LevelName = ""

	e, err := json.Marshal(packet)

	if err != nil {
		log.Fatal("Parse Error")
		return
	}
	stre := string(e)
	ObjectManagerInst().SquidPlayers.Range(func(key, value interface{}) bool {
		if key.(string) != p.Id {
			if _, ok := value.(*SquidGamePlayer).NearPlayers.Load(p.Id); ok {
				value.(*SquidGamePlayer).NearPlayers.Delete(p.Id)
				game_net.NetManagerInst().SendString(value.(*SquidGamePlayer).context, stre)
			}
		}

		return true

	})

}
