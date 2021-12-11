package objects

import "hmc_server.com/hmc_server/src/transform"

type TreasureBox struct {
	Id       string
	Point    int32
	Position transform.Vector3
	
	LevelName	string
}
