package thordb

import (
	"github.com/jaybennett89/thorium-go/generate"
	"time"
)

type AccountSession struct {
	UserID int    `json:"uid"`
	Token  string `json:"token"`
	Account
}

type Account struct {
	UserID         int       `json:"uid"`
	CharacterIDs   []int     `json:"characters"`
	HashedPassword []byte    `json:"hashedPassword"`
	Salt           []byte    `json:"salt"`
	Algorithm      string    `json:"hashAlgorithm"`
	CreatedOn      time.Time `json:"createdOn"`
	LastLogin      time.Time `json:"lastLogin"`
}

type AccountPublicView struct {
	UserID       int       `json:"uid"`
	CharacterIDs []int     `json:"characters"`
	CreatedOn    time.Time `json:"createdOn"`
	LastLogin    time.Time `json:"lastLogin"`
}

func (account *Account) NewPublicView() AccountPublicView {

	var publicView AccountPublicView
	publicView.UserID = account.UserID
	publicView.CharacterIDs = account.CharacterIDs
	publicView.CreatedOn = account.CreatedOn
	publicView.LastLogin = account.LastLogin
	return publicView
}

func (account *Account) Validate() (bool, error) {
	// helper function to keep basic validation here?
	return true, nil
}

type CharacterSession struct {
	ID     int    `json:"id"`
	UserID int    `json:"uid"`
	Token  string `json:"charSessionToken"`
	GameId int    `json:"gameId"`
	*CharacterData
}

type Vector3 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type CharacterData struct {
	Name          string                `json:"name"`
	Soul          int                   `json:"soul"`
	World         generate.Coordinate2D `json:"worldCoord"`
	Position      Vector3               `json:"worldPosition"`
	Weapons       []int                 `json:"weapons"`
	Inventory     []int                 `json:"inventory"`
	Health        float64               `json:"health"`
	Power         float64               `json:"powerLevel"`
	Experience    int                   `json:"experienceLevel"`
	MasteryPoints []int                 `json"masteryPoints"`
}

func NewCharacterSession() *CharacterSession {
	var charSession CharacterSession
	charSession.CharacterData = NewCharacterData()
	return &charSession
}

func NewCharacterSessionFrom(character *CharacterData) *CharacterSession {
	var charSession CharacterSession
	charSession.CharacterData = character
	return &charSession
}

func NewCharacterData() *CharacterData {

	var character CharacterData
	character.Weapons = make([]int, 2)
	character.Inventory = make([]int, 25)
	character.MasteryPoints = make([]int, 10)
	return &character
}

type World struct {
	ID         int
	CreatedOn  time.Time
	LastUpdate time.Time
	WorldData
}
type WorldData struct {
	Name        string                `json:"worldName"`
	Location    generate.Coordinate2D `json:"worldCoordinate"`
	Soul        int                   `json:"soul"`
	TerrainType int                   `json:"terrainType"`
	GameMode    int                   `json:"gameMode"`
	Players     []CharacterSession    `json:"playerSessions"`
	NPCs        []CharacterData       `json:"nonPlayerCharacters"`
}
