package request

import "thorium-go/model"

type CreateNewGame struct {
	SessionKey   string `json:"sessionKey"`
	Map          string `json:"map"`
	GameMode     string `json:"gameMode"`
	MinimumLevel int    `json:"minimumLevel"`
	MaxPlayers   int    `json:"maxPlayers"`
}

type NewGameServer struct {
	Game model.Game `json:"game"`
}

type RegisterGameServer struct {
	MachineId int `json:"machineId"`
	GameId    int `json:"gameId"`
	Port      int `json:"gameListenPort"`
}

type RegisterMachine struct {
	Port int `json:"serviceListenPort"`
}

type UnregisterMachine struct {
	MachineKey string `json:"machineToken"`
}

type MachineStatus struct {
	MachineKey     string  `json:"machineToken"`
	UsageCPU       float64 `json:"cpuUsagePct"`
	UsageNetwork   float64 `json:"networkUsagePct"`
	PlayerCapacity float64 `json:"playerCapacityPct"`
}

type Authentication struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CreateCharacter struct {
	SessionKey string `json:"sessionKey"`
	Name       string `json:"name"`
	ClassId    int    `json:"classId"`
}
