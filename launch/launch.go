package launch

import (
	"log"
	"strconv"
	"github.com/jaybennett89/thorium-go/cmd/host-server/hostconf"
	"github.com/jaybennett89/thorium-go/model"
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
var baseListenPort int = 10100

func NewGameServer(machineKey string, servicePort int, gameId int, mapName string, mode string, minLevel int, maxPlayers int) error {
	listenPort := baseListenPort + len(list)
	log.Printf("Starting new game server: \nmachineId %d\nservicePort %d\n listenPort %d\n map %s\n mode %s\n minLevel %d\n maxPlayers %d\n", gameId, servicePort, listenPort, mapName, mode, minLevel, maxPlayers)

	cmd := exec.Command(
		hostconf.GameserverBinaryPath(),
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

		ApplicationName: hostconf.GameserverBinaryPath(),
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
