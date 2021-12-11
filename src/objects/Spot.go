package objects

import "hmc_server.com/hmc_server/src/transform"

type Spot struct {
	Id         string
	SpotId     string
	SpawnPoint transform.Vector3

	LevelName	string
}
