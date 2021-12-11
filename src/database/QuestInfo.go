package database

type QuestInfo struct {
	Id                string `dynamodbav:"playerid"`
	Quest_Overview    []bool `dynamodbav:"questoverview"`
	Quest_Brand       []bool `dynamodbav:"questbrand"`
	Quest_Product     []bool `dynamodbav:"questproduct"`
	Quest_LiveStation []bool `dynamodbav:"questlivestation"`
	Quest_RHQ         []bool `dynamodbav:"questrhq"`
	Quest_Fuel        []bool `dynamodbav:"questfuel"`
}
