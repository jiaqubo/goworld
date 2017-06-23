package server

import (
	"flag"

	"math/rand"
	"time"

	"github.com/xiaonanln/goworld/components/dispatcher/dispatcher_client"
	"github.com/xiaonanln/goworld/config"
	"github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/netutil"
	"github.com/xiaonanln/goworld/proto"
)

var (
	serverid    uint16
	configFile  string
	gameService *GameService
	gateService *GateService
)

func init() {
	parseArgs()
}

func parseArgs() {
	var serveridArg int
	flag.IntVar(&serveridArg, "sid", 0, "set serverid")
	flag.StringVar(&configFile, "configfile", "", "set config file path")
	flag.Parse()
	serverid = uint16(serveridArg)
}

func Run(delegate IServerDelegate) {
	rand.Seed(time.Now().UnixNano())

	if configFile != "" {
		config.SetConfigFile(configFile)
	}

	//f, err := os.OpenFile("server.log", os.O_APPEND|os.O_CREATE, 0644)
	//if err != nil {
	//	panic(err)
	//}
	//gwlog.SetOutput(f)

	entity.CreateSpaceLocally(0) // create to be the nil space
	gameService = newGameService(serverid, delegate)

	dispatcher_client.Initialize(serverid, &dispatcherClientDelegate{})

	gateService = newGateService()
	go gateService.run() // run gate service in another goroutine

	gameService.run()
}

type dispatcherClientDelegate struct {
}

func (delegate *dispatcherClientDelegate) OnDispatcherClientConnect() {
	// called when connected / reconnected to dispatcher (not in main routine)

}
func (delegate *dispatcherClientDelegate) HandleDispatcherClientPacket(msgtype proto.MsgType_t, packet *netutil.Packet) {
	if msgtype < proto.MT_GATE_SERVICE_MSG_TYPE_START {
		gameService.packetQueue <- packetQueueItem{ // may block the dispatcher client routine
			msgtype: msgtype,
			pkt:     packet,
		}
	} else {
		gateService.HandleDispatcherClientPacket(msgtype, packet)
	}
}

func GetServerID() uint16 {
	return serverid
}
