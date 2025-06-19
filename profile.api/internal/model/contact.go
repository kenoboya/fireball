package model

type Contact struct {
	UserRequest UserRequest `bson:"user_request" json:"user_request"`
	Alias       string      `bson:"alias" json:"alias"`
}
