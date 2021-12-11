package core

import (
	"hmc_server.com/hmc_server/src/JFSM"
)

type GM_Idle struct {
	JFSM.FSM_State_Struct
	name string
}

func (f *GM_Idle) InitState() {
	f.FSM_State_Struct.InitState()
}

func (f *GM_Idle) BeginState() {
	f.FSM_State_Struct.BeginState()
}

func (f *GM_Idle) EndState() {
	f.FSM_State_Struct.EndState()
}

func (f *GM_Idle) UpdateState(delta float64) {
	f.FSM_State_Struct.UpdateState(delta)
}

func NewIdle() JFSM.FSM_State_Interface {
	return &GM_Idle{}
}
