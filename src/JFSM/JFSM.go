package JFSM

import (
	"github.com/looplab/fsm"
)

type FSM_State_Struct struct {
}

func (f *FSM_State_Struct) InitState() {

}

func (f *FSM_State_Struct) BeginState() {

}

func (f *FSM_State_Struct) EndState() {

}

func (f *FSM_State_Struct) UpdateState(delta float64) {

}

type FSM_State_Interface interface {
	InitState()
	BeginState()
	EndState()
	UpdateState(delta float64)
}

func MakeFsm(initState string, data map[string]FSM_State_Interface) *FSM_Wrapper {
	fsm_events := fsm.Events{}

	for key, _ := range data {
		for key2, _ := range data {
			d := fsm.EventDesc{Name: key + "_to_" + key2, Src: []string{key}, Dst: key2}
			fsm_events = append(fsm_events, d)
		}
	}

	callbacks := fsm.Callbacks{}
	for key, _ := range callbacks {
		callbacks["before_"+key] = func(e *fsm.Event) { data[e.Src].EndState() }
		callbacks["after_"+key] = func(e *fsm.Event) { data[e.Dst].BeginState() }
	}

	fsm := fsm.NewFSM(
		initState,
		fsm_events,
		callbacks,
	)

	wrapper := &FSM_Wrapper{}
	wrapper.Fsm = fsm
	wrapper.States = data

	for _, d := range wrapper.States {
		d.InitState()
	}

	if val, ok := wrapper.States[wrapper.Fsm.Current()]; ok {

		val.BeginState()
	}

	return wrapper
}

type FSM_Wrapper struct {
	Fsm    *fsm.FSM
	States map[string]FSM_State_Interface
}

func (fw *FSM_Wrapper) GetCurState() FSM_State_Interface {
	return fw.States[fw.Fsm.Current()]
}
