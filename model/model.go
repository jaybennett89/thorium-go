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
	CharacterState `json:"characterState"`
}

type CharacterState struct {
	ClassId   int   `json:"classId"`
	WeaponId  int   `json:"weaponId"`
	Inventory []int `json:"inventory"`
}

func NewCharacter() *Character {

	var character Character
	character.Inventory = make([]int, 0)
	return &character
}
