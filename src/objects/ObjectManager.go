package objects

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"reflect"
	"sync"
	"time"

	"github.com/mitchellh/mapstructure"
	"hmc_server.com/hmc_server/src/database"
	"hmc_server.com/hmc_server/src/game_net"
	"hmc_server.com/hmc_server/src/helper"
	"hmc_server.com/hmc_server/src/structs"
	"hmc_server.com/hmc_server/src/transform"
)

type VoiceUserState struct {
	OnHandsUp bool
	OnMic     bool
}

type ObjectManager struct {
	Players sync.Map

	TreasureBoxs sync.Map
	Cars         sync.Map

	VoiceGroup sync.Map //map[string]map[string]*VoiceUserState

	TotalLoginUserCount int

	LMTH []string
	RMTH []string

	XPositions []int32
	YPositions []int32

	SquidPlayers    sync.Map
	SquidGame       map[int32]map[string]interface{}
	SquidGameMinute int
}

var instance *ObjectManager
var once sync.Once

func ObjectManagerInst() *ObjectManager {
	once.Do(func() {
		instance = &ObjectManager{}
	})
	return instance
}

func (m *ObjectManager) Init() {

	go func() {
		for {
			m.SquidPlayers.Range(func(key, value interface{}) bool {
				player := value.(*SquidGamePlayer)
				if player.TransformUpdate {
					player.TransformUpdate = false
					packet := game_net.New_FRecvPacket_OtherPlayerMove()
					packet.Id = player.Id
					packet.Destination = player.Position
					packet.DestRotation = player.Rotation
					packet.MoveSpeed = player.MoveSpeed
					packet.RotateSpeed = player.RotateSpeed

					player.NearPlayers.Range(func(key, value interface{}) bool {
						m.SendPacketToTarget(packet, key.(string))
						return true
					})
				}
				return true
			})
			time.Sleep(time.Millisecond * 200)
		}
	}()

	go func() {
		for {
			m.Players.Range(func(key, value interface{}) bool {
				player := value.(*Player)
				if player.TransformUpdate {
					player.TransformUpdate = false
					packet := game_net.New_FRecvPacket_OtherPlayerMove()
					packet.Id = player.Id
					packet.Destination = player.Position
					packet.DestRotation = player.Rotation
					packet.MoveSpeed = player.MoveSpeed
					packet.RotateSpeed = player.RotateSpeed
					e, err := json.Marshal(packet)
					if err != nil {
						log.Println("Parse Error")
						return true
					}
					packet_str := string(e)
					player.NearPlayers.Range(func(key, value interface{}) bool {
						if player.Id != key.(string) {
							game_net.NetManagerInst().SendString(value.(*Player).context, packet_str)
						}
						return true
					})
				}
				return true
			})
			time.Sleep(time.Millisecond * 200)
		}
	}()

	fmt.Println("INIT_ObjectManager")

	m.Players = sync.Map{}
	m.SquidPlayers = sync.Map{}
	m.TreasureBoxs = sync.Map{}
	m.Cars = sync.Map{}

	m.VoiceGroup = sync.Map{}

	m.TotalLoginUserCount = 0

	go func(deleteplayer chan *net.TCPConn) {
		for {
			select {
			case v, ok := <-deleteplayer:
				if !ok {
					continue
				}

				log.Println("LogoutProcess")

				id := m.GetIdByConn(v)
				game_net.NetManagerInst().RemoveContextRecevier(v)

				if p, ok := m.Players.Load(id); ok {
					p.(*Player).Destroy()

					playerCount := 0
					m.Players.Range(func(key, value interface{}) bool {
						playerCount++
						return true
					})

					log.Println("NewPlayer EXIT. Total Count of Online Players =", playerCount)
				}
				if p, ok := m.SquidPlayers.Load(id); ok {
					p.(*SquidGamePlayer).Destroy()
					m.SquidPlayers.Delete(id)
					delete(m.SquidGame[p.(*SquidGamePlayer).RoomIndex], id)

					//m.SquidGame[]
					log.Println("SquidGame EXIT")
				}
			}
		}
	}(game_net.NetManagerInst().DeletePlayerChan)

	game_net.NetManagerInst().SettingPlayerFunc = func(player *net.TCPConn) {
		m.RunPlayerContent(player)
	}
	m.DummyPosInit()

	// Database Data Update
	go func() {
		for {
			m.BroadCastWRCRank()
			time.Sleep(time.Second)

			//m.BroadCastLoveForest()
			//time.Sleep(time.Second)
		}
	}()

	// Near Player
	go func() {
		for {
			m.SpawnProcess()
			m.DummyPosChange()
			time.Sleep(time.Second * 3)
		}
	}()

	// SquidGame Near Player
	go func() {
		for {
			m.SquidGameSpawnProcess()
			time.Sleep(time.Second * 2)
		}
	}()

	// SquidGame Time Check
	m.SquidGameMinute = time.Now().Minute()
	m.SquidGameTimeCheck()

	m.SquidGame = map[int32]map[string]interface{}{}
	for i := int32(0); i < 10; i++ {
		m.SquidGame[i] = map[string]interface{}{}
	}
}

func (m *ObjectManager) RunHeartBeatProcess(id string, playerConn *net.TCPConn) {

	for {
		time.Sleep(time.Minute * 30)

		if v, ok := m.Players.Load(id); ok {
			if !v.(*Player).RecevieHearBeat {

				if v.(*Player).GetTCPContext().RemoteAddr().String() != playerConn.RemoteAddr().String() {
					return
				}

				v.(*Player).Destroy()
				game_net.NetManagerInst().RemoveContextRecevier(v.(*Player).GetTCPContext())
				m.Players.Delete(id)
				return
			} else {
				v.(*Player).RecevieHearBeat = false
				continue
			}
		} else {
			game_net.NetManagerInst().RemoveContextRecevier(playerConn)
			return
		}
	}
}

func (m *ObjectManager) RunPlayerContent(playerId *net.TCPConn) {
	game_net.NetManagerInst().Callbacks["FSendPacket_PlayerLogout"] = func(v interface{}) {
		type Logout struct {
			Id string
		}
		logout := Logout{}
		helper.FillStruct_Interface(v, &logout)

		fmt.Println("LogoutPlayer : ", logout.Id)

		if p, ok := m.Players.Load(logout.Id); ok {
			p.(*Player).Destroy()
		}
	}
	game_net.NetManagerInst().Callbacks["FSendPacket_PlayerActionEvent"] = func(v interface{}) {
		type RecvData struct {
			Id       string
			ActionId string
		}

		recvdata := RecvData{}
		helper.FillStruct_Interface(v, &recvdata)

		if p, ok := m.Players.Load(recvdata.Id); ok {
			if recvdata.ActionId == "CharacterDance" {
				p.(*Player).IsDancing = true
			}

			if recvdata.ActionId == "CharacterWalk" {
				p.(*Player).IsDancing = false
			}

			if recvdata.ActionId == "CharacterStandToSit" {
				p.(*Player).IsSitting = true
			}

			if recvdata.ActionId == "CharacterSitToStand" {
				p.(*Player).IsSitting = false
			}

			p.(*Player).NearPlayers.Range(func(key, value interface{}) bool {
				value.(*Player).PlayerActionEvent(recvdata.Id, recvdata.ActionId)
				return true
			})
		}
	}
	game_net.NetManagerInst().Callbacks["FSendPacket_Voice"] = func(v interface{}) {
		type VoiceData struct {
			Id          string
			VoiceData   []uint8
			Numchannels int32
			SampleRate  int32
			PCMSize     int32
		}
		voiceData := VoiceData{}
		helper.FillStruct_Interface(v, &voiceData)
		packet := game_net.New_FRecvPacket_Voice()

		packet.Id = voiceData.Id
		for _, voiceData := range voiceData.VoiceData {
			packet.VoiceData = append(packet.VoiceData, uint16(voiceData))
		}

		packet.SampleRate = voiceData.SampleRate
		packet.Numchannels = voiceData.Numchannels
		packet.PCMSize = voiceData.PCMSize

		e, err := json.Marshal(packet)

		if err != nil {
			log.Fatal("Parse Error")
			return
		}

		if p, ok := m.Players.Load(packet.Id); ok {

			if p.(*Player).VoiceGroup != "" {

				// 마이크가 켜져있는 플레이어가 아니면 송출을 막는다.
				if c, ok := m.VoiceGroup.Load(p.(*Player).VoiceGroup); ok {
					if d, ok := (c.(*sync.Map).Load(packet.Id)); ok {
						// 방장이 아닐경우에만 Mic를 체크한다.
						if p.(*Player).VoiceGroup != packet.Id {
							if !d.(*VoiceUserState).OnMic {
								return
							}
						}
					}
				}

				if p, ok := m.VoiceGroup.Load(p.(*Player).VoiceGroup); ok {
					p.(*sync.Map).Range(func(key, value interface{}) bool {
						if c, ok := m.Players.Load(key.(string)); ok {
							game_net.NetManagerInst().SendString(c.(*Player).GetTCPContext(), string(e))
						}

						return true
					})
				}
				return
			}

			//log.Println("voice send Start")

			p.(*Player).NearPlayers.Range(func(key, value interface{}) bool {
				if p.(*Player).Id != key.(string) && value.(*Player).VoiceGroup == "" {
					game_net.NetManagerInst().SendString(value.(*Player).GetTCPContext(), string(e))
				}

				return true
			})

			//log.Println("voice send End")
		}
	}
	game_net.NetManagerInst().Callbacks["FSendPacket_Notice"] = func(v interface{}) {
		type RecvData struct {
			Message string
		}

		recvdata := RecvData{}
		helper.FillStruct_Interface(v, &recvdata)

		packet := game_net.New_FRecvPacket_Notice()
		packet.Message = recvdata.Message

		m.BroadCastPacketToAll(packet)
	}
	game_net.NetManagerInst().Callbacks["FSendPacket_CreateTreasureBox"] = func(v interface{}) {
		boxinfo := game_net.New_FRecvPacket_CreateTreasureBox()
		helper.FillStruct_Interface(v, &boxinfo)

		m.AddNewTreasureBox(boxinfo.Id, boxinfo.Point, boxinfo.Position, boxinfo.LevelName)
	}
	game_net.NetManagerInst().Callbacks["FSendPacket_DestroyTreasureBox"] = func(v interface{}) {
		type RecvData struct {
			Id string
		}

		recvdata := RecvData{}
		helper.FillStruct_Interface(v, &recvdata)

		packet := game_net.New_FRecvPacket_DestroyTreasureBox()
		packet.Id = recvdata.Id

		m.TreasureBoxs.Delete(packet.Id)

		m.BroadCastPacketExceptMe(packet, packet.Id)
	}
	game_net.NetManagerInst().Callbacks["FSendPacket_WorldTeleport"] = func(v interface{}) {
		packet := game_net.New_FRecvPacket_WorldTeleport()
		helper.FillStruct_Interface(v, &packet)
		packet.PacketName = "FRecvPacket_WorldTeleport"

		m.BroadCastPacketExceptMe(packet, packet.Id)
	}
	game_net.NetManagerInst().Callbacks["FSendPacket_CarMove"] = func(v interface{}) {
		packet := game_net.New_FRecvPacket_OtherCarMove()
		helper.FillStruct_Interface(v, &packet)
		packet.PacketName = "FRecvPacket_OtherCarMove"

		m.UpdateCar(packet.Id, packet.ServerDistacne)

		m.BroadCastPacketExceptMe(packet, packet.Id)
	}
	game_net.NetManagerInst().Callbacks["FSendPacket_CarSpawnInfo"] = func(v interface{}) {
		packet := game_net.New_FRecvPacket_OtherCarSpawnInfo()
		helper.FillStruct_Interface(v, &packet)
		packet.PacketName = "FRecvPacket_OtherCarSpawnInfo"

		m.AddCar(packet.Id, packet.TypeNum, packet.PathTag)

		m.BroadCastPacketExceptMe(packet, packet.Id)
	}
	game_net.NetManagerInst().Callbacks["FSendPacket_CarDestroy"] = func(v interface{}) {
		packet := game_net.New_FRecvPacket_OtherCarDestroy()
		helper.FillStruct_Interface(v, &packet)
		packet.PacketName = "FRecvPacket_OtherCarDestroy"

		m.RemoveCar(packet.Id)

		m.BroadCastPacketExceptMe(packet, packet.Id)
	}
	game_net.NetManagerInst().Callbacks["FSendPacket_Wearable"] = func(v interface{}) {
		packet := game_net.New_FRecvPacket_Wearable()
		helper.FillStruct_Interface(v, &packet)
		packet.PacketName = "FRecvPacket_Wearable"

		if p, ok := m.Players.Load(packet.Id); ok {
			if packet.IsWear {
				p.(*Player).IsWearableWear = true
			} else {
				p.(*Player).IsWearableWear = false
			}
		}

		m.BroadCastPacketExceptMe(packet, packet.Id)
	}
	game_net.NetManagerInst().Callbacks["FSendPacket_SpotSpawn"] = func(v interface{}) {
		type SpotSpawn struct {
			Id         string
			SpawnPoint transform.Vector3
		}

		spawnInfo := SpotSpawn{}
		helper.FillStruct_Interface(v, &spawnInfo)

		if p, ok := m.Players.Load(spawnInfo.Id); ok {
			p.(*Player).IsSpotSpawn = true
			p.(*Player).SpotPosition = spawnInfo.SpawnPoint

			packet := game_net.New_FRecvPacket_SpotSpawn()

			packet.Id = spawnInfo.Id
			packet.SpawnPoint = spawnInfo.SpawnPoint

			packet.PacketName = "FRecvPacket_SpotSpawn"

			p.(*Player).NearPlayers.Range(func(key, value interface{}) bool {
				if p.(*Player).Id != key.(string) {
					m.SendPacketToTarget(packet, key.(string))
				}

				return true
			})
		}
	}
	game_net.NetManagerInst().Callbacks["FSendPacket_SpotMove"] = func(v interface{}) {
		type SpotMove struct {
			Id          string
			Position    transform.Vector3
			Rotation    transform.Vector3
			MoveSpeed   float32
			RotateSpeed float32
		}
		pos := SpotMove{}
		helper.FillStruct_Interface(v, &pos)

		packet := game_net.New_FRecvPacket_SpotMove()
		packet.Id = pos.Id
		packet.Destination = pos.Position
		packet.DestRotation = pos.Rotation
		packet.MoveSpeed = pos.MoveSpeed
		packet.RotateSpeed = pos.RotateSpeed

		if p, ok := m.Players.Load(packet.Id); ok {
			p.(*Player).SpotPosition = pos.Position
			p.(*Player).SpotRotation = pos.Rotation
			p.(*Player).NearPlayers.Range(func(Key, Value interface{}) bool {
				if p.(*Player).Id != Key.(string) {
					m.SendPacketToTarget(packet, Key.(string))
				}

				return true
			})
		}
	}
	game_net.NetManagerInst().Callbacks["FSendPacket_SpotDestroy"] = func(v interface{}) {
		type SpotDestroy struct {
			Id string
		}

		destroyInfo := SpotDestroy{}
		helper.FillStruct_Interface(v, &destroyInfo)

		if p, ok := m.Players.Load(destroyInfo.Id); ok {
			p.(*Player).IsSpotSpawn = false
			packet := game_net.New_FRecvPacket_SpotDestroy()
			helper.FillStruct_Interface(v, &packet)
			packet.PacketName = "FRecvPacket_SpotDestroy"

			p.(*Player).NearPlayers.Range(func(key, value interface{}) bool {
				if p.(*Player).Id != key.(string) {
					m.SendPacketToTarget(packet, key.(string))
				}

				return true
			})
		}
	}
	game_net.NetManagerInst().Callbacks["FSendPacket_AtlasSpawn"] = func(v interface{}) {
		type AtlasSpawn struct {
			Id         string
			SpawnPoint transform.Vector3
		}

		spawnInfo := AtlasSpawn{}
		helper.FillStruct_Interface(v, &spawnInfo)

		if p, ok := m.Players.Load(spawnInfo.Id); ok {
			p.(*Player).IsAtlasSpawn = true
			p.(*Player).AtlasPosition = spawnInfo.SpawnPoint

			packet := game_net.New_FRecvPacket_AtlasSpawn()

			packet.Id = spawnInfo.Id
			packet.SpawnPoint = spawnInfo.SpawnPoint

			packet.PacketName = "FRecvPacket_AtlasSpawn"

			p.(*Player).NearPlayers.Range(func(Key, Value interface{}) bool {
				if p.(*Player).Id != Key.(string) {
					m.SendPacketToTarget(packet, Key.(string))
				}

				return true
			})
		}
	}
	game_net.NetManagerInst().Callbacks["FSendPacket_AtlasMove"] = func(v interface{}) {
		type AtlasMove struct {
			Id          string
			Position    transform.Vector3
			Rotation    transform.Vector3
			MoveSpeed   float32
			RotateSpeed float32
		}
		pos := AtlasMove{}
		helper.FillStruct_Interface(v, &pos)

		packet := game_net.New_FRecvPacket_AtlasMove()
		packet.Id = pos.Id
		packet.Destination = pos.Position
		packet.DestRotation = pos.Rotation
		packet.MoveSpeed = pos.MoveSpeed
		packet.RotateSpeed = pos.RotateSpeed

		if p, ok := m.Players.Load(packet.Id); ok {
			p.(*Player).AtlasPosition = pos.Position
			p.(*Player).AtlasRotation = pos.Rotation

			p.(*Player).NearPlayers.Range(func(Key, Value interface{}) bool {
				if p.(*Player).Id != Key.(string) {
					m.SendPacketToTarget(packet, Key.(string))
				}

				return true
			})
		}
	}
	game_net.NetManagerInst().Callbacks["FSendPacket_AtlasDestroy"] = func(v interface{}) {
		packet := game_net.New_FRecvPacket_AtlasDestroy()
		helper.FillStruct_Interface(v, &packet)
		packet.PacketName = "FRecvPacket_AtlasDestroy"

		if p, ok := m.Players.Load(packet.Id); ok {

			p.(*Player).IsAtlasSpawn = false

			p.(*Player).NearPlayers.Range(func(Key, Value interface{}) bool {
				if p.(*Player).Id != Key.(string) {
					m.SendPacketToTarget(packet, Key.(string))
				}

				return true
			})
		}
	}
	game_net.NetManagerInst().Callbacks["FSendPacket_SetCostume"] = func(v interface{}) {
		packet := game_net.New_FRecvPacket_SetCostume()
		helper.FillStruct_Interface(v, &packet)
		packet.PacketName = "FRecvPacket_SetCostume"

		if packet.ClothType == 0 {
			if p, ok := m.Players.Load(packet.Id); ok {
				p.(*Player).TopIndex = packet.ClothIndex
			}
			game_net.DBManagerInst().UpdateCostume(packet.Id, int(packet.ClothIndex), 0)
		}
		if packet.ClothType == 1 {
			if p, ok := m.Players.Load(packet.Id); ok {
				p.(*Player).BottomIndex = packet.ClothIndex
			}
			game_net.DBManagerInst().UpdateCostume(packet.Id, int(packet.ClothIndex), 1)
		}
		if packet.ClothType == 2 {
			if p, ok := m.Players.Load(packet.Id); ok {
				p.(*Player).ShoesIndex = packet.ClothIndex
			}
			game_net.DBManagerInst().UpdateCostume(packet.Id, int(packet.ClothIndex), 2)
		}
		if packet.ClothIndex == 3 {
			if p, ok := m.Players.Load(packet.Id); ok {
				p.(*Player).AccessoryIndex = packet.ClothIndex
			}
			game_net.DBManagerInst().UpdateCostume(packet.Id, int(packet.ClothIndex), 3)
		}

		m.BroadCastPacketExceptMe(packet, packet.Id)
	}
	game_net.NetManagerInst().Callbacks["FSendPacket_CreateVoiceGroup"] = func(v interface{}) {

	}
	game_net.NetManagerInst().Callbacks["FSendPacket_JoinVoiceGroup"] = func(v interface{}) {
		type JoinVoiceGroup struct {
			Id       string
			VoiceKey string
			Password string
		}
		voiceGroup := JoinVoiceGroup{}
		helper.FillStruct_Interface(v, &voiceGroup)
		// 모더레이터 진입시 방생성
		// if voiceGroup.Id == voiceGroup.VoiceKey && game_net.DBManagerInst().IsModerator(voiceGroup.Id) {
		// 	syncMap := &sync.Map{}
		// 	voiceData := &VoiceUserState{}
		// 	voiceData.OnMic = true
		// 	syncMap.Store(voiceGroup.Id, voiceData)
		// 	m.VoiceGroup.Store(voiceGroup.VoiceKey, syncMap)
		// }

		var IswrongPassword bool = false

		{

			// 모더레이터가 없는방이면 입장 불가
			if _, ok := m.VoiceGroup.Load(voiceGroup.VoiceKey); !ok {

				log.Println("moderator not found")

				errpacket1 := game_net.New_FRecvPacket_ModeratorNotInRoom()
				if errp1, ok := m.Players.Load(voiceGroup.Id); ok {
					m.SendPacketToTarget(errpacket1, errp1.(*Player).Id)
				}
				return
			}
		}

		if p1, ok := m.Players.Load(voiceGroup.Id); ok {

			p1.(*Player).VoiceGroup = voiceGroup.VoiceKey

			if p, ok := m.VoiceGroup.Load(p1.(*Player).VoiceGroup); ok {
				if voiceGroup.Id != voiceGroup.VoiceKey {
					//비번 맞는지 검사
					moderatorPacket := game_net.New_FRecvPacket_GetModerator()
					//moderatorPacket.Moderators = game_net.DBManagerInst().GetModerators()

					for _, v := range moderatorPacket.Moderators {
						if _, ok := m.Players.Load(v.PlayerId); ok {
							if voiceGroup.VoiceKey != v.PlayerId {
								continue
							}
							if voiceGroup.Password != v.Password {
								IswrongPassword = true
								break
							}
						}
					}
				}

				if IswrongPassword {
					p1.(*Player).VoiceGroup = ""
					log.Println("passworld error")

					errpacket2 := game_net.New_FRecvPacket_ModeratorPasswordInvalid()
					if errp2, ok := m.Players.Load(voiceGroup.Id); ok {
						m.SendPacketToTarget(errpacket2, errp2.(*Player).Id)
					}
					return
				}

				// 처음 접속한 유저는 만들어준다.
				if _, ok := p.(*sync.Map).Load(voiceGroup.Id); !ok {
					voiceData := &VoiceUserState{}
					voiceData.OnHandsUp = false
					voiceData.OnMic = false
					p.(*sync.Map).Store(voiceGroup.Id, voiceData)
				}
				keys := make([]string, 0, 0)

				// 맨앞에 VoiceKey 즉 Moderator
				keys = append(keys, voiceGroup.VoiceKey)

				p.(*sync.Map).Range(func(key, value interface{}) bool {
					if key != voiceGroup.VoiceKey {
						keys = append(keys, key.(string))
					}
					return true
				})

				pac := game_net.New_FRecvPacket_JoinVoiceGroupUpdate()

				Ids := game_net.ModeratorUserInfo{}

				for _, key := range keys {
					Ids.Id = key
					Ids.URL = "https://h-festival.s3.ap-northeast-2.amazonaws.com/images/T_HD.png"

					if key == voiceGroup.VoiceKey {
						Ids.IsModerator = true
					} else {
						Ids.IsModerator = false
					}

					if i, ok := m.Players.Load(Ids.Id); ok {
						Ids.UserName = i.(*Player).FirstName + " " + i.(*Player).LastName
					}
					if d, ok := p.(*sync.Map).Load(key); ok {
						Ids.OnHandsUp = d.(*VoiceUserState).OnHandsUp
						Ids.OnMic = d.(*VoiceUserState).OnMic
					}

					pac.Ids = append(pac.Ids, Ids)
				}

				e, err := json.Marshal(pac)
				if err != nil {
					log.Fatal("Parse Error")
					return
				}

				p.(*sync.Map).Range(func(key, value interface{}) bool {
					{
						if _p, ok := m.Players.Load(key.(string)); ok {
							game_net.NetManagerInst().SendString(_p.(*Player).GetTCPContext(), string(e))

							//모데라이터면 처음부터 마이크를 켜주자
							if voiceGroup.Id == voiceGroup.VoiceKey {
								packet2 := game_net.New_FRecvPacket_VoicePlayerChoice()
								packet2.Id = _p.(*Player).Id
								packet2.ChoiceUpId = _p.(*Player).Id

								e2, _ := json.Marshal(packet2)

								game_net.NetManagerInst().SendString(_p.(*Player).GetTCPContext(), string(e2))
							}
						}
					}
					return true
				})
			}
		}

		if p, ok := m.Players.Load(voiceGroup.VoiceKey); ok {
			if voiceGroup.Id == voiceGroup.VoiceKey {
				return
			} else {
				newplayerpacket := game_net.New_FRecvPacket_NewPlayerJoinRoom()
				newplayerpacket.PlayerId = voiceGroup.Id
				m.SendPacketToTarget(newplayerpacket, p.(*Player).Id)
			}
		}
	}
	game_net.NetManagerInst().Callbacks["FSendPacket_LeaveVoiceGroup"] = func(v interface{}) {
		type LeaveVoiceGroup struct {
			Id string
		}
		voiceGroup := LeaveVoiceGroup{}
		helper.FillStruct_Interface(v, &voiceGroup)

		m.LeaveVoiceGroup(voiceGroup.Id)
	}
	game_net.NetManagerInst().Callbacks["FSendPacket_SendToVoiceChat"] = func(v interface{}) {
		type SendToVoiceChat struct {
			Id       string
			UserName string
			Message  string
		}
		SendToVoice := SendToVoiceChat{}
		helper.FillStruct_Interface(v, &SendToVoice)

		if _p, ok := m.Players.Load(SendToVoice.Id); ok {

			if p, ok := m.VoiceGroup.Load(_p.(*Player).VoiceGroup); ok {

				packet := game_net.New_FRecvPacket_SendToVoiceChat()
				packet.Id = SendToVoice.Id
				packet.UserName = SendToVoice.UserName
				packet.Message = SendToVoice.Message

				e, err := json.Marshal(packet)
				if err != nil {
					log.Println("Parse Error")
					return
				}

				p.(*sync.Map).Range(func(key, value interface{}) bool {
					if _p, ok := m.Players.Load(key.(string)); ok {
						game_net.NetManagerInst().SendString(_p.(*Player).GetTCPContext(), string(e))
					}

					return true
				})
			}
		}
	}
	// game_net.NetManagerInst().Callbacks["FSendPacket_GetModerator"] = func(v interface{}) {
	// 	type GetModerator struct {
	// 		Id string
	// 	}

	// 	getModerator := GetModerator{}
	// 	helper.FillStruct_Interface(v, &getModerator)

	// 	if p, ok := m.Players.Load(getModerator.Id); ok {
	// 		moderatorPacket := game_net.New_FRecvPacket_GetModerator()
	// 		moderatorPacket.Moderators = game_net.DBManagerInst().GetModerators()
	// 		for k, v := range moderatorPacket.Moderators {
	// 			if _, ok := m.Players.Load(v.PlayerId); ok {
	// 				moderatorPacket.Moderators[k].IsLogin = true
	// 			}
	// 		}
	// 		m.SendPacketToTarget(moderatorPacket, p.(*Player).Id)
	// 	}
	// }
	game_net.NetManagerInst().Callbacks["FSendPacket_PlayerLevelChange"] = func(v interface{}) {
		type LevelChange struct {
			Id             string
			LevelName      string
			MoveSpeed      float32
			RotateSpeed    float32
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
		}

		levelChange := LevelChange{}
		helper.FillStruct_Interface(v, &levelChange)

		if p, ok := m.Players.Load(levelChange.Id); ok {
			p.(*Player).LevelName = levelChange.LevelName

			m.SpawnProcessPlayer(p.(*Player))

			levelChangePacket := game_net.New_FRecvPacket_PlayerLevelChange()
			levelChangePacket.LevelName = p.(*Player).LevelName
			m.SendPacketToTarget(levelChangePacket, p.(*Player).Id)
		}

		if levelChange.LevelName != "SquidGame" {
			if p, ok := m.SquidPlayers.Load(levelChange.Id); ok {
				// m.SpawnSquidPlayer(p.(*SquidGamePlayer))
				// p.(*SquidGamePlayer).LevelName = levelChange.LevelName

				p.(*SquidGamePlayer).Destroy()
				m.SquidPlayers.Delete(levelChange.Id)
				delete(m.SquidGame[p.(*SquidGamePlayer).RoomIndex], levelChange.Id)
			}
		}
	}
	game_net.NetManagerInst().Callbacks["FSendPacket_HandUp"] = func(v interface{}) {
		packet := game_net.New_FRecvPacket_HandUp()
		helper.FillStruct_Interface(v, &packet)
		packet.PacketName = "FRecvPacket_HandUp"

		m.BroadCastPacketExceptMe(packet, packet.Id)
	}
	game_net.NetManagerInst().Callbacks["FSendPacket_WRCRank"] = func(v interface{}) {
		data := &database.WRCRank{}
		helper.FillStruct_Interface(v, data)

		game_net.DBManagerInst().AddWRCRank(data)
	}
	game_net.NetManagerInst().Callbacks["FSendPacket_CharacterStatus"] = func(v interface{}) {
		packet := game_net.New_FRecvPacket_CharacterStatus()
		if err := mapstructure.Decode(v, &packet); err != nil {
			fmt.Println(err)
		}

		if p, ok := m.Players.Load(packet.Id); ok {
			p.(*Player).StatusNum = packet.Status
		}
		packet.PacketName = "FRecvPacket_CharacterStatus"
		m.BroadCastPacketExceptMe(packet, packet.Id)
	}
	game_net.NetManagerInst().Callbacks["FSendPacket_StartLuckyDraw"] = func(v interface{}) {
		packet := game_net.New_FRecvPacket_StartLuckyDraw()
		m.BroadCastPacketExceptMe(packet, "GM")

		//time.Sleep(time.Second * 5)

		m.DrawingAndSendResult()
	}
	game_net.NetManagerInst().Callbacks["FSendPacket_RequestRemoveNearPlayer"] = func(v interface{}) {
		type RecvData struct {
			Id       string
			targetId string
		}
		recvdata := RecvData{}
		helper.FillStruct_Interface(v, &recvdata)

		if p, ok := m.Players.Load(recvdata.Id); ok {
			p.(*Player).RequestRemoveNearPlayers.Store(recvdata.targetId, true)
		}
	}
	game_net.NetManagerInst().Callbacks["FSendPacket_MiniGameCount"] = func(v interface{}) {
		data := database.MinigameCount{}
		if err := mapstructure.Decode(v, &data); err != nil {
			fmt.Println(err)
		}

		game_net.DBManagerInst().SyncMinigameCount(&data)
	}
	game_net.NetManagerInst().Callbacks["FSendPacket_HCoinRank"] = func(v interface{}) {
		type HCoinRank struct {
			Id string
		}

		data := &HCoinRank{}
		helper.FillStruct_Interface(v, data)

		packet := game_net.New_FRecvPacket_HCoinRank()
		packet.Rank = game_net.DBManagerInst().GetHCoinRank(data.Id)

		if p, ok := m.Players.Load(data.Id); ok {
			m.SendPacketToTarget(packet, p.(*Player).Id)
		}
	}
	game_net.NetManagerInst().Callbacks["FSendPacket_MTHText"] = func(v interface{}) {
		type MTH struct {
			Id            string
			Region        string
			Status        int32
			ScreenMessage string
		}

		data := &MTH{}
		helper.FillStruct_Interface(v, data)

		if data.Status == 0 {
			m.LMTH = append([]string{data.ScreenMessage}, m.LMTH...)
			if len(m.LMTH) > 6 {
				m.LMTH = m.LMTH[:6]
			}
		}

		if data.Status == 1 {
			m.RMTH = append([]string{data.ScreenMessage}, m.RMTH...)
			if len(m.RMTH) > 6 {
				m.RMTH = m.RMTH[:6]
			}
		}

		if data.ScreenMessage != "" {
			game_net.DBManagerInst().SaveHTM(data.Region, data.ScreenMessage)
		}

		packet := game_net.New_FRecvPacket_MTHText()
		packet.LeftMTHArr = m.LMTH
		packet.RightMTHArr = m.RMTH

		m.Players.Range(func(key, value interface{}) bool {
			m.SendPacketToTarget(packet, value.(*Player).Id)

			return true
		})
	}
	game_net.NetManagerInst().Callbacks["FSendPacket_StartAudio"] = func(v interface{}) {
		packet := game_net.New_FRecvPacket_StartAudio()
		helper.FillStruct_Interface(v, &packet)
		packet.PacketName = "FRecvPacket_StartAudio"

		m.Players.Range(func(key, value interface{}) bool {
			if value.(*Player).VoiceGroup == packet.ModeratorId {
				m.SendPacketToTarget(packet, value.(*Player).Id)
			}

			return true
		})
	}
	game_net.NetManagerInst().Callbacks["FSendPacket_StopAudio"] = func(v interface{}) {
		packet := game_net.New_FRecvPacket_StopAudio()
		helper.FillStruct_Interface(v, &packet)
		packet.PacketName = "FRecvPacket_StopAudio"

		m.Players.Range(func(key, value interface{}) bool {
			if value.(*Player).VoiceGroup == packet.ModeratorId {
				m.SendPacketToTarget(packet, value.(*Player).Id)
			}

			return true
		})
	}
	game_net.NetManagerInst().Callbacks["FSendPacket_SyncAudioForNewPlayer"] = func(v interface{}) {
		type RecvData struct {
			ModeratorId string
			TargetId    string
			AudioTime   float32
		}

		data := RecvData{}
		helper.FillStruct_Interface(v, &data)

		packet := game_net.New_FRecvPacket_SyncAudioForNewPlayer()
		packet.ModeratorId = data.ModeratorId
		packet.AudioTime = data.AudioTime

		if p, ok := m.Players.Load(data.TargetId); ok {
			fmt.Println(packet)
			m.SendPacketToTarget(packet, p.(*Player).Id)
		}
	}
	game_net.NetManagerInst().Callbacks["FSendPacket_SetVoiceVolume"] = func(v interface{}) {
		type RecvData struct {
			Id     string
			Volume float32
		}

		recvdata := RecvData{}
		helper.FillStruct_Interface(v, &recvdata)

		packet := game_net.New_FRecvPacket_SetVoiceVolume()
		packet.Id = recvdata.Id
		packet.Volume = recvdata.Volume

		m.BroadCastPacketExceptMe(packet, packet.Id)
	}
	game_net.NetManagerInst().Callbacks["FSendPacket_HeartBeat"] = func(v interface{}) {
		type RecvData struct {
			Id string
		}

		recvdata := RecvData{}
		helper.FillStruct_Interface(v, &recvdata)

		if v, ok := m.SquidPlayers.Load(recvdata.Id); ok {
			v.(*SquidGamePlayer).RecevieHearBeat = true
		}
	}
	game_net.NetManagerInst().Callbacks["FSendPacket_PlayerMove"] = func(v interface{}) {

		type Position struct {
			Id          string
			Position    transform.Vector3
			Rotation    transform.Vector3
			MoveSpeed   float32
			RotateSpeed float32
			SquidGame   bool
		}
		pos := Position{}
		helper.FillStruct_Interface(v, &pos)

		if pos.SquidGame {
			if p, ok := m.SquidPlayers.Load(pos.Id); ok {
				p.(*SquidGamePlayer).UpdateMovement(&pos.Position, &pos.Rotation, pos.MoveSpeed, pos.RotateSpeed)
			}
		} else {
			if p, ok := m.Players.Load(pos.Id); ok {
				p.(*Player).UpdateMovement(&pos.Position, &pos.Rotation, pos.MoveSpeed, pos.RotateSpeed)
			}
		}

	}

	game_net.NetManagerInst().Callbacks["FSendPacket_PlayerLogin"] = func(v interface{}) {
		data := structs.PlayerLoginInfo{}
		if err := mapstructure.Decode(v, &data); err != nil {
			fmt.Println(err)
		}

		m.NewPlayer(data.Conn, data.Data)
	}

	game_net.NetManagerInst().Callbacks["FSendPacket_DummyPosList"] = func(v interface{}) {
		type Dummy struct {
			Id string
		}

		data := Dummy{}
		helper.FillStruct_Interface(v, &data)

		packet := game_net.New_FRecvPacket_DummyPosList()

		for i := 0; i < 10; i++ {
			packet.XPositions = append(packet.XPositions, m.XPositions[i])
			packet.YPositions = append(packet.YPositions, m.YPositions[i])
		}

		if p, ok := m.Players.Load(data.Id); ok {
			m.SendPacketToTarget(packet, p.(*Player).Id)
		}
	}

	game_net.NetManagerInst().Callbacks["FSendPacket_TimeCheck"] = func(v interface{}) {
		type Time struct {
			Id string
		}

		data := Time{}
		helper.FillStruct_Interface(v, &data)

		packet := game_net.New_FRecvPacket_TimeCheck()
		packet.Hour = int32(time.Now().Hour())
		packet.Minute = int32(time.Now().Minute())
		packet.Second = int32(time.Now().Second())

		if _, ok := m.Players.Load(data.Id); ok {
			m.SendPacketToTarget(packet, data.Id)
		}
		if _, ok := m.SquidPlayers.Load(data.Id); ok {
			m.SendPacketToTarget(packet, data.Id)
		}
	}

	game_net.NetManagerInst().Callbacks["FSendPacket_SquidGameEnter"] = func(v interface{}) {
		type Enter struct {
			Id    string
			Index int32
		}

		data := Enter{}
		helper.FillStruct_Interface(v, &data)
		packet := game_net.New_FRecvPacket_SquidGameEnter()

		if len(m.SquidGame[data.Index]) < 50 {
			packet.State = true
		} else {
			packet.State = false
		}

		if p, ok := m.Players.Load(data.Id); ok {
			m.SendPacketToTarget(packet, p.(*Player).Id)
		}
	}

	game_net.NetManagerInst().Callbacks["FSendPacket_SquidGameLogin"] = func(v interface{}) {
		type Enter struct {
			Id             string
			Position       transform.Vector3
			Rotation       transform.Vector3
			IsMan          bool
			SkinIndex      int32
			TopIndex       int32
			BottomIndex    int32
			HairIndex      int32
			ShoesIndex     int32
			HairColorIndex int32
			FaceIndex      int32
			AccessoryIndex int32
			FirstName      string
			LastName       string
			BirthDay       int32
			BirthMonth     int32
			DealerType     string
			Country        string
			RoomIndex      int32
		}

		recvplayer := &SquidGamePlayer{}
		data := Enter{}
		helper.FillStruct_Interface(v, &data)

		recvplayer.Id = data.Id
		recvplayer.Position = data.Position
		recvplayer.Rotation = data.Rotation
		recvplayer.IsMan = data.IsMan
		recvplayer.SkinIndex = data.SkinIndex
		recvplayer.TopIndex = data.TopIndex
		recvplayer.BottomIndex = data.BottomIndex
		recvplayer.HairIndex = data.HairIndex
		recvplayer.ShoesIndex = data.ShoesIndex
		recvplayer.HairColorIndex = data.HairColorIndex
		recvplayer.FaceIndex = data.FaceIndex
		recvplayer.AccessoryIndex = data.AccessoryIndex
		recvplayer.FirstName = data.FirstName
		recvplayer.LastName = data.LastName
		recvplayer.BirthDay = data.BirthDay
		recvplayer.BirthMonth = data.BirthMonth
		recvplayer.DealerType = data.DealerType
		recvplayer.Country = data.Country
		recvplayer.RoomIndex = data.RoomIndex
		recvplayer.Wearable = false

		if i, ok := m.SquidPlayers.Load(recvplayer.Id); ok {
			i.(*SquidGamePlayer).Destroy()
			m.SquidPlayers.Delete(recvplayer.Id)
			delete(m.SquidGame[data.RoomIndex], data.Id)
		}

		m.SquidPlayers.Store(recvplayer.Id, recvplayer)
		m.SquidGame[data.RoomIndex][data.Id] = recvplayer

		m.SpawnSquidPlayer(recvplayer, recvplayer.RoomIndex)
		m.SendSquidPlayerNum(data.RoomIndex)
	}

	game_net.NetManagerInst().Callbacks["FSendPacket_SquidGameDie"] = func(v interface{}) {
		type Die struct {
			Id        string
			RoomIndex int32
		}

		data := Die{}
		helper.FillStruct_Interface(v, &data)

		packet := game_net.New_FRecvPacket_SquidGameDie()

		m.SquidPlayers.Range(func(key, value interface{}) bool {
			if value.(*SquidGamePlayer).RoomIndex == data.RoomIndex {
				m.SendPacketToTarget(packet, key.(string))
			}
			return true
		})

		m.SendSquidPlayerNum(data.RoomIndex)
	}

	game_net.NetManagerInst().Callbacks["FSendPacket_SquidRefresh"] = func(v interface{}) {
		type Refresh struct {
			Id string
		}

		data := Refresh{}
		helper.FillStruct_Interface(v, &data)

		packet := game_net.New_FRecvPacket_SquidRefresh()
		for i := int32(0); i < 10; i++ {
			packet.RoomPlayerNum = append(packet.RoomPlayerNum, int32(len(m.SquidGame[i])))
		}

		if _, ok := m.Players.Load(data.Id); ok {
			m.SendPacketToTarget(packet, data.Id)
		}
	}

	game_net.NetManagerInst().Callbacks["FSendPacket_SquidExit"] = func(v interface{}) {
		type RecvData struct {
			Id string
		}

		data := RecvData{}
		helper.FillStruct_Interface(v, &data)

		if v, ok := m.SquidPlayers.Load(data.Id); ok {
			// Destroy List Send
			packet := game_net.New_FRecvPacket_NearPlayerUpdate()
			spawnInfo := game_net.New_FRecvPacket_OtherPlayerDestroyInfo()
			spawnInfo.Id = data.Id
			packet.DestroyList = append(packet.DestroyList, spawnInfo)

			m.SquidPlayers.Range(func(key, value interface{}) bool {
				if value.(*SquidGamePlayer).RoomIndex == v.(*SquidGamePlayer).RoomIndex {
					m.SendPacketToTarget(packet, key.(string))
				}
				return true
			})

			//Destroy
			v.(*SquidGamePlayer).Destroy()
			m.SquidPlayers.Delete(data.Id)
			delete(m.SquidGame[v.(*SquidGamePlayer).RoomIndex], v.(*SquidGamePlayer).Id)

			//PlayerNum Refresh
			m.SendSquidPlayerNum(v.(*SquidGamePlayer).RoomIndex)
		}

	}

	game_net.NetManagerInst().Callbacks["FSendPacket_SquidPlayerNum"] = func(v interface{}) {
		type Refresh struct {
			Id        string
			RoomIndex int32
		}

		data := Refresh{}
		helper.FillStruct_Interface(v, &data)

		packet := game_net.New_FRecvPacket_SquidPlayerNum()
		packet.PlayerNum = int32(len(m.SquidGame[data.RoomIndex]))

		if _, ok := m.Players.Load(data.Id); ok {
			m.SendPacketToTarget(packet, data.Id)
		}
	}

	game_net.NetManagerInst().Callbacks["FSendPacket_FriendList"] = func(v interface{}) {
		type Firend struct {
			Id string
		}

		data := &Firend{}
		if err := mapstructure.Decode(v, data); err != nil {
			fmt.Println(err)
		}

		packet := game_net.New_FRecvPacket_FriendList()

		packet.FriendList = game_net.DBManagerInst().SearchFriendList(data.Id, m.Players)
		if _, ok := m.Players.Load(data.Id); ok {
			m.SendPacketToTarget(packet, data.Id)
		}
	}

	game_net.NetManagerInst().Callbacks["FSendPacket_FriendRequest"] = func(v interface{}) {
		type Request struct {
			Id       string
			TargetId string
		}
		data := &Request{}
		if err := mapstructure.Decode(v, data); err != nil {
			fmt.Println(err)
		}

		packet := game_net.New_FRecvPacket_FriendRequest()
		packet.Name = game_net.DBManagerInst().FriendRequest(data.Id)
		packet.TargetId = data.Id

		if _, ok := m.Players.Load(data.Id); ok {
			m.SendPacketToTarget(packet, data.TargetId)
		}
	}

	game_net.NetManagerInst().Callbacks["FSendPacket_AcceptFriend"] = func(v interface{}) {
		type Accept struct {
			Id       string
			TargetId string
			BeFriend bool
		}
		data := &Accept{}
		if err := mapstructure.Decode(v, data); err != nil {
			fmt.Println(err)
		}

		if data.BeFriend {
			if p, ok := m.Players.Load(data.TargetId); ok {
				game_net.DBManagerInst().SetFriend(data.TargetId, data.Id, p.(*Player).FirstName+p.(*Player).LastName)
			}
			if p, ok := m.Players.Load(data.Id); ok {
				game_net.DBManagerInst().SetFriend(data.Id, data.TargetId, p.(*Player).FirstName+p.(*Player).LastName)
			}
		}
	}

	game_net.NetManagerInst().Callbacks["FSendPacket_FriendSearch"] = func(v interface{}) {
		type Search struct {
			Id          string
			SearchInput string
		}
		data := &Search{}
		if err := mapstructure.Decode(v, data); err != nil {
			fmt.Println(err)
		}

		packet := game_net.New_FRecvPacket_FriendSearch()

		packet.FriendSearchArray = game_net.DBManagerInst().SearchList(data.SearchInput, m.Players)
		if _, ok := m.Players.Load(data.Id); ok {
			m.SendPacketToTarget(packet, data.Id)
		}
	}

	game_net.NetManagerInst().Callbacks["FSendPacket_FriendInfo"] = func(v interface{}) {
		type Search struct {
			Id       string
			FriendId string
		}
		data := &Search{}
		if err := mapstructure.Decode(v, data); err != nil {
			fmt.Println(err)
		}

		packet := game_net.New_FRecvPacket_FriendInfo()

		packet = game_net.DBManagerInst().SetFriendInfo(data.Id, data.FriendId)
		if _, ok := m.Players.Load(data.Id); ok {
			m.SendPacketToTarget(packet, data.Id)
		}
	}
}

func (m *ObjectManager) SpawnProcessPlayer(_p *Player) {
	distSq := 4000.0 * 4000.0

	newNearPlayers := map[string]*Player{}
	removePlayers := map[string]*Player{}
	spawnPlayers := map[string]*Player{}

	newNearBoxs := map[string]*TreasureBox{}
	removeBoxs := map[string]*TreasureBox{}
	spawnBoxs := map[string]*TreasureBox{}

	_p.RequestRemoveNearPlayers.Range(func(key, value interface{}) bool {
		_p.NearPlayers.Delete(key)
		return true
	})
	_p.RequestRemoveNearPlayers = sync.Map{}

	_, SquidExist := m.SquidPlayers.Load(_p.Id)

	if !SquidExist {
		_p.NearPlayers.Range(func(key, value interface{}) bool {

			if _p.Id != value.(*Player).Id {
				diffVec := transform.Sub3(_p.Position, value.(*Player).Position)
				if distSq > diffVec.LengthSq() && _p.LevelName == value.(*Player).LevelName {
					newNearPlayers[value.(*Player).Id] = value.(*Player)
				} else {
					removePlayers[value.(*Player).Id] = value.(*Player)
				}
			}

			return true
		})

		m.Players.Range(func(key, value interface{}) bool {
			if _p.Id != value.(*Player).Id {
				diffVec := transform.Sub3(_p.Position, value.(*Player).Position)
				if distSq > diffVec.LengthSq() && _p.LevelName == value.(*Player).LevelName {
					_, ok := newNearPlayers[value.(*Player).Id]
					if !ok {
						newNearPlayers[value.(*Player).Id] = value.(*Player)
						spawnPlayers[value.(*Player).Id] = value.(*Player)
					}
				}
			}
			return true
		})
	}

	/////////////////////////////////////////////////////////

	_p.NearBoxs.Range(func(key, value interface{}) bool {
		diffVec := transform.Sub3(_p.Position, value.(*TreasureBox).Position)
		if distSq > diffVec.LengthSq() && _p.LevelName == value.(*TreasureBox).LevelName {
			newNearBoxs[value.(*TreasureBox).Id] = value.(*TreasureBox)
		} else {
			removeBoxs[value.(*TreasureBox).Id] = value.(*TreasureBox)
		}

		return true
	})

	m.TreasureBoxs.Range(func(key, value interface{}) bool {
		diffVec := transform.Sub3(_p.Position, value.(*TreasureBox).Position)
		if distSq > diffVec.LengthSq() && _p.LevelName == value.(*TreasureBox).LevelName {
			_, ok := newNearBoxs[value.(*TreasureBox).Id]
			if !ok {
				newNearBoxs[value.(*TreasureBox).Id] = value.(*TreasureBox)
				spawnBoxs[value.(*TreasureBox).Id] = value.(*TreasureBox)
			}
		}
		return true
	})

	packet := game_net.New_FRecvPacket_NearPlayerUpdate()
	for _, item := range spawnPlayers {
		spawnInfo := game_net.New_FRecvPacket_OtherPlayerSpawnInfo()
		spawnInfo.Id = item.Id
		spawnInfo.Position = item.Position
		spawnInfo.Rotation = item.Rotation
		spawnInfo.IsMan = item.IsMan
		spawnInfo.SkinIndex = item.SkinIndex
		spawnInfo.TopIndex = item.TopIndex
		spawnInfo.BottomIndex = item.BottomIndex
		spawnInfo.HairIndex = item.HairIndex
		spawnInfo.ShoesIndex = item.ShoesIndex
		spawnInfo.HairColorIndex = item.HairColorIndex
		spawnInfo.FaceIndex = item.FaceIndex
		spawnInfo.AccessoryIndex = item.AccessoryIndex

		spawnInfo.FirstName = item.FirstName
		spawnInfo.LastName = item.LastName
		spawnInfo.BirthDay = item.BirthDay
		spawnInfo.BirthMonth = item.BirthMonth
		spawnInfo.DealerType = item.DealerType
		spawnInfo.Country = item.Country

		spawnInfo.IsWearableWear = item.IsWearableWear
		spawnInfo.IsDancing = item.IsDancing
		spawnInfo.IsSitting = item.IsSitting
		spawnInfo.StatusNum = item.StatusNum

		spawnInfo.IsAtlasSpawn = item.IsAtlasSpawn
		spawnInfo.AtlasPosition = item.AtlasPosition
		spawnInfo.AtlasRotation = item.AtlasRotation

		spawnInfo.IsSpotSpawn = item.IsSpotSpawn
		spawnInfo.SpotPosition = item.SpotPosition
		spawnInfo.SpotRotation = item.SpotRotation

		packet.SpawnList = append(packet.SpawnList, spawnInfo)
	}

	for _, item := range removePlayers {
		spawnInfo := game_net.New_FRecvPacket_OtherPlayerDestroyInfo()
		spawnInfo.Id = item.Id
		packet.DestroyList = append(packet.DestroyList, spawnInfo)
	}

	_p.NearPlayers = sync.Map{}

	for _, item := range newNearPlayers {
		_p.NearPlayers.Store(item.Id, item)
	}

	e, err := json.Marshal(packet)

	if err != nil {
		log.Fatal("Parse Error")
		return
	}

	if len(packet.SpawnList) > 0 || len(packet.DestroyList) > 0 {
		game_net.NetManagerInst().SendString(_p.GetTCPContext(), string(e))
	}

	boxpacket := game_net.New_FRecvPacket_TreasureBoxInfo()
	for _, boxitem := range spawnBoxs {
		spawnBoxInfo := game_net.New_FRecvPacket_CreateTreasureBox()
		spawnBoxInfo.Id = boxitem.Id
		spawnBoxInfo.Point = boxitem.Point
		spawnBoxInfo.Position = boxitem.Position
		boxpacket.BoxList = append(boxpacket.BoxList, spawnBoxInfo)
	}

	for _, boxitem := range removeBoxs {
		spawnBoxInfo := game_net.New_FRecvPacket_DestroyTreasureBox()
		spawnBoxInfo.Id = boxitem.Id
		boxpacket.DestroyList = append(boxpacket.DestroyList, spawnBoxInfo)
	}

	_p.NearBoxs = sync.Map{}

	for _, item := range newNearBoxs {
		_p.NearBoxs.Store(item.Id, item)
	}

	_e, _err := json.Marshal(boxpacket)

	if _err != nil {
		log.Fatal("Parse Error")
		return
	}

	if len(boxpacket.BoxList) > 0 || len(boxpacket.DestroyList) > 0 {
		game_net.NetManagerInst().SendString(_p.GetTCPContext(), string(_e))
	}

}

func (m *ObjectManager) BroadCastPacketToAll(s interface{}) {

	packet := reflect.ValueOf(s).Interface()

	e, err := json.Marshal(packet)
	if err != nil {
		log.Fatal("Parse Error")
		return
	}

	m.Players.Range(func(key, value interface{}) bool {
		game_net.NetManagerInst().SendString(value.(*Player).GetTCPContext(), string(e))
		return true
	})
}

func (m *ObjectManager) BroadCastPacketExceptMe(s interface{}, Id string) {

	packet := reflect.ValueOf(s).Interface()

	e, err := json.Marshal(packet)
	if err != nil {
		log.Fatal("Parse Error")
		return
	}

	m.Players.Range(func(key, value interface{}) bool {
		if value.(*Player).Id != Id {
			if _, ok := m.Players.Load(value.(*Player).Id); ok {
				game_net.NetManagerInst().SendString(value.(*Player).GetTCPContext(), string(e))
			}
		}

		return true
	})
}

func (m *ObjectManager) SendPacketToTarget(s interface{}, Id string) {

	packet := reflect.ValueOf(s).Interface()

	e, err := json.Marshal(packet)
	if err != nil {
		log.Fatal("Parse Error")
		return
	}

	if p, ok := m.Players.Load(Id); ok {
		game_net.NetManagerInst().SendString(p.(*Player).GetTCPContext(), string(e))
	}
}

func (m *ObjectManager) SpawnProcess() {
	m.Players.Range(func(key, value interface{}) bool {
		m.SpawnProcessPlayer(value.(*Player))
		return true
	})
}

func (m *ObjectManager) SquidGameSpawnProcess() {
	m.SquidPlayers.Range(func(key, value interface{}) bool {
		m.SpawnSquidPlayer(value.(*SquidGamePlayer), value.(*SquidGamePlayer).RoomIndex)
		return true
	})
}

func (m *ObjectManager) NewPlayer(conn *net.TCPConn, data map[string]interface{}) {
	m.TotalLoginUserCount += 1

	playerCount := 0
	m.Players.Range(func(key, value interface{}) bool {
		playerCount++
		return true
	})

	log.Println("NewPlayer Entered. Total Count of Online Players =", playerCount+1)
	log.Println("Total Login User Count =", m.TotalLoginUserCount)

	recvplayer := &Player{}
	if err := mapstructure.Decode(data, recvplayer); err != nil {
		fmt.Println(err)
	}

	if i, ok := m.Players.Load(recvplayer.Id); ok {
		i.(*Player).Destroy()
		m.Players.Delete(recvplayer.Id)
	}

	recvplayer.context = conn
	recvplayer.IsDancing = false
	recvplayer.IsSitting = false
	recvplayer.StatusNum = 0
	recvplayer.IsWearableWear = false

	m.Players.Store(recvplayer.Id, recvplayer)

	m.SpawnProcessPlayer(recvplayer)

	packet := game_net.New_FRecvPacket_LoginInfo()

	carInfoPacket := game_net.New_FRecvPacket_CarInfos()

	m.Cars.Range(func(key, value interface{}) bool {
		pInfo := game_net.New_FRecvPacket_OtherCarSpawnInfo()
		pInfo.Id = value.(*Car).Id
		pInfo.PathTag = value.(*Car).PathTag
		pInfo.TypeNum = value.(*Car).TypeNum
		pInfo.ServerDistacne = value.(*Car).Distance

		carInfoPacket.List = append(carInfoPacket.List, pInfo)
		return true
	})

	{
		e, err := json.Marshal(carInfoPacket)

		if err != nil {
			log.Fatal("Parse Error")
			return
		}

		game_net.NetManagerInst().SendString(recvplayer.GetTCPContext(), string(e))
	}

	packet.Position = recvplayer.Position

	e, err := json.Marshal(packet)

	if err != nil {
		log.Fatal("Parse Error")
		return
	}

	game_net.NetManagerInst().SendString(recvplayer.GetTCPContext(), string(e))
	//	go m.RunHeartBeatProcess(recvplayer.Id, recvplayer.GetTCPContext(), false)
	//	go m.RunHeartBeatProcess(recvplayer.Id, recvplayer.GetTCPContext(), true)

}

func (m *ObjectManager) AddNewTreasureBox(id string, point int32, position transform.Vector3, levelname string) {

	newBox := &TreasureBox{}
	newBox.Id = id
	newBox.Point = point
	newBox.Position = position
	newBox.LevelName = levelname

	m.TreasureBoxs.Store(id, newBox)

	if _, ok := m.Players.Load("GM"); ok {
		// 즉시 업데이트
		m.Players.Range(func(key, value interface{}) bool {

			if newBox.LevelName == value.(*Player).LevelName {
				m.SpawnProcessPlayer(value.(*Player))
			}

			return true
		})
	}
}

func (m *ObjectManager) AddCar(id string, typeNum int32, pathTag string) {

	newCar := &Car{}
	newCar.Id = id
	newCar.PathTag = pathTag
	newCar.TypeNum = typeNum
	newCar.Distance = 0

	m.Cars.Store(id, newCar)
}

func (m *ObjectManager) RemoveCar(id string) {
	m.Cars.Delete(id)
}

func (m *ObjectManager) UpdateCar(id string, distance float64) {
	if v, ok := m.Cars.Load(id); ok {
		v.(*Car).Distance = distance
	}
}

func (m *ObjectManager) BroadCastWRCRank() {

	packet := game_net.DBManagerInst().BroadCastWRCRank()

	e, err := json.Marshal(packet)
	if err != nil {
		log.Fatal("Parse Error")
		return
	}

	m.Players.Range(func(key, value interface{}) bool {
		game_net.NetManagerInst().SendString(value.(*Player).GetTCPContext(), string(e))
		return true
	})

}

// func (m *ObjectManager) BroadCastLoveForest() {
// 	packet := game_net.New_FRecvPacket_LoveForest()
// 	packet.Progress = game_net.DBManagerInst().GetLoveForestInfo()
// 	m.BroadCastPacketToAll(packet)
// }

func (m *ObjectManager) DrawingAndSendResult() {

	playerCount := 0
	m.Players.Range(func(key, value interface{}) bool {
		playerCount++
		return true
	})

	totaluser := playerCount

	var drawList []string
	var winner []string
	var picked map[int]bool = make(map[int]bool)

	if totaluser < 5 {
		fmt.Println("Total user less than 5. Can't Run Lucky Draw Event")
		return
	} else {
		m.Players.Range(func(key, value interface{}) bool {
			if key != "GM" {
				drawList = append(drawList, key.(string))
			}

			return true
		})
		for i := 0; i < 5; i++ {
			picknum := helper.GetUniqueRandom(picked, totaluser-1)
			winner = append(winner, drawList[picknum])
		}
	}

	drawpacket := game_net.New_FRecvPacket_LuckyDrawWinner()
	drawpacket.Reward = 5000
	winnerpacket := game_net.New_FRecvPacket_StartLuckyDraw()

	for _, i := range winner {
		winnercarnum := game_net.DBManagerInst().GetCarNum(i)
		winnerpacket.LuckyNumbers = append(winnerpacket.LuckyNumbers, winnercarnum)
	}
	m.BroadCastPacketExceptMe(winnerpacket, "GM")
}

func (m *ObjectManager) GetIdByConn(c *net.TCPConn) string {

	result := ""
	m.Players.Range(func(key, value interface{}) bool {
		if value.(*Player).GetTCPContext() == c {
			result = value.(*Player).Id

			return false
		}

		return true
	})
	return result
}

func (m *ObjectManager) LeaveVoiceGroup(voiceGroupId string) {

	if _p, ok := m.Players.Load(voiceGroupId); ok {
		if p, ok := m.VoiceGroup.Load(_p.(*Player).VoiceGroup); ok {
			p.(*sync.Map).Delete(_p.(*Player).Id)

			VoiceKey := _p.(*Player).VoiceGroup
			_p.(*Player).VoiceGroup = ""

			keys := make([]string, 0, 0)

			// 방장이 나가서 방이 없어진다.
			// 방장 소유의 방만 파괴된다
			// if game_net.DBManagerInst().IsModerator(voiceGroupId) && VoiceKey == voiceGroupId {

			// 	if vData, ok := m.VoiceGroup.Load(VoiceKey); ok {
			// 		vData.(*sync.Map).Range(func(key, value interface{}) bool {
			// 			if __p, ok := m.Players.Load(key.(string)); ok {
			// 				__p.(*Player).VoiceGroup = ""
			// 			}
			// 			return true
			// 		})
			// 		vData.(*sync.Map).Delete(VoiceKey)
			// 	}
			// 	// 나만 나갔다.
			// } else {
			// 	//맨앞에 VoiceKey 즉 Moderator
			// 	keys = append(keys, VoiceKey)

			// 	p.(*sync.Map).Range(func(k, v interface{}) bool {
			// 		if k.(string) != VoiceKey {
			// 			keys = append(keys, k.(string))
			// 		}
			// 		return true
			// 	})
			// }

			pac := game_net.New_FRecvPacket_LeaveVoiceGroupUpdate()

			Ids := game_net.ModeratorUserInfo{}

			for _, key := range keys {
				if key == "" {
					continue
				}

				Ids.Id = key
				Ids.IsModerator = key == VoiceKey
				Ids.URL = "https://h-festival.s3.ap-northeast-2.amazonaws.com/images/T_HD.png"

				// Leave시에 아이디 안나오는 문제 수정
				if i, ok := m.Players.Load(Ids.Id); ok {
					Ids.UserName = i.(*Player).FirstName + " " + i.(*Player).LastName
				}

				if data, ok := p.(*sync.Map).Load(key); ok {
					Ids.OnHandsUp = data.(*VoiceUserState).OnHandsUp
					Ids.OnMic = data.(*VoiceUserState).OnMic
				}

				pac.Ids = append(pac.Ids, Ids)
			}

			e, err := json.Marshal(pac)
			if err != nil {
				log.Fatal("Parse Error")
				return
			}

			p.(*sync.Map).Range(func(key, value interface{}) bool {
				if _p, ok := m.Players.Load(key.(string)); ok {
					game_net.NetManagerInst().SendString(_p.(*Player).GetTCPContext(), string(e))
				}

				return true
			})
		}
	}
}

func (m *ObjectManager) GetSQL() *sql.DB {
	dbs, err := sql.Open("mysql", "root:0000@tcp(127.0.0.1:3306)/hmc")
	if err != nil {
		panic(err)
		log.Printf("SQL_Open Fail : %s", err)
	}
	return dbs
}

func (m *ObjectManager) DummyPosInit() {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	for i := 0; i < 10; i++ {
		m.XPositions = append(m.XPositions, r1.Int31n(100))
		m.YPositions = append(m.YPositions, r1.Int31n(100))
	}
}

func (m *ObjectManager) DummyPosChange() {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	Index := r1.Int31n(10)
	Xpos := r1.Int31n(100)
	Ypos := r1.Int31n(100)
	m.XPositions[Index] = Xpos
	m.YPositions[Index] = Ypos

	packet := game_net.New_FRecvPacket_DummyPosSet()
	packet.ListIndex = Index
	packet.XPosition = Xpos
	packet.YPosition = Ypos

	m.BroadCastPacketToAll(packet)
}

func (m *ObjectManager) SpawnSquidPlayer(_p *SquidGamePlayer, RoomIndex int32) {
	distSq := 100000.0 * 100000.0

	newNearPlayers := map[string]*SquidGamePlayer{}
	removePlayers := map[string]*SquidGamePlayer{}
	spawnPlayers := map[string]*SquidGamePlayer{}

	_p.RequestRemoveNearPlayers.Range(func(key, value interface{}) bool {
		_p.NearPlayers.Delete(key)
		return true
	})
	_p.RequestRemoveNearPlayers = sync.Map{}

	_p.NearPlayers.Range(func(key, value interface{}) bool {

		if _p.Id != value.(*SquidGamePlayer).Id {
			diffVec := transform.Sub3(_p.Position, value.(*SquidGamePlayer).Position)
			if distSq > diffVec.LengthSq() || _p.LevelName == value.(*SquidGamePlayer).LevelName {
				newNearPlayers[value.(*SquidGamePlayer).Id] = value.(*SquidGamePlayer)
			} else {
				removePlayers[value.(*SquidGamePlayer).Id] = value.(*SquidGamePlayer)
			}
		}

		return true
	})

	m.SquidPlayers.Range(func(key, value interface{}) bool {
		if _p.Id != value.(*SquidGamePlayer).Id && _p.RoomIndex == value.(*SquidGamePlayer).RoomIndex {
			diffVec := transform.Sub3(_p.Position, value.(*SquidGamePlayer).Position)
			if distSq > diffVec.LengthSq() {
				_, ok := newNearPlayers[value.(*SquidGamePlayer).Id]
				if !ok {
					newNearPlayers[value.(*SquidGamePlayer).Id] = value.(*SquidGamePlayer)
					spawnPlayers[value.(*SquidGamePlayer).Id] = value.(*SquidGamePlayer)
				}
			}
		}
		return true
	})

	/////////////////////////////////////////////////////////

	packet := game_net.New_FRecvPacket_NearPlayerUpdate()
	for _, item := range spawnPlayers {
		spawnInfo := game_net.New_FRecvPacket_OtherPlayerSpawnInfo()
		spawnInfo.Id = item.Id
		spawnInfo.Position = item.Position
		spawnInfo.Rotation = item.Rotation
		spawnInfo.IsMan = item.IsMan
		spawnInfo.SkinIndex = item.SkinIndex
		spawnInfo.TopIndex = item.TopIndex
		spawnInfo.BottomIndex = item.BottomIndex
		spawnInfo.HairIndex = item.HairIndex
		spawnInfo.ShoesIndex = item.ShoesIndex
		spawnInfo.HairColorIndex = item.HairColorIndex
		spawnInfo.FaceIndex = item.FaceIndex
		spawnInfo.AccessoryIndex = item.AccessoryIndex

		spawnInfo.FirstName = item.FirstName
		spawnInfo.LastName = item.LastName
		spawnInfo.BirthDay = item.BirthDay
		spawnInfo.BirthMonth = item.BirthMonth
		spawnInfo.DealerType = item.DealerType
		spawnInfo.Country = item.Country

		packet.SpawnList = append(packet.SpawnList, spawnInfo)
	}

	for _, item := range removePlayers {
		spawnInfo := game_net.New_FRecvPacket_OtherPlayerDestroyInfo()
		spawnInfo.Id = item.Id
		packet.DestroyList = append(packet.DestroyList, spawnInfo)
	}

	_p.NearPlayers = sync.Map{}

	//NearPlayerUpdate Store
	for _, item := range newNearPlayers {
		_p.NearPlayers.Store(item.Id, item)
	}

	if len(packet.SpawnList) > 0 || len(packet.DestroyList) > 0 {
		m.SendPacketToTarget(packet, _p.Id)
	}
}

func (m *ObjectManager) SquidGameTimeCheck() {
	go func() {
		for {
			CurrentTime := time.Now().Minute()
			if m.SquidGameMinute != CurrentTime {
				m.SquidGameMinute = CurrentTime
				if CurrentTime%5 == 3 {
					packet := game_net.New_FRecvPacket_SquidGameStart()

					s1 := rand.NewSource(time.Now().UnixNano())
					r1 := rand.New(s1)

					for i := int32(0); i < 10; i++ {
						packet.DieSpeed = nil
						packet.DummyFate = nil
						packet.SoundSpeed = nil

						NPCLen := 50 - len(m.SquidGame[i])
						//	m.SquidGame[]
						for i := 0; i < NPCLen; i++ {
							packet.DieSpeed = append(packet.DieSpeed, r1.Int31n(10))
							packet.DummyFate = append(packet.DummyFate, r1.Int31n(4)+1)
						}
						for i := 0; i < 50; i++ {
							packet.SoundSpeed = append(packet.SoundSpeed, r1.Int31n(10))
						}

						m.SquidPlayers.Range(func(key, value interface{}) bool {
							if value.(*SquidGamePlayer).RoomIndex == i {
								m.SendPacketToTarget(packet, key.(string))
							}
							return true
						})
					}

				}
				if CurrentTime%5 == 0 {
					// Room People Reset
					//m.SquidGame = make(map[int32]map[string]interface{})
					//m.SquidPlayers = sync.Map{}
					m.SquidGame = map[int32]map[string]interface{}{}
					for i := int32(0); i < 10; i++ {
						m.SquidGame[i] = map[string]interface{}{}
					}
					m.SquidPlayers.Range(func(Key, Value interface{}) bool {
						m.SquidPlayers.Delete(Key)
						return true
					})
				}
			}
		}
	}()
}

func (m *ObjectManager) SendSquidPlayerNum(RoomIndex int32) {
	packet := game_net.New_FRecvPacket_SquidPlayerNum()
	packet.PlayerNum = int32(len(m.SquidGame[RoomIndex]))

	m.SquidPlayers.Range(func(key, value interface{}) bool {
		if value.(*SquidGamePlayer).RoomIndex == RoomIndex {
			m.SendPacketToTarget(packet, key.(string))
		}
		return true
	})
}
