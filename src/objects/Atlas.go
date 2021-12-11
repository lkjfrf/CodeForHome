package objects

import "hmc_server.com/hmc_server/src/transform"

type Atlas struct {
	Id         string
	AtlasId    string
	SpawnPoint transform.Vector3
	
	LevelName	string
}
