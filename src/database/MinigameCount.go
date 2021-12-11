package database

type MinigameCount struct {
	Id             string `dynamodbav:"playerid"`
	Region         string `dynamodbav:"region"`
	LastLoginDay   int32  `dynamodbav:"lastloginday"`
	Wrc_Count      int32  `dynamodbav:"wrccount"`
	G80ev_Count    int32  `dynamodbav:"g80count"`
	Ioniq_Count    int32  `dynamodbav:"ioniqcount"`
	Elevate_Count  int32  `dynamodbav:"elevatecount"`
	Minigame_Count int32  `dynamodbav:"minigamecount"`
}
