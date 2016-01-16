package launch

import (
	"log"
	"strconv"
	"thorium-go/model"
)
import "os"
import "os/exec"

type GameServerProcess struct {
	ApplicationName string
	Game            *model.Game
	Process         *os.Process
	ListenPort      int
}

var list []GameServerProcess = make([]GameServerProcess, 0)
var program string = "bin/example-gameserver"
var baseListenPort int = 12690

func NewGameServer(machineKey string, servicePort int, gameId int, mapName string, mode string, minLevel int, maxPlayers int) error {

	log.Printf("Starting new game server (gameId %d, map %s, mode %s, minLevel %d, maxPlayers %d", gameId, mapName, mode, minLevel, maxPlayers)

	listenPort := baseListenPort + len(list)

	cmd := exec.Command(
		program,
		"-key", machineKey,
		"-id", strconv.Itoa(gameId),
		"-listen", strconv.Itoa(listenPort),
		"-service", strconv.Itoa(servicePort),
		"-map", mapName,
		"-mode", mode,
		"-minlvl", strconv.Itoa(minLevel),
		"-maxplayers", strconv.Itoa(maxPlayers),
	)

	// setup log file
	log, err := os.Create("example-gameserver.log")
	if err != nil {
		return err
	}

	cmd.Stdout = log

	err = cmd.Start()
	if err != nil {

		return err
	}

	game := model.Game{

		GameId:         gameId,
		Map:            mapName,
		Mode:           mode,
		MinimumLevel:   minLevel,
		PlayerCount:    0,
		MaximumPlayers: maxPlayers,
	}

	gameServer := GameServerProcess{

		ApplicationName: program,
		Game:            &game,
		Process:         cmd.Process,
		ListenPort:      listenPort,
	}

	list = append(list, gameServer)

	return nil
}

func GetServerList() []GameServerProcess {

	return list
}
