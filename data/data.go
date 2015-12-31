package data

// simple types
type ClassId int
type ItemId int

// struct types
type Account struct {
	UserId       int    `json:"uid"`
	Username     string `json:"username"`
	SessionKey   string `json:"sessionKey"`
	CharacterIds []int  `json:"characterIds"`
}

type Character struct {
	CharacterId int     `json:"characterId"`
	Name        string  `json:"name"`
	ClassId     ClassId `json:"classId"`
}
