package request

import "github.com/jaybennett89/thorium-go/model"

type LoginResponse struct {
	SessionKey   string `json:"sessionKey"`
	CharacterIDs []int  `json:"characters"`
}

type NewCharacterResponse struct {
	CharacterId int `json:"characterId"`
}

type MachineRegisterResponse struct {
	MachineId  int    `json:"machineId"`
	MachineKey string `json:"machineKey"`
}

type ServerInfoResponse struct {
	RemoteAddress string `json:"remoteAddress"`
	ListenPort    int    `json:"listenPort"`
}

type GetGamesResponse struct {
	List []model.Game `json:"list"`
}

type CreateNewGameResponse struct {
	GameId int `json:"gameId"`
}

type NewGameServerResponse struct {
	MachineKey string `json:"machineKey"`
}

type JoinGameResponse struct {
	RemoteAddress string `json:"remoteAddress"`
	ListenPort    int    `json:"listenPort"`
}

type PlayerConnectResponse struct {
	Character *model.Character `json:"character"`
}
