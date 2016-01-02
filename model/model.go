package model

type Account struct {
	UserId       int    `json:"uid"`
	Username     string `json:"username"`
	SessionKey   string `json:"sessionKey"`
	CharacterIds []int  `json:"characterIds"`
}

type Character struct {
	CharacterId    int    `json:"characterId"`
	Name           string `json:"name"`
	ClassId        int    `json:"classId"`
	CharacterState `json:"characterState"`
}

type CharacterState struct {
	BaseMeshId     int     `json:"baseMeshId"`
	Position       Vector3 `json:"position"`
	Alive          bool    `json:"alive"`
	Health         Vital   `json:"health"`
	Stamina        Vital   `json:"stamina"`
	Power          Vital   `json:"power"`
	BaseMovespeed  float64 `json:"baseMovespeed"`
	SelectedWeapon int     `json:"selectedWeapon"`
	Weapons        []int   `json:"weapons"`
	Inventory      []Item  `json:"inventory"`
}

type Vector3 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type Item struct {
	ItemId int `json:"itemId"`
	Stacks int `json:"stacks"`
}

type Vital struct {
	Current   float64 `json:"current"`
	Max       float64 `json:"max"`
	RegenRate float64 `json:"regenRate"`
}

func NewCharacter() *Character {

	var character Character
	character.Inventory = make([]Item, 0)
	character.Weapons = make([]int, 0)
	character.SelectedWeapon = -1
	return &character
}
