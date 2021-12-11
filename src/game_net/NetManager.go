package game_net

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"syscall"

	"github.com/mitchellh/mapstructure"
	"hmc_server.com/hmc_server/src/database"
	"hmc_server.com/hmc_server/src/helper"
	"hmc_server.com/hmc_server/src/structs"
)

type Test interface {
	Testfun(string) string
}

type NetManager struct {
	receviers         map[string]map[string]chan interface{}
	contextReceviers  sync.Map
	DeletePlayerChan  chan *net.TCPConn
	ContextPlayerChan chan *net.TCPConn
	Callbacks         map[string]func(interface{})
	SettingPlayerFunc func(player *net.TCPConn)
}

func (nm *NetManager) AddContextRecevier(packetName string, uniqueName *net.TCPConn, c chan interface{}) {
	if val, ok := nm.contextReceviers.Load(uniqueName.RemoteAddr().String()); ok {
		if val2, ok2 := val.(*sync.Map).Load(packetName); ok2 {
			close(val2.(chan interface{}))
		}
		val.(*sync.Map).Store(packetName, c)
	} else {
		syncMap := &sync.Map{}
		syncMap.Store(packetName, c)
		nm.contextReceviers.Store(uniqueName.RemoteAddr().String(), syncMap)
	}
}

func (nm *NetManager) RemoveContextRecevier(uniqueName *net.TCPConn) {
	if v2, ok := nm.contextReceviers.Load(uniqueName.RemoteAddr().String()); ok {
		if v3, ok3 := v2.(*sync.Map).Load("StopGoroutine"); ok3 {
			v3.(chan interface{}) <- true
		}
	}

	if val, ok := nm.contextReceviers.Load(uniqueName.RemoteAddr().String()); ok {
		val.(*sync.Map).Range(func(key, value interface{}) bool {
			close(value.(chan interface{}))
			return true
		})

		nm.contextReceviers.Delete(uniqueName.RemoteAddr().String())
	}
}

func (nm *NetManager) AddRecevier(packetName string, uniqueName string, c chan interface{}) {
	if val, ok := nm.receviers[packetName]; ok {
		if val2, ok2 := val[uniqueName]; ok2 {
			close(val2)
		}
		val[uniqueName] = c
	} else {
		nm.receviers[packetName] = map[string]chan interface{}{}
		nm.receviers[packetName][uniqueName] = c
	}
}

func (nm *NetManager) RemoveRecevier(packetName string, uniqueName string) {
	if val, ok := nm.receviers[packetName]; ok {
		if val2, ok2 := val[uniqueName]; ok2 {
			// 기존 channel이 있으면 그냥 close하고 덮어 쒸운다.
			close(val2)
			delete(val, uniqueName)
		}
	}
}

var instance *NetManager
var once sync.Once

func NetManagerInst() *NetManager {
	once.Do(func() {
		instance = &NetManager{}
		instance.receviers = map[string]map[string]chan interface{}{}
		instance.contextReceviers = sync.Map{}
		instance.DeletePlayerChan = make(chan *net.TCPConn)
		instance.ContextPlayerChan = make(chan *net.TCPConn)
		instance.Callbacks = make(map[string]func(interface{}))
	})
	return instance
}

func (nm *NetManager) Init() {
	fmt.Println("INIT_NetManager")
	addAddr, err := net.ResolveTCPAddr("tcp", ":8002")
	if err != nil {
	}

	ln, err := net.ListenTCP("tcp", addAddr) // TCP 프로토콜에 8000 포트로 연결을 받음

	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		conn, err := ln.AcceptTCP() // 클라이언트가 연결되면 TCP 연결을 리턴
		conn.SetReadBuffer(4096)
		conn.SetWriteBuffer(2 * 1024 * 1024)

		if err != nil {
			fmt.Println(err)
			continue
		}

		newConns := make(chan net.TCPConn)
		go func() {
			//for {

			cConnData, _ := <-newConns

			cConn := &cConnData
			//buffer := bytes.Buffer{} //버퍼

			log.Printf("client connect : %v", cConn)
			buffered := make([]byte, 0)
			count_jm := int(0)
			//count_jm := int(0)
			//다 받을때까지 반복하며 읽음

			// 플레이어 접속과 동시에 컨텍스트 건다.
			nm.SettingPlayerFunc(cConn)

			//objects.ObjectManagerInst().RunPlayerContent(cConn)
			for {
				bufConn := bufio.NewReaderSize(cConn, 4096)
				_, err1 := bufConn.Peek(1)

				if err1 != nil {
					if io.EOF == err1 {
						nm.DeletePlayerChan <- cConn
						(*cConn).Close()

						return
					} else {
						// nm.DeletePlayerChan <- cConn
						// nm.RemoveContextRecevier(cConn)
						// (*cConn).Close()
						// log.Printf("connection Fail: %v", err1)

						continue
					}

				}

				buffredLength := bufConn.Buffered()

				bufferedPeek, err2 := bufConn.Peek(buffredLength)

				if err2 != nil {
					if io.EOF == err2 {
						nm.DeletePlayerChan <- cConn
						(*cConn).Close()

						return
					} else {
						// nm.DeletePlayerChan <- cConn
						// nm.RemoveContextRecevier(cConn)
						// (*cConn).Close()
						// log.Printf("connection Fail2: %v", err2)

						continue
					}
				}

				// fmt.Println(string(bufferedPeek))

				buffered = append(buffered, bufferedPeek...)

				splitedStrs := strings.Split(string(buffered), "\n")

				buffered = make([]byte, 0)

				for _, v := range splitedStrs {
					if len(v) == 0 {
						continue
					}
					var dat map[string]interface{}
					byteData := []byte(v)
					if unmarshalErr := json.Unmarshal(byteData, &dat); unmarshalErr != nil {

						//log.Println(v)

						if v[len(v)-1] == '}' {
							reUseData := []byte(v + "\n")
							buffered = append(buffered, reUseData...)
						} else {
							reUseData := []byte(v)
							buffered = append(buffered, reUseData...)
						}
					} else {
						packetName := dat["packetName"].(string)

						//fmt.Println(packetName)

						// PlayerActionEvent 씹히는거 체크
						//if packetName == "FSendPacket_PlayerActionEvent" {
						//log.Println("recevie Packet")
						// debug

						// if packetName != "FSendPacket_PlayerMove" && packetName != "FSendPacket_Voice" {
						// 	log.Println(packetName)
						// }
						//}

						if packetName == "FSendPacket_TestPacket" {
							type CountStruct struct {
								Count int64
								Trash string
							}

							countStruct := &CountStruct{}

							json.Unmarshal(byteData, &countStruct)

							//log.Println(countStruct.Count, count_jm)

							if countStruct.Count != int64(count_jm) {
								log.Fatalln("mismatch")
							}

							count_jm++
						}

						if v, ok := nm.Callbacks[packetName]; ok {
							if packetName == "FSendPacket_PlayerLogin" {
								loginInfo := structs.PlayerLoginInfo{}
								loginInfo.Conn = cConn
								loginInfo.Data = dat
								v(loginInfo)
							} else if packetName == "FSendPacket_DBSignin" {
								type signin struct {
									Conn     *net.TCPConn
									Playerid string
									Password string
								}
								presignin := &signin{}
								if err := mapstructure.Decode(dat, presignin); err != nil {
									fmt.Println(err)
								}

								log.Println("DBSignin Player :", presignin.Playerid)

								presignin.Conn = cConn

								v(presignin)
							} else {
								v(dat)
							}

							continue
						}

						for _, v2 := range nm.receviers[packetName] {
							if packetName == "FSendPacket_DBSignup" {
								presignup := database.PlayerInfo{}
								helper.FillStruct_Interface(dat, &presignup)
								presignup.Conn = cConn

								v2 <- presignup
							} else if packetName == "FSendPacket_FindPassword" {
								type findpw struct {
									Id   string
									Conn *net.TCPConn
								}
								fpwdata := findpw{}
								helper.FillStruct_Interface(dat, &fpwdata)
								fpwdata.Conn = cConn

								v2 <- fpwdata
							} else {
								v2 <- dat
							}
						}

						if v2, ok := nm.contextReceviers.Load(cConn.RemoteAddr().String()); ok {
							if v3, ok3 := v2.(*sync.Map).Load(packetName); ok3 {
								if packetName == "FSendPacket_PlayerLogin" {
									loginInfo := structs.PlayerLoginInfo{}
									loginInfo.Conn = cConn
									loginInfo.Data = dat
									v3.(chan interface{}) <- loginInfo
								} else {
									if !nm.IsChanClosed(v3.(chan interface{})) {
										v3.(chan interface{}) <- dat
									}
								}
							}
						}
					}
				}
			}
		}()

		newConns <- *conn
	}
}

func (nm *NetManager) IsChanClosed(ch <-chan interface{}) bool {
	select {
	case <-ch:
		return true
	default:
	}
	return false
}

func (nm *NetManager) SendInterface(c *net.TCPConn, i interface{}) {
	e, err := json.Marshal(i)
	if err != nil {
		return
	}

	nm.SendString(c, string(e))
}

func (nm *NetManager) SendString(c *net.TCPConn, str string) {
	go func() {
		if c != nil {
			sendData := []byte(str)
			sendLength := uint32(len(sendData))
			sizeBytes := make([]byte, 4)

			binary.LittleEndian.PutUint32(sizeBytes, sendLength)

			sizeBytes = append(sizeBytes, sendData...)

			sent, err := (*c).Write(sizeBytes)
			if err != nil {
				if errors.Is(err, syscall.EPIPE) {
					//log.Println("error EPIPE : ", err)
				} else {
					//log.Println("error send : ", err, str)

					if sent > 0 {
						if sent != len(sizeBytes) {
							log.Println("sent diffrent sent : ", sent, "sizeBytes : ", len(sizeBytes))
						}
					}
				}
			}
		}
	}()
}
