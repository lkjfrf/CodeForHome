package database

type WRCRank struct {
	// Playerid = Email
	UserName string `dynamodbav:"username"`
	Laptime  int    `dynamodbav:"laptime"`
}
