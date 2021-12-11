package main

import (
	"fmt"
	"runtime"

	_ "github.com/go-sql-driver/mysql"

	"hmc_server.com/hmc_server/src/chat"
	core "hmc_server.com/hmc_server/src/gameCore"
	"hmc_server.com/hmc_server/src/game_net"
	"hmc_server.com/hmc_server/src/objects"
)

func main() {

	fmt.Println("INIT_Main")
	// gorm.PlayerInfoCreate()

	runtime.GOMAXPROCS(runtime.NumCPU())

	core.GameManagerInst().Init()
	game_net.DBManagerInst().Init()
	objects.ObjectManagerInst().Init()
	chat.ChatManagerInst().Init()
	game_net.EmailManagerInst().Init()
	game_net.NetManagerInst().Init()

	fmt.Println("INIT_END_")
}
