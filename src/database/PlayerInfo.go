package database

import "net"

// type PlayerInfo struct {
// 	Conn       *net.TCPConn
// 	FirstName  string `dynamodbav:"firstname"`
// 	LastName   string `dynamodbav:"lastname"`
// 	BirthMonth int32  `dynamodbav:"birthmonth"`
// 	BirthDay   int32  `dynamodbav:"birthday"`
// 	IsMan      bool   `dynamodbav:"isman"`
// 	Country    string `dynamodbav:"country"`
// 	Region     string `dynamodbav:"region"`
// 	DealerType string `dynamodbav:"dealertype"`
// 	Timezone   int32  `dynamodbav:"timezone"`

// 	// Playerid = Email
// 	PlayerId string `dynamodbav:"playerid"`
// 	Password string `dynamodbav:"password"`

// 	// For Moderator
// 	IsModerator bool   `dynamodbav:"ismoderator"`
// 	Category    string `dynamodbav:"category"`

// 	// InGame
// 	HCoin              int32 `dynamodbav:"hcoin"`
// 	TotalHCoin         int32 `dynamodbav:"totalhcoin"`
// 	IsTutorial         bool  `dynamodbav:"istutorial"`
// 	IsTutorialRewarded bool  `dynamodbav:"istutorialrewarded"`
// }

type PlayerInfo struct {
	Conn       *net.TCPConn
	FirstName  string
	LastName   string
	BirthMonth int32
	BirthDay   int32
	IsMan      bool
	Country    string
	Region     string
	DealerType string
	Timezone   int32
	IsCostumed bool

	// Playerid = Email
	PlayerId string
	Password string

	// InGame
	HCoin      int32
	TotalHCoin int32
}
