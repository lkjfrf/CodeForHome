package core

import (
	"fmt"
	"sync"

	"hmc_server.com/hmc_server/src/JFSM"
)

type GameManager struct {
	Fsm_Wrapper *JFSM.FSM_Wrapper
}

var instance *GameManager
var once sync.Once

func GameManagerInst() *GameManager {
	once.Do(func() {
		instance = &GameManager{}

		states := make(map[string]JFSM.FSM_State_Interface)
		states["fsm1"] = NewIdle()
		instance.Fsm_Wrapper = JFSM.MakeFsm("fsm1", states)
	})
	return instance
}

func (nm *GameManager) Init() {
	fmt.Println("INIT_GameManager")
}

func (gm *GameManager) Update(delta float64) {
	//gm.Fsm_Wrapper.GetCurState().UpdateState(delta)

	//game_net.NetManagerInst().Update(delta)
	//objects.ObjectManagerInst().Update(delta)
}
