package request

import "thorium-go/model"

type LoginResponse struct {
	UserToken    string `json:"userToken"`
	CharacterIDs []int  `json:"characters"`
}

type NewCharacterResponse struct {
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

type GetGamesResponse struct {
	List []model.Game `json:"list"`
}
