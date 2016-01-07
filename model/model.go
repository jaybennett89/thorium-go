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
	ClassId        int     `json:"classId"`
	BaseMeshId     int     `json:"baseMeshId"`
	Position       Vector3 `json:"position"`
	Alive          bool    `json:"alive"`
	Health         Vital   `json:"health"`
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

type Game struct {
	GameId         int    `json:"gameId"`
	Map            string `json:"map"`
	Mode           string `json:"mode"`
	MinimumLevel   int    `json:"minimumLevel"`
	PlayerCount    int    `json:"playerCount"`
	MaximumPlayers int    `json:"maxPlayers"`
}

type Machine struct {
	MachineId     int    `json:"machineId"`
	RemoteAddress string `json:"remoteAddress"`
	ListenPort    int    `json:"listenPort"`
	MachineKey    string `json:"machineKey"`
}

func NewCharacter() *Character {

	var character Character
	character.Inventory = make([]Item, 0)
	character.Weapons = make([]int, 0)
	character.SelectedWeapon = -1
	return &character
}

func (c *Character) SetClassAttributes(classId int) {

	c.ClassId = classId
	c.BaseMeshId = 1
	c.Alive = true
	c.Health.Max = 100
	c.Health.Current = 100
	c.Health.RegenRate = 10
	c.Power.Current = 100
	c.Power.Max = 100
	c.Power.RegenRate = 30
	c.BaseMovespeed = 8
	c.Weapons = append(c.Weapons, 1)
	c.SelectedWeapon = 0
}
