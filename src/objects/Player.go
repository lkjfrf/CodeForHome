package objects

import (
	"encoding/json"
	"log"
	"net"
	"sync"

	"hmc_server.com/hmc_server/src/game_net"
	"hmc_server.com/hmc_server/src/transform"
)

type Player struct {
	Id              string
	Position        transform.Vector3
	Rotation        transform.Vector3
	TransformUpdate bool
	MoveSpeed       float32
	RotateSpeed     float32
	context         *net.TCPConn
	NearPlayers     sync.Map

	RequestRemoveNearPlayers sync.Map
	NearBoxs                 sync.Map

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

	IsWearableWear bool
	IsDancing      bool
	IsSitting      bool
	StatusNum      int32

	LevelName  string
	VoiceGroup string

	IsSpotSpawn  bool
	SpotPosition transform.Vector3
	SpotRotation transform.Vector3

	IsAtlasSpawn  bool
	AtlasPosition transform.Vector3
	AtlasRotation transform.Vector3

	RecevieHearBeat bool
}

func (p *Player) GetTCPContext() *net.TCPConn {
	return p.context
}

func (p *Player) GetNearPlayers() sync.Map {
	return p.NearPlayers
}

func (p *Player) UpdateMovement(pos *transform.Vector3, rot *transform.Vector3, MoveSpeed float32, RotateSpeed float32) {
	p.TransformUpdate = true
	p.Position = *pos
	p.Rotation = *rot
	p.MoveSpeed = MoveSpeed
	p.RotateSpeed = RotateSpeed
}

func (p *Player) PlayerActionEvent(Id string, ActionId string) {
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

func (p *Player) SendGlobalChat(Message string, UserName string, Players sync.Map) {
	packet := game_net.New_FRecvPacket_GlobalChat()

	packet.Id = p.Id
	packet.UserName = UserName
	packet.Message = Message

	e, err := json.Marshal(packet)
	if err != nil {
		log.Fatal("Parse Error")
		return
	}

	Players.Range(func(key, value interface{}) bool {
		if p.Id != key.(string) {
			game_net.NetManagerInst().SendString(value.(*Player).context, string(e))
		}

		return true
	})
}

func (p *Player) Init() {
	p.IsWearableWear = false
}

func (p *Player) Destroy() {
	//game_net.NetManagerInst().RemoveContextRecevier(p.GetTCPContext())
	//ObjectManagerInst().LeaveVoiceGroup(p.Id)
	ObjectManagerInst().Players.Delete(p.Id)

	packet := game_net.New_FRecvPacket_OtherPlayerDestroyInfo()

	packet.Id = p.Id
	p.LevelName = ""

	e, err := json.Marshal(packet)

	if err != nil {
		log.Fatal("Parse Error")
		return
	}
	stre := string(e)
	ObjectManagerInst().Players.Range(func(key, value interface{}) bool {
		if key.(string) != p.Id {
			if _, ok := value.(*Player).NearPlayers.Load(p.Id); ok {
				value.(*Player).NearPlayers.Delete(p.Id)
				game_net.NetManagerInst().SendString(value.(*Player).context, stre)
			}
		}

		return true

	})

}
