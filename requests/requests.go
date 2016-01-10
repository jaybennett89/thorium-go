package request

type CreateNewGame struct {
	SessionKey   string `json:"sessionKey"`
	Map          string `json:"map"`
	GameMode     string `json:"gameMode"`
	MinimumLevel int    `json:"minimumLevel"`
	MaxPlayers   int    `json:"maxPlayers"`
}

type NewGameServer struct {
	GameId         int    `json:"gameId"`
	Map            string `json:"map"`
	Mode           string `json:"mode"`
	MinimumLevel   int    `json:"minimumLevel"`
	MaximumPlayers int    `json:"maxPlayers"`
}

type RegisterGameServer struct {
	MachineKey string `json:"machineKey"`
	GameId     int    `json:"gameId"`
	Port       int    `json:"gameListenPort"`
}

type RegisterMachine struct {
	Port int `json:"serviceListenPort"`
}

type UnregisterMachine struct {
	MachineKey string `json:"machineKey"`
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
