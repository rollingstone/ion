package main

import (
	"net/http"
	_ "net/http/pprof"

	conf "github.com/pion/ion/pkg/conf/sfu"
	"github.com/pion/ion/pkg/discovery"
	"github.com/pion/ion/pkg/log"
	"github.com/pion/ion/pkg/node/sfu"
	"github.com/pion/ion/pkg/rtc"
	"github.com/pion/webrtc/v2"
)

func init() {
	var icePortStart, icePortEnd uint16

	if len(conf.WebRTC.ICEPortRange) == 2 {
		icePortStart = conf.WebRTC.ICEPortRange[0]
		icePortEnd = conf.WebRTC.ICEPortRange[1]
	}

	log.Init(conf.Log.Level)
	var iceServers []webrtc.ICEServer
	for _, iceServer := range conf.WebRTC.ICEServers {
		s := webrtc.ICEServer{
			URLs:       iceServer.URLs,
			Username:   iceServer.Username,
			Credential: iceServer.Credential,
		}
		iceServers = append(iceServers, s)
	}
	if err := rtc.InitIce(iceServers, icePortStart, icePortEnd); err != nil {
		panic(err)
	}

	if err := rtc.InitRTP(conf.Rtp.Port, "", ""); err != nil {
		panic(err)
	}
}

func main() {
	log.Infof("--- Starting SFU Node ---")

	if conf.Global.Pprof != "" {
		go func() {
			log.Infof("Start pprof on %s", conf.Global.Pprof)
			http.ListenAndServe(conf.Global.Pprof, nil)
		}()
	}

	serviceNode := discovery.NewServiceNode(conf.Etcd.Addrs, conf.Global.Dc)
	serviceNode.RegisterNode("sfu", "node-sfu", "sfu-channel-id")

	rpcID := serviceNode.GetRPCChannel()
	eventID := serviceNode.GetEventChannel()
	sfu.Init(conf.Global.Dc, serviceNode.NodeInfo().ID, rpcID, eventID, conf.Nats.URL)
	select {}
}
