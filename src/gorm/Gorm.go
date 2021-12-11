package gorm

import (
	"net"
)

type PlayerInfo struct {
	PlayerId   string `gorm:"primaryKey"`
	Password   string
	FirstName  string
	LastName   string
	IsMan      bool
	IsCostumed bool
	HCoin      int32
	TotalHCoin int32
	CarNum     int32
	Team       string
	Conn       *net.TCPConn `gorm:"type:text"`
}

type CostumeInfo struct {
	Id             string
	SkinIndex      int32
	TopIndex       int32
	BottomIndex    int32
	HairIndex      int32
	ShoesIndex     int32
	HairColorIndex int32
	FaceIndex      int32
	AccessoryIndex int32
}

type Inventory struct {
	PlayerId string `gorm:"primaryKey"`

	TopInven    []Top    `gorm:"foreignKey:Id"`
	BottomInven []Bottom `gorm:"foreignKey:Id"`
	ShoesInven  []Shoes  `gorm:"foreignKey:Id"`
	AcceInven   []Acce   `gorm:"foreignKey:Id"`
}

type WrcRank struct {
	Id      string
	IapTime int32
}

type MTH struct {
	Index   int32 `gorm:"primaryKey;autoIncrement:true"`
	Message string
	Region  string
}

type MinigameCount struct {
	Id            string
	ElevateCount  int32
	G80Count      int32
	IoniqCount    int32
	WrcCount      int32
	MiniGameCount int32
	LastLoginDay  int32
	Region        string
}

type Top struct {
	Id       string
	Index    int32 `gorm:"primaryKey;autoIncrement:true"`
	ItemNum  int32
	IsStored bool
}

type Bottom struct {
	Id       string
	Index    int32 `gorm:"primaryKey;autoIncrement:true"`
	ItemNum  int32
	IsStored bool
}

type Shoes struct {
	Id       string
	Index    int32 `gorm:"primaryKey;autoIncrement:true"`
	ItemNum  int32
	IsStored bool
}

type Acce struct {
	Id       string
	Index    int32 `gorm:"primaryKey;autoIncrement:true"`
	ItemNum  int32
	IsStored bool
}

type Community struct {
	Id         string
	Name       string
	Team       string
	Message    string
	FriendList []Friend `gorm:"foreignKey:Id"`
}

type Friend struct {
	Index      int32 `gorm:"primaryKey;autoIncrement:true"`
	Id         string
	FriendId   string
	FriendName string
}
