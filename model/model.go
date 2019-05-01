package model

type Account struct {
	UserId       int    `json:"uid"`
	Username     string `json:"username"`
	SessionKey   string `json:"sessionKey"`
	CharacterIds []int  `json:"characterIds"`
}

type Machine struct {
	MachineId     int    `json:"machineId"`
	RemoteAddress string `json:"remoteAddress"`
	ListenPort    int    `json:"listenPort"`
	MachineKey    string `json:"machineKey"`
}

type HostServer struct {
	GameId        int    `json:"gameId"`
	RemoteAddress string `json:"remoteAddress"`
	ListenPort    int    `json:"listenPort"`
}

type Game struct {
	GameId         int    `json:"gameId"`
	Map            string `json:"map"`
	Mode           string `json:"mode"`
	MinimumLevel   int    `json:"minimumLevel"`
	PlayerCount    int    `json:"playerCount"`
	MaximumPlayers int    `json:"maxPlayers"`
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

type Character struct {
	CharacterId    int    `json:"characterId"`
	Name           string `json:"name"`
	LastGameId     int    `json:"lastGameId"`
	CharacterState `json:"characterState"`
}

type CharacterState struct {
	ClassId        int     `json:"classId"`
	BaseMeshId     int     `json:"baseMeshId"`
	Alive          bool    `json:"alive"`
	Position       Vector3 `json:"position"`
	FacingDir      float64 `json:"facingDir"`
	BaseMovespeed  float64 `json:"baseMovespeed"`
	Level          int     `json:"level"`
	XP             int     `json:"xp"`
	Team           int     `json:"team"`
	Health         Vital   `json:"health"`
	Energy         Vital   `json:"energy"`
	Power          Vital   `json:"power"`
	Armor          int     `json:"armor"`
	Inventory      []Item  `json:"inventory"`
	Weapons        []int   `json:"weapons"`
	SelectedWeapon int     `json:"selectedWeapon"`
	Stunned        bool    `json:"stunned"`
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
	c.Alive = false
	c.BaseMovespeed = 8
	c.Level = 1
	c.XP = 0
	c.Health.Max = 300
	c.Health.Current = 300
	c.Health.RegenRate = 10
	c.Energy.Max = 100
	c.Energy.Current = 100
	c.Energy.RegenRate = 20
	c.Power.Current = 100
	c.Power.Max = 100
	c.Power.RegenRate = 30
	c.Armor = 0
	c.Weapons = append(c.Weapons, 0)
	c.Weapons = append(c.Weapons, 100)
	c.SelectedWeapon = 1
	c.Stunned = false
}
