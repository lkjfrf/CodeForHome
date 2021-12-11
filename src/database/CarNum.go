package database

type CarNum struct {
	PlayerId string `dynamodbav:"playerid"`
	CarNum   string `dynamodbav:"carnum"`
}
