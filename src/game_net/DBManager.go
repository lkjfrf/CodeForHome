package game_net

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"

	"github.com/mitchellh/mapstructure"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"hmc_server.com/hmc_server/src/database"
	gorms "hmc_server.com/hmc_server/src/gorm"
	"hmc_server.com/hmc_server/src/helper"

	_ "github.com/go-sql-driver/mysql"
	//"github.com/jinzhu/gorm"
)

type DBManager struct {
	CarNumLock *sync.Mutex
	MTHLock    *sync.Mutex
	//MySQLDB       *sql.DB
	GORM        *gorm.DB
	Callbacks   map[string]func(interface{})
	playerinfo  gorms.PlayerInfo
	costumeinfo gorms.CostumeInfo
	inventory   gorms.Inventory
	top         gorms.Top
	bottom      gorms.Bottom
	shoes       gorms.Shoes
	acce        gorms.Acce

	wrcrank       gorms.WrcRank
	mth           gorms.MTH
	minigamecount gorms.MinigameCount

	community gorms.Community
	friend    gorms.Friend
}

var instance_DB *DBManager
var once_DB sync.Once
var IdMapMutex sync.RWMutex
var Idmap map[string][]string

func DBManagerInst() *DBManager {
	once_DB.Do(func() {
		instance_DB = &DBManager{}
	})
	return instance_DB
}

// func (db *DBManager) GetDynamo() *dynamodb.DynamoDB {
// 	sess := session.Must(session.NewSessionWithOptions(session.Options{
// 		SharedConfigState: session.SharedConfigEnable,
// 		Config:            *aws.NewConfig().WithDisableSSL(true),
// 	}))

// 	return dynamodb.New(sess)
// }

func (db *DBManager) Init() {

	fmt.Println("INIT_DBManager")
	// https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/using-dynamodb-with-go-sdk.html
	// 환경변수 등록이 필요하다.
	// window 기준
	// set AWS_ACCESS_KEY_ID=AKIAQJMFNQIKI256DVFB
	// set AWS_SECRET_ACCESS_KEY=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
	// set AWS_DEFAULT_REGION=ap-northeast-2
	// config파일로 불러오려고 했으나 잘 되지 않아서 env로 불러옴

	db.InitGORMConnection()
	db.GormInit()
	db.Test()

	db.CarNumLock = &sync.Mutex{}
	db.MTHLock = &sync.Mutex{}

	// dbs, err := sql.Open("mysql", "root:0000@tcp(127.0.0.1:3306)/hmc")
	// if err != nil {
	// 	panic(err)
	// 	log.Printf("SQL_Open Fail : %s", err)
	// } else {
	// 	db.MySQLDB = dbs
	// }

	NetManagerInst().Callbacks["FSendPacket_DBSignup"] = func(v interface{}) {
		data := &gorms.PlayerInfo{}

		if err := mapstructure.Decode(v, data); err != nil {
			fmt.Println(err)
		}

		db.NewPlayerSignup(data, data.Conn)
	}
	NetManagerInst().Callbacks["FSendPacket_DBSignin"] = func(v interface{}) {
		type signin struct {
			Conn     *net.TCPConn
			Playerid string
			Password string
		}

		data := &signin{}
		if err := mapstructure.Decode(v, data); err != nil {
			fmt.Println(err)
		}

		//sqlConn.QueryRow("select PassWord from playerinfo where Id = ?", data.Playerid).Scan(&Password)
		packet := New_FRecvPacket_DBSignin()

		if err := db.GORM.Table("player_info").Select("password").Where("player_id = ?", data.Playerid).Scan(&db.playerinfo.Password).Error; err != nil {
			fmt.Println(err)
			packet.Status = false
		}

		if db.playerinfo.Password == data.Password {
			packet.Status = true
		} else {
			packet.Status = false
		}

		e, err := json.Marshal(packet)
		if err != nil {
			log.Fatal("Parse Error")
			return
		}

		if !packet.Status {
			NetManagerInst().SendString(data.Conn, string(e))
			return
		}

		// 기존 중복체크 제거 ( 로직이 변경됨에 따라 스폰은 1번만 됨 )
		go db.SetupPlayer_sql(data.Playerid, data.Password, data.Conn)
	}

	NetManagerInst().Callbacks["FSendPacket_SetCostumeInfo"] = func(v interface{}) {
		data := &gorms.CostumeInfo{}
		helper.FillStruct_Interface(v, data)
		db.NewCostumeInfo(data)
	}

	NetManagerInst().Callbacks["FSendPacket_SyncInventory"] = func(v interface{}) {
		data := &database.Inventory{}
		helper.FillStruct_Interface(v, data)

		db.SyncInventory(data)
	}

	NetManagerInst().Callbacks["FSendPacket_UpdateCoin"] = func(v interface{}) {
		type CoinInfo struct {
			Id         string
			HCoin      int32
			TotalHCoin int32
		}

		data := &CoinInfo{}
		if err := mapstructure.Decode(v, data); err != nil {
			fmt.Println(err)
		}

		db.UpdateUserCoin(data.Id, data.HCoin, data.TotalHCoin)
	}

	NetManagerInst().Callbacks["FSendPacket_IntroduceMessage"] = func(v interface{}) {
		type Message struct {
			Id      string
			Message string
		}

		data := &Message{}
		if err := mapstructure.Decode(v, data); err != nil {
			fmt.Println(err)
		}

		infodata := db.playerinfo
		db.GORM.Table("player_info").Where("player_id = ?", data.Id).Scan(&infodata)
		msg := gorms.Community{Id: data.Id, Name: infodata.FirstName + " " + infodata.LastName, Team: infodata.Team, Message: data.Message}
		db.GORM.Create(&msg)
	}

	NetManagerInst().Callbacks["FSendPacket_FriendList"] = func(v interface{}) {
		type Firend struct {
			Id string
		}

		data := &Firend{}
		if err := mapstructure.Decode(v, data); err != nil {
			fmt.Println(err)
		}
		//packet := New_FRecvPacket_FriendList()

		// result := db.community
		// db.GORM.Table("community").Where("id = ?", data.Id).Scan(&result)
		// rows, _ := db.GORM.Model(&gorms.Community{}).Where("player_id = ?", data.Id).Rows()
		// for rows.Next() {
		// 	rows.Scan(&result)

		// 	packet.FriendList = append(packet.FriendList, FriendSearch{Name: result.FriendName, IsOnline: true})
		// }
		// rows.Close()

		// _e, _err := json.Marshal(packet)
		// if _err != nil {
		// 	log.Fatal("Parse Error")
		// 	return
		// }
		//NetManagerInst().SendString(c, string(_e))

	}

}

func (db *DBManager) GetMTHnumber() int32 {
	var counts int64
	//db.MySQLDB.QueryRow("select count(*) from mth").Scan(&counts)
	db.GORM.Model(&gorms.MTH{}).Select("index").Count(&counts)
	return int32(counts + 1)
}

func (db *DBManager) SaveHTM(Region string, Message string) {

	mth := &database.MTH{}

	mth.Region = Region
	mth.Message = Message

	// _, err := db.MySQLDB.Exec("insert into mth value (?, ?)", mth.Message, mth.Region)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	MthData := gorms.MTH{Message: mth.Message, Region: mth.Region}
	if err := db.GORM.Create(&MthData).Error; err != nil {
		fmt.Println(err)
	}
}

func (db *DBManager) GetHCoinRank(id string) int32 {

	var rankresult int32
	//db.MySQLDB.QueryRow("select ranking from ( select id, rank() over (order by totalcoin desc) as ranking from playerinfo) as t where t.id = ?", id).Scan(&rankresult)

	return rankresult
}

func (db *DBManager) SyncInventory(Items *database.Inventory) {
	// for i := 0; i < len(Items.ItemListTop); i++ {
	// 	inventory := &gorms.Inventory{}
	// 	inventory.PlayerId = Items.Id
	// 	inventory.Item = "Top"
	// 	inventory.ItemNum = int32(i)
	// 	inventory.IsStored = Items.ItemListTop[i]
	// 	db.GORM.Create(&inventory)
	// }
	// for i := 0; i < len(Items.ItemListBottom); i++ {
	// 	inventory := &gorms.Inventory{}
	// 	inventory.PlayerId = Items.Id
	// 	inventory.Item = "Bottom"
	// 	inventory.ItemNum = int32(i)
	// 	inventory.IsStored = Items.ItemListTop[i]
	// 	db.GORM.Create(&inventory)
	// }
	// for i := 0; i < len(Items.ItemListShoes); i++ {
	// 	inventory := &gorms.Inventory{}
	// 	inventory.PlayerId = Items.Id
	// 	inventory.Item = "Shoes"
	// 	inventory.ItemNum = int32(i)
	// 	inventory.IsStored = Items.ItemListTop[i]
	// 	db.GORM.Create(&inventory)
	// }
	// for i := 0; i < len(Items.ItemListAcce); i++ {
	// 	inventory := &gorms.Inventory{}
	// 	inventory.PlayerId = Items.Id
	// 	inventory.Item = "Acce"
	// 	inventory.ItemNum = int32(i)
	// 	inventory.IsStored = Items.ItemListTop[i]
	// 	db.GORM.Create(&inventory)
	// }

	inven := gorms.Inventory{PlayerId: Items.Id}
	//db.GORM.Where("player_id = ?", Items.Id).Find(&db.inventory)
	r := db.GORM.Where("player_id = ?", Items.Id).Limit(1).Find(&db.inventory)
	if r.Error != nil {
		fmt.Println(r.Error)
	}
	exists := r.RowsAffected > 0

	// 상위삭제시 하위테이블 모두 삭제되도록 수정예정 (지금은 쓰레기값 남음)
	if exists {
		db.GORM.Where("id = ?", Items.Id).Delete(&db.inventory.TopInven)
		db.GORM.Where("id = ?", Items.Id).Delete(&db.inventory.BottomInven)
		db.GORM.Where("id = ?", Items.Id).Delete(&db.inventory.ShoesInven)
		db.GORM.Where("id = ?", Items.Id).Delete(&db.inventory.AcceInven)
		db.GORM.Where("player_id = ?", Items.Id).Delete(&db.inventory)
	}

	var TopArr []gorms.Top
	var BottomArr []gorms.Bottom
	var ShoesArr []gorms.Shoes
	var AcceArr []gorms.Acce

	for i := 0; i < len(Items.ItemListTop); i++ {
		TopArr = append(TopArr, gorms.Top{ItemNum: int32(i), IsStored: Items.ItemListTop[i]})
	}
	for i := 0; i < len(Items.ItemListBottom); i++ {
		BottomArr = append(BottomArr, gorms.Bottom{ItemNum: int32(i), IsStored: Items.ItemListBottom[i]})
	}
	for i := 0; i < len(Items.ItemListShoes); i++ {
		ShoesArr = append(ShoesArr, gorms.Shoes{ItemNum: int32(i), IsStored: Items.ItemListShoes[i]})
	}
	for i := 0; i < len(Items.ItemListAcce); i++ {
		AcceArr = append(AcceArr, gorms.Acce{ItemNum: int32(i), IsStored: Items.ItemListAcce[i]})
	}
	inven = gorms.Inventory{PlayerId: Items.Id, TopInven: TopArr, BottomInven: BottomArr, ShoesInven: ShoesArr, AcceInven: AcceArr}
	db.GORM.Create(&inven)

}

func (db *DBManager) SyncMinigameCount(Items *database.MinigameCount) {

	// SQLConn := db.MySQLDB
	// var exist bool

	// err := SQLConn.QueryRow("select exists( select id from minigamecount where id = ?) as result", Items.Id).Scan(&exist)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// if exist {
	// 	//	_, err = SQLConn.Exec("delete from minigamecount where id = ?", playerId)
	// 	_, err = db.MySQLDB.Exec("delete from minigamecount where id = ?", Items.Id)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// }

	// _, err = db.MySQLDB.Exec("insert into minigamecount value (?, ?, ?, ?, ?, ?, ?, ?)", Items.Id, Items.Elevate_Count, Items.G80ev_Count, Items.Ioniq_Count, Items.Wrc_Count, Items.Minigame_Count, Items.LastLoginDay, Items.Region)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	if err := db.GORM.Unscoped().Where("Id = ?", Items.Id).Delete(&db.minigamecount).Error; err != nil {
		fmt.Println(err)
	}

	MinigameData := gorms.MinigameCount{Id: Items.Id, LastLoginDay: Items.LastLoginDay, ElevateCount: Items.Elevate_Count, G80Count: Items.G80ev_Count, IoniqCount: Items.Ioniq_Count, WrcCount: Items.Wrc_Count, MiniGameCount: Items.Minigame_Count, Region: Items.Region}
	if err := db.GORM.Create(&MinigameData).Error; err != nil {
		fmt.Println(err)
	}

}

func (db *DBManager) SetInventory(PlayerId string) database.Inventory {

	// item := database.Inventory{}

	// SQLConn := db.MySQLDB

	// var count int
	// var result bool
	// //Top
	// err := SQLConn.QueryRow("select count(isstored) from inventory where id = ? and item = ?", PlayerId, "Top").Scan(&count)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// item.ItemListTop = make([]bool, count)

	// for i := 0; i < count; i++ {
	// 	err = SQLConn.QueryRow("select isstored from inventory where id = ? and item = ? and itemnum = ?", PlayerId, "Top", i).Scan(&result)
	// 	item.ItemListTop[i] = result
	// }
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// //Bottom
	// err = SQLConn.QueryRow("select count(isstored) from inventory where id = ? and item = ?", PlayerId, "Bottom").Scan(&count)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// item.ItemListBottom = make([]bool, count)

	// for i := 0; i < count; i++ {
	// 	err = SQLConn.QueryRow("select isstored from inventory where id = ? and item = ? and itemnum = ?", PlayerId, "Bottom", i).Scan(&result)
	// 	item.ItemListBottom[i] = result
	// }
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// //Shoes
	// err = SQLConn.QueryRow("select count(isstored) from inventory where id = ? and item = ?", PlayerId, "Shoes").Scan(&count)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// item.ItemListShoes = make([]bool, count)

	// for i := 0; i < count; i++ {
	// 	err = SQLConn.QueryRow("select isstored from inventory where id = ? and item = ? and itemnum = ?", PlayerId, "Shoes", i).Scan(&result)
	// 	item.ItemListShoes[i] = result
	// }
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// //Accessory
	// err = SQLConn.QueryRow("select count(isstored) from inventory where id = ? and item = ?", PlayerId, "Acce").Scan(&count)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// item.ItemListAcce = make([]bool, count)

	// for i := 0; i < count; i++ {
	// 	err = SQLConn.QueryRow("select isstored from inventory where id = ? and item = ? and itemnum = ?", PlayerId, "Acce", i).Scan(&result)
	// 	item.ItemListAcce[i] = result
	// }
	// if err != nil {
	// 	fmt.Println(err)
	// }

	item := database.Inventory{}
	//var result bool
	//var count int64

	// 함수화로 정리 예정
	inven := gorms.Inventory{PlayerId: PlayerId}
	db.GORM.Where("id = ?", PlayerId).Find(&inven.TopInven)
	item.ItemListTop = make([]bool, len(inven.TopInven))
	db.GORM.Where("id = ?", PlayerId).Find(&inven.BottomInven)
	item.ItemListBottom = make([]bool, len(inven.BottomInven))
	db.GORM.Where("id = ?", PlayerId).Find(&inven.ShoesInven)
	item.ItemListShoes = make([]bool, len(inven.ShoesInven))
	db.GORM.Where("id = ?", PlayerId).Find(&inven.AcceInven)
	item.ItemListAcce = make([]bool, len(inven.AcceInven))

	for i, t := range inven.TopInven {
		item.ItemListTop[i] = t.IsStored
	}
	for i, t := range inven.BottomInven {
		item.ItemListBottom[i] = t.IsStored
	}
	for i, t := range inven.ShoesInven {
		item.ItemListShoes[i] = t.IsStored
	}
	for i, t := range inven.AcceInven {
		item.ItemListAcce[i] = t.IsStored
	}

	// db.GORM.Model(&gorms.Inventory{}).Where("player_id = ?", PlayerId).Select("TopInven").Count(&count)
	// rows, err := db.GORM.Model(&gorms.Inventory{}).Where("player_id = ? AND item = ?", PlayerId, "Top").Select("is_stored").Rows()
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// i := 0
	// item.ItemListTop = make([]bool, count)
	// for rows.Next() {
	// 	rows.Scan(&result)
	// 	item.ItemListTop[i] = result
	// 	i++
	// }
	// rows.Close()

	// db.GORM.Model(&gorms.Inventory{}).Where("player_id = ? AND item = ?", PlayerId, "Bottom").Select("is_stored").Count(&count)
	// rows, err = db.GORM.Model(&gorms.Inventory{}).Where("player_id = ? AND item = ?", PlayerId, "Bottom").Select("is_stored").Rows()
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// i = 0
	// item.ItemListBottom = make([]bool, count)
	// for rows.Next() {
	// 	rows.Scan(&result)
	// 	item.ItemListBottom[i] = result
	// 	i++
	// }
	// rows.Close()

	// db.GORM.Model(&gorms.Inventory{}).Where("player_id = ? AND item = ?", PlayerId, "Shoes").Select("is_stored").Count(&count)
	// rows, err = db.GORM.Model(&gorms.Inventory{}).Where("player_id = ? AND item = ?", PlayerId, "Shoes").Select("is_stored").Rows()
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// defer rows.Close()
	// i = 0
	// item.ItemListShoes = make([]bool, count)
	// for rows.Next() {
	// 	rows.Scan(&result)
	// 	item.ItemListShoes[i] = result
	// 	i++
	// }
	// rows.Close()

	// db.GORM.Model(&gorms.Inventory{}).Where("player_id = ? AND item = ?", PlayerId, "Acce").Select("is_stored").Count(&count)
	// rows, err = db.GORM.Model(&gorms.Inventory{}).Where("player_id = ? AND item = ?", PlayerId, "Acce").Select("is_stored").Rows()
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// i = 0
	// item.ItemListAcce = make([]bool, count)
	// for rows.Next() {
	// 	rows.Scan(&result)
	// 	item.ItemListAcce[i] = result
	// 	i++
	// }
	// rows.Close()

	return item
}

func (db *DBManager) UpdateCostume(PlayerId string, ClothIndex int, ClothType int) {
	// if ClothType == 0 {
	// 	_, err := db.MySQLDB.Exec("update costumeinfo set topindex = ? where id = ?", ClothIndex, PlayerId)
	// 	if err != nil {
	// 		log.Println(err)
	// 	}
	// } else if ClothType == 1 {
	// 	_, err := db.MySQLDB.Exec("update costumeinfo set bottomindex = ? where id = ?", ClothIndex, PlayerId)
	// 	if err != nil {
	// 		log.Println(err)
	// 	}
	// } else if ClothType == 2 {
	// 	_, err := db.MySQLDB.Exec("update costumeinfo set shoesindex = ? where id = ?", ClothIndex, PlayerId)
	// 	if err != nil {
	// 		log.Println(err)
	// 	}
	// } else if ClothType == 3 {
	// 	_, err := db.MySQLDB.Exec("update costumeinfo set accessoryindex = ? where id = ?", ClothIndex, PlayerId)
	// 	if err != nil {
	// 		log.Println(err)
	// 	}
	// }

	if ClothType == 0 {
		if err := db.GORM.Model(&gorms.CostumeInfo{}).Where("id = ?", PlayerId).Update("top_index", int32(ClothIndex)).Error; err != nil {
			fmt.Println(err)
		}
	} else if ClothType == 1 {
		if err := db.GORM.Model(&gorms.CostumeInfo{}).Where("id = ?", PlayerId).Update("bottom_index", int32(ClothIndex)).Error; err != nil {
			fmt.Println(err)
		}
	} else if ClothType == 2 {
		if err := db.GORM.Model(&gorms.CostumeInfo{}).Where("id = ?", PlayerId).Update("shoes_index", int32(ClothIndex)).Error; err != nil {
			fmt.Println(err)
		}
	} else if ClothType == 3 {
		if err := db.GORM.Model(&gorms.CostumeInfo{}).Where("id = ?", PlayerId).Update("accessory_index", int32(ClothIndex)).Error; err != nil {
			fmt.Println(err)
		}
	}
}

func (db *DBManager) UpdateUserCoin(PlayerId string, PlayerCoin int32, TotalHCoin int32) {

	// _, err := db.MySQLDB.Exec("update playerinfo set hcoin = ?, totalcoin = ? where id = ?", PlayerCoin, TotalHCoin, PlayerId)
	// if err != nil {
	// 	log.Println(err)
	// }

	if err := db.GORM.Model(&gorms.PlayerInfo{}).Where("player_id = ?", PlayerId).Updates(gorms.PlayerInfo{TotalHCoin: TotalHCoin, HCoin: PlayerCoin}).Error; err != nil {
		fmt.Println(err)
	}
}

func (db *DBManager) GetCostume(PlayerId string, Cdata *FRecvPacket_DBSignin) bool {
	// SQLConn := db.MySQLDB
	// err := SQLConn.QueryRow("select exists (select id from costumeinfo where id = ?)", PlayerId).Scan(&exists)

	// if err != nil {
	// 	fmt.Println(err)
	// }

	// if exists {
	// 	SQLConn.QueryRow("select topindex from costumeinfo where id = ?", PlayerId).Scan(&Cdata.TopIndex)
	// 	SQLConn.QueryRow("select bottomindex from costumeinfo where id = ?", PlayerId).Scan(&Cdata.BottomIndex)
	// 	SQLConn.QueryRow("select shoesindex from costumeinfo where id = ?", PlayerId).Scan(&Cdata.ShoesIndex)
	// 	SQLConn.QueryRow("select hairindex from costumeinfo where id = ?", PlayerId).Scan(&Cdata.HairIndex)
	// 	SQLConn.QueryRow("select haircolorindex from costumeinfo where id = ?", PlayerId).Scan(&Cdata.HairColorIndex)
	// 	SQLConn.QueryRow("select faceindex from costumeinfo where id = ?", PlayerId).Scan(&Cdata.FaceIndex)
	// 	SQLConn.QueryRow("select accessoryindex from costumeinfo where id = ?", PlayerId).Scan(&Cdata.AccessoryIndex)
	// }awd

	r := db.GORM.Where("id = ?", PlayerId).Limit(1).Find(&gorms.CostumeInfo{})
	if r.Error != nil {
		fmt.Println(r.Error)
	}
	exists := r.RowsAffected > 0

	TableData := gorms.CostumeInfo{Id: PlayerId}
	if exists {
		exists = true
		db.GORM.Table("costume_info").Where("id = ?", PlayerId).Find(&TableData)
		Cdata.TopIndex = TableData.TopIndex
		Cdata.BottomIndex = TableData.BottomIndex
		Cdata.ShoesIndex = TableData.ShoesIndex
		Cdata.AccessoryIndex = TableData.AccessoryIndex
		Cdata.HairIndex = TableData.HairIndex
		Cdata.HairColorIndex = TableData.HairColorIndex
		Cdata.FaceIndex = TableData.FaceIndex
	} else {
		exists = false
	}

	return exists
}

func (db *DBManager) NewCostumeInfo(Cdata *gorms.CostumeInfo) {
	// _, err := db.MySQLDB.Exec("insert into costumeinfo(id, skinindex, topindex, bottomindex, shoesindex, hairindex, haircolorindex, faceindex, accessoryindex) value (?, ?, ?, ?, ?, ?, ?, ?, ?)", Cdata.Id, Cdata.SkinIndex, Cdata.TopIndex, Cdata.BottomIndex, Cdata.ShoesIndex, Cdata.HairIndex, Cdata.HairColorIndex, Cdata.FaceIndex, Cdata.AccessoryIndex)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	count := int64(0)
	db.GORM.Table("costume_info").Where("id = ?", Cdata.Id).Count(&count)

	if count > 0 {
		if err := db.GORM.Unscoped().Where("Id = ?", Cdata.Id).Delete(&db.costumeinfo).Error; err != nil {
			fmt.Println(err)
		}
	}
	if err := db.GORM.Create(&Cdata).Error; err != nil {
		fmt.Println(err)
	}

}

func (db *DBManager) GetCarNum(PlayerId string) string {
	var CarNumInt int
	// err := db.MySQLDB.QueryRow("select carnum from playerinfo where id = ?", PlayerId).Scan(&CarNumInt)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	if err := db.GORM.Table("player_info").Select("car_num").Where("player_id = ?", PlayerId).Scan(&CarNumInt).Error; err != nil {
		fmt.Println(err)
	}

	CarNumStr := strconv.Itoa(CarNumInt)
	var carNum string

	if len(CarNumStr) == 1 {
		carNum = "HMC " + "0000" + CarNumStr
	} else if len(CarNumStr) == 2 {
		carNum = "HMC " + "000" + CarNumStr
	} else if len(CarNumStr) == 3 {
		carNum = "HMC " + "00" + CarNumStr
	} else if len(CarNumStr) == 4 {
		carNum = "HMC " + "0" + CarNumStr
	} else {
		carNum = "HMC " + CarNumStr
	}

	return carNum
}

func (db *DBManager) SetupPlayer_sql(playerid string, password string, c *net.TCPConn) {

	packet := New_FRecvPacket_DBSignin()
	packet.Status = true

	// infodata := db.playerinfo
	// if err := db.GORM.Table("player_info").Where("player_id = ?", playerid).First(&infodata).Error; err != nil {
	// 	fmt.Println(err)
	// }

	infodata := db.playerinfo
	if err := db.GORM.Table("player_info").Where("player_id = ?", playerid).Scan(&infodata).Error; err != nil {
		fmt.Println(err)
	}

	// playerinfo := &database.PlayerInfo{}
	// rows, err := sqlConn.Query("select country, firstname, lastname, birthday, birthmonth, dealertype, isman, region, timezone, iscostumed, hcoin, totalcoin  from playerinfo where id = ?", playerid)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer rows.Close()

	// for rows.Next() {
	// 	if err := rows.Scan(&playerinfo.Country, &playerinfo.FirstName, &playerinfo.LastName, &playerinfo.BirthDay, &playerinfo.BirthMonth, &playerinfo.DealerType, &playerinfo.IsMan, &playerinfo.Region, &playerinfo.Timezone, &playerinfo.IsCostumed, &playerinfo.HCoin, &playerinfo.TotalHCoin); err != nil {
	// 		fmt.Println(err)
	// 	}
	// }

	packet.CarNum = db.GetCarNum(playerid)
	packet.IsCostumed = db.GetCostume(playerid, &packet)
	packet.FirstName = infodata.FirstName
	packet.LastName = infodata.LastName
	packet.HCoin = infodata.HCoin
	packet.Team = infodata.Team
	packet.IsMan = infodata.IsMan
	packet.TotalHCoin = infodata.TotalHCoin

	MiniGame := db.SetMinigame(playerid)

	packet.LastLoginDay = MiniGame.LastLoginDay
	packet.Wrc_Count = MiniGame.WrcCount
	packet.G80ev_Count = MiniGame.G80Count
	packet.Ioniq_Count = MiniGame.IoniqCount
	packet.Elevate_Count = MiniGame.ElevateCount
	packet.Minigame_Count = MiniGame.MiniGameCount

	item := db.SetInventory(playerid)

	packet.ItemListTop = item.ItemListTop
	packet.ItemListBottom = item.ItemListBottom
	packet.ItemListShoes = item.ItemListShoes
	packet.ItemListAcce = item.ItemListAcce

	// questitem := db.GetQuest(playerid

	// packet.Quest_Overview = questitem.Quest_Overview
	// packet.Quest_Brand = questitem.Quest_Brand
	// packet.Quest_Product = questitem.Quest_Product
	// packet.Quest_LiveStation = questitem.Quest_LiveStation
	// packet.Quest_RHQ = questitem.Quest_RHQ
	// packet.Quest_Fuel = questitem.Quest_Fuel

	e, err := json.Marshal(packet)
	if err != nil {
		log.Fatal("Parse Error")
		return
	}

	NetManagerInst().SendString(c, string(e))
}

func (db *DBManager) NewPlayerSignup(playerinfo *gorms.PlayerInfo, c *net.TCPConn) {

	//sqlConn := db.MySQLDB
	//sqlConn.QueryRow("select MAX(carnum) from playerinfo").Scan(&CarNums)
	var CarNums int32

	db.GORM.Last(&db.playerinfo).Order("CarNum")
	CarNums = db.playerinfo.CarNum
	CarNums++
	playerinfo.CarNum = CarNums

	//infodata := gorms.PlayerInfo{Id: playerinfo.PlayerId, Password: playerinfo.Password, FirstName: playerinfo.FirstName, LastName: playerinfo.LastName, BirthMonth: playerinfo.BirthMonth, BirthDay: playerinfo.BirthDay, IsMan: playerinfo.IsMan, Country: playerinfo.Country, Region: playerinfo.Region, DealerType: playerinfo.DealerType, Timezone: playerinfo.Timezone, IsCostumed: playerinfo.IsCostumed, HCoin: playerinfo.HCoin, TotalHCoin: playerinfo.TotalHCoin, CarNum: CarNums}
	packet := New_FRecvPacket_DBSignup()
	if err := db.GORM.Create(&playerinfo).Error; err != nil {
		fmt.Println(err)
		packet.Status = false
	} else {
		packet.Status = true
	}
	//_, err := sqlConn.Exec("insert into playerinfo(id, country, firstname, lastname, birthday, birthmonth, dealertype, isman, password, carnum, timezone, iscostumed, region) value (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", playerinfo.PlayerId, playerinfo.Country, playerinfo.FirstName, playerinfo.LastName, playerinfo.BirthDay, playerinfo.BirthMonth, playerinfo.DealerType, playerinfo.IsMan, playerinfo.Password, CarNums, playerinfo.Timezone, false, playerinfo.Region)

	_e, _err := json.Marshal(packet)
	if _err != nil {
		log.Fatal("Parse Error")
		return
	}
	NetManagerInst().SendString(c, string(_e))

}

func (db *DBManager) AddWRCRank(rankinfo *database.WRCRank) {

	// _, err := db.MySQLDB.Exec("insert into wrcrank value (?, ?)", rankinfo.UserName, rankinfo.Laptime)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	//db.GORM.
}

func (db *DBManager) BroadCastWRCRank() FRecvPacket_WRCRankUpdate {

	packet := New_FRecvPacket_WRCRankUpdate()
	// SQLConn := db.MySQLDB
	// //var resultLen int
	// //SQLConn.QueryRow("select count(*) from wrcrank").Scan(&resultLen)

	// rows, err := SQLConn.Query("select id, laptime from wrcrank order by laptime asc")
	// if err != nil {
	// 	log.Println(err)
	// }

	// for rows.Next() {
	// 	rankinfo := WRCRankInfo{}

	// 	err := rows.Scan(&rankinfo.UserName, &rankinfo.Laptime)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// 	packet.Ranking = append(packet.Ranking, rankinfo)
	// }

	// rows.Close()
	return packet
}

func (db *DBManager) SetMinigame(playerId string) *gorms.MinigameCount {
	minigame := &gorms.MinigameCount{}

	// SQLConn := db.MySQLDB
	// rows, err := SQLConn.Query("select id, elevatecount, g80count, ioniqcount, wrccount, minigamecount, lastloginday, region from minigamecount where id = ?", playerId)

	// for rows.Next() {
	// 	if err := rows.Scan(&minigame.Id, &minigame.Elevate_Count, &minigame.G80ev_Count, &minigame.Ioniq_Count, &minigame.Wrc_Count, &minigame.Minigame_Count, &minigame.LastLoginDay, &minigame.Region); err != nil {
	// 		fmt.Println(err)
	// 	}
	// }

	// if err != nil {
	// 	fmt.Println(err)
	// }

	if err := db.GORM.Table("minigame_count").Where("Id = ?", playerId).Scan(&minigame).Error; err != nil {
		fmt.Println(err)
	}

	return minigame
}

func (db *DBManager) GormInit() {

	//db.GORM.AutoMigrate(&gorms.PlayerInfo{}, &gorms.Top{})
	//, &gorms.Top{}, &gorms.Bottom{}, &gorms.Shoes{}, &gorms.Acce{})
	// model := gorms.PlayerInfo{}
	// db.GORM.Model(&model).Relat

	db.GORM.AutoMigrate(&gorms.PlayerInfo{})
	db.GORM.AutoMigrate(&gorms.Inventory{}, &gorms.Top{}, &gorms.Bottom{}, &gorms.Shoes{}, &gorms.Acce{})

	//db.GORM.AutoMigrate(&gorms.User{})

	//db.GORM.Preload("Orders").Find()

	db.GORM.AutoMigrate(&gorms.CostumeInfo{})
	db.GORM.AutoMigrate(&gorms.WrcRank{})
	db.GORM.AutoMigrate(&gorms.MTH{})
	db.GORM.AutoMigrate(&gorms.MinigameCount{})
	db.GORM.AutoMigrate(&gorms.Friend{})
	db.GORM.AutoMigrate(&gorms.Community{})

	//p := &gorms.PlayerInfo{PlayerId: "e1@e1"}
	//p.Tops = append(p.Tops, gorms.Top{ItemNum: 3})
	//db.GORM.Create(p)

	//var found []*gorms.PlayerInfo
	//db.GORM.Unscoped().Preload("InvenTop").Scan(&found)

	// rows, err := db.GORM.Table("player_info").Where("player_info.player_id = ?", "q1@q1").
	// 	Joins("Join top on top.player_id = player_info.player_id").Select("player_info.player_id, top.item_num, top.is_stored").Rows()

	// if err != nil {
	// 	fmt.Println(err)
	// }

	// defer rows.Close()

	// newplayerinfo := gorms.PlayerInfo{}
	// newplayerinfo.InvenTop = make([]gorms.Top, 0)
	// for rows.Next() {
	// 	newTop := gorms.Top{}
	// 	err = rows.Scan(&newplayerinfo.PlayerId, &newTop.ItemNum, &newTop.IsStored)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// 	newplayerinfo.InvenTop = append(newplayerinfo.InvenTop, newTop)
	// }

	// var TopArr []gorms.Top
	// TopArr = append(TopArr, gorms.Top{ItemNum: 1, IsStored: false})
	// TopArr = append(TopArr, gorms.Top{ItemNum: 2, IsStored: true})

	// inven := gorms.Inventory{PlayerId: "q1@q1", TopInven: TopArr}
	// //inven2 := gorms.Inventory{PlayerId: "q1@q1", TopInven: []gorms.Top{{PlayerId: "q1@q1"}}}
	// //top := gorms.Top{}
	// //db.GORM.Where("player_id = ?", "q1@q1").Find(gorms.Inventory.TopInven)
	// //	db.GORM.Find(&inven.TopInven)
	// //	db.GORM.Find(gorms.Inventory{PlayerId: "q1@q1"}.TopInven)
	// db.GORM.Create(&inven)

	// db.GORM.Create(&inven.TopInven)

	// //db.GORM.Model(&gorms.Inventory{}).Where("player_id = ?", "q1@q1").Update("TopInven", []gorms.Top{{PlayerId: "real a1", ItemNum: 3, IsStored: true}})
	// //db.GORM.Unscoped().Where("player_id = ?", "a1@a1").Delete(&inven.TopInven)

	// db.GORM.Select(clause.Associations).Where("id = ?", "q1@q1").Delete(&db.inventory.TopInven)
	// db.GORM.Select(clause.Associations).Where("id = ?", "q2@q2").Delete(&db.top)

	// db.GORM.Where("player_id = ?", "q1@q1").Delete(&db.inventory)
}

func (db *DBManager) InitGORMConnection() {
	//dsn := "root:1q2w3e4r@tcp(database-2.cgd4b0uzz35l.ap-northeast-2.rds.amazonaws.com:3306)/hmc?charset=utf8mb4&parseTime=True&loc=Local"
	dsn := "root:0000@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"

	// dbss, err := gorm.Open("mysql", dsn)
	// if err != nil {
	// 	panic("failed to connect database")
	// }

	dbss, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
			//NameReplacer:  strings.NewReplacer("PlayerInfo", "playerinfo"),
		},
	})
	if err != nil {
		log.Println(err)
	} else {
		db.GORM = dbss
	}

}

func (db *DBManager) FriendRequest(playerId string) string {

	if err := db.GORM.Table("community").Select("name").Where("id = ?", playerId).Scan(&db.community.Name).Error; err != nil {
		fmt.Println(err)
	}
	return db.community.Name
}

func (db *DBManager) SearchList(searchInput string, Players sync.Map) []FriendSearch {

	result := db.community
	rows, _ := db.GORM.Model(&gorms.Community{}).Where("id like ?", "%"+searchInput+"%").Select("id", "name").Rows()
	packet := New_FRecvPacket_FriendList()

	for rows.Next() {
		rows.Scan(&result.Id, &result.Name)
		_, ok := Players.Load(result.Id)

		packet.FriendList = append(packet.FriendList, FriendSearch{Name: result.Name, FriendId: result.Id, IsOnline: ok})
	}
	rows.Close()

	return packet.FriendList
}

func (db *DBManager) SearchFriendList(playerId string, Players sync.Map) []FriendSearch {

	result := db.community
	db.GORM.Table("friend").Where("id = ?", playerId).Scan(&result.FriendList)

	packet := New_FRecvPacket_FriendList()
	for i := 0; i < len(result.FriendList); i++ {
		_, ok := Players.Load(result.Id)
		packet.FriendList = append(packet.FriendList, FriendSearch{Name: result.FriendList[i].FriendName, FriendId: result.FriendList[i].FriendId, IsOnline: ok})
	}

	return packet.FriendList
}

func (db *DBManager) SetFriend(playerId string, targetId string, targetName string) {
	var friend gorms.Friend
	//friend = append(friend, gorms.Friend{Id: playerId, FriendId: targetId, FriendName: targetName})
	friend = gorms.Friend{Id: playerId, FriendId: targetId, FriendName: targetName}
	//db.friend = {id}
	db.GORM.Create(&friend)
}

func (db *DBManager) SetFriendInfo(playerId string, FriendId string) FRecvPacket_FriendInfo {
	result := FRecvPacket_FriendInfo{}
	community := db.community
	db.GORM.Table("community").Where("id = ?", FriendId).Scan(&community)
	result.FriendId = community.Id
	result.Message = community.Message
	result.Name = community.Name
	result.Team = community.Team
	result.PacketName = "FRecvPacket_FriendInfo"

	db.GORM.Table("friend").Where("id = ?", FriendId).Scan(&community.FriendList)

	result.FriendNum = int32(len(community.FriendList))

	return result
}

func (db *DBManager) Test() {

	// result := db.community
	// rows, err := db.GORM.Model(&gorms.Community{}).Where("id = ?", "q2@q2").Select("id", "name").Rows()
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// for rows.Next() {
	// 	rows.Scan(&result.Id, &result.Name)

	// }
	// rows.Close()

	// result := db.community
	// db.GORM.Table("friend").Where("id = ?", "q1@q1").Scan(&result.FriendList)
	// a := len(result.FriendList)
	// fmt.Println(a)

	// community := db.community
	// db.GORM.Table("community").Where("id = ?", "q1@q1").Scan(&community)
}
