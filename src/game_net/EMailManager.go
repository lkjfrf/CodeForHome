package game_net

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/smtp"
	"strconv"
	"sync"
	"time"

	"github.com/mitchellh/mapstructure"
)

type EmailManager struct {
	VerifyCodes map[string]string
	Conns       map[string]*net.TCPConn
}

var Instance_Email *EmailManager
var once_Email sync.Once

func EmailManagerInst() *EmailManager {
	once_Email.Do(func() {
		Instance_Email = &EmailManager{}
	})
	return Instance_Email
}

func (em *EmailManager) Init() {
	fmt.Println("INIT_EmailManager")

	em.VerifyCodes = make(map[string]string)
	em.Conns = make(map[string]*net.TCPConn)

	// NetManagerInst().Callbacks["FSendPacket_FindPassword"] = func(v interface{}) {
	// 	type findpw struct {
	// 		Id   string
	// 		Conn *net.TCPConn
	// 	}

	// 	data := &findpw{}
	// 	if err := mapstructure.Decode(v, data); err != nil {
	// 		fmt.Println(err)
	// 	}

	// 	packet := New_FRecvPacket_FindPassword()

	// 	if DBManagerInst().CheckEmail(data.Id) {
	// 		packet.Status = true
	// 	} else {
	// 		packet.Status = false
	// 	}

	// 	if packet.Status {
	// 		em.VerifyCodes[data.Id] = GenerateVerifyCode()
	// 		em.Conns[data.Id] = data.Conn

	// 		e, err := json.Marshal(packet)
	// 		if err != nil {
	// 			log.Fatal("Parse Error")
	// 			return
	// 		}

	// 		NetManagerInst().SendString(data.Conn, string(e))

	// 		em.SendEmail(data.Id)
	// 	} else {

	// 		e, err := json.Marshal(packet)
	// 		if err != nil {
	// 			log.Fatal("Parse Error")
	// 			return
	// 		}

	// 		NetManagerInst().SendString(data.Conn, string(e))
	// 	}
	// }
	NetManagerInst().Callbacks["FSendPacket_VerifyCode"] = func(v interface{}) {
		type Verify struct {
			Id         string
			VerifyCode string
		}

		data := &Verify{}
		if err := mapstructure.Decode(v, data); err != nil {
			fmt.Println(err)
		}

		packet := New_FRecvPacket_VerifyCode()

		if code, ok := em.VerifyCodes[data.Id]; ok {
			if code == data.VerifyCode {
				packet.Status = true
			} else {
				packet.Status = false
			}
		} else {
			packet.Status = false
		}

		e, err := json.Marshal(packet)
		if err != nil {
			log.Fatal("Parse Error")
			return
		}

		if c, ok := em.Conns[data.Id]; ok {
			NetManagerInst().SendString(c, string(e))
		}
	}
	// NetManagerInst().Callbacks["FSendPacket_ResetPassword"] = func(v interface{}) {
	// 	type ResetPassword struct {
	// 		Id          string
	// 		NewPassword string
	// 	}

	// 	data := &ResetPassword{}
	// 	if err := mapstructure.Decode(v, data); err != nil {
	// 		fmt.Println(err)
	// 	}

	// 	packet := New_FRecvPacket_ResetPassword()

	// 	if DBManagerInst().ResetPassword(data.Id, data.NewPassword) {
	// 		packet.Status = true
	// 	} else {
	// 		packet.Status = false
	// 	}

	// 	e, err := json.Marshal(packet)
	// 	if err != nil {
	// 		log.Fatal("Parse Error")
	// 		return
	// 	}

	// 	if c, ok := em.Conns[data.Id]; ok {
	// 		NetManagerInst().SendString(c, string(e))

	// 		delete(em.Conns, data.Id)
	// 		delete(em.VerifyCodes, data.Id)
	// 	}
	// }
}

func (em *EmailManager) SendEmail(Destemail string) {

	code := em.VerifyCodes[Destemail]

	from := "hmc.troubleshoot@gmail.com"
	password := "hmc2021!"
	to := []string{
		Destemail,
	}

	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	subject := "Subject: Verify Code From HMC2021\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := "<h2>We send to you the Verify Code of the Hyundai Metaverse Convention 2021.</h2><h3>In the Hyundai Metaverse Convention 2021 log-in page, you can set your new password after entering Verify Code.</h3><p>Please check below for your Verify Code.</p><p><strong>Your Verify Code is : " + code + "</strong></p><p>Have a great time at Hyundai Metaverse Convention 2021.</p>"

	message := []byte(subject + mime + body)

	auth := smtp.PlainAuth("", from, password, smtpHost)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func GenerateVerifyCode() string {
	rand.Seed(time.Now().UnixNano())

	// Generate Random 6-digit Verification Code within 100000 - 999999 Range
	verification := rand.Intn((999999 - 100000 + 1) + 100000)

	return strconv.Itoa(verification)
}
