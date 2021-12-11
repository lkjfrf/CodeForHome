package objects

import "sync"

type ChatGroup struct {
	GroupId  string
	OwnerId  string
	Members  sync.Map
	Password string
}

func (g *ChatGroup) GetGroupMembers(GroupId string) sync.Map {
	return g.Members
}
