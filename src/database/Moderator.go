package database

type Moderator struct {
	PlayerId    string `dynamodbav:"playerid"`
	Region      string `dynamodbav:"region"`
	Content     string `dynamodbav:"content"`
	IsModerator bool   `dynamodbav:"ismoderator"`
	Category    string `dynamodbav:"category"`
	PhotoURL    string `dynamodbav:"photourl"`
	Title       string `dynamodbav:"title"`
	IsLogin     bool   `dynamodbav:"islogin"`
	Password    string `dynamodbav:"password"`
	StartTime   string `dynamodbav:"starttime"`
}
