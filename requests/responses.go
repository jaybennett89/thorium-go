package request

type LoginResponse struct {
	UserToken    string `json:"userToken"`
	CharacterIDs []int  `json:"characters"`
}

type CharacterSessionResponse struct {
	CharacterId int `json:"characterId"`
}

type MachineRegisterResponse struct {
	MachineId    int    `json:"machineId"`
	MachineToken string `json:"machineToken"`
}

type ServerInfoResponse struct {
	RemoteAddress string `json:"remoteAddress"`
	Port          int    `json:"port"`
}
