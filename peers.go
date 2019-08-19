package main

import (
	"log"
	"github.com/niftynei/glightning/glightning"
	"github.com/niftynei/glightning/jrpc2"
	"os"
)

type PeerRpc struct {
	NoChannel bool `json:"no-channel,omitempty"`
}

func (r *PeerRpc) Name() string {
	return "peers"
}

func (r *PeerRpc) New() interface{} {
	return &PeerRpc{}
}

func (r *PeerRpc) Call() (jrpc2.Result, error) {

	peers, err := lightning.ListPeers()
	if err != nil {
		return nil, err
	}

	nodes, err := lightning.ListNodes()
	if err != nil {
		return nil, err
	}

	lookup := buildLookup(nodes)

	result := &PeerResult{}
	for _, peer := range peers {
		node := lookup[peer.Id]
		if node == nil {
			continue
		}

		hasChannel := len(peer.Channels) > 0
		if r.NoChannel && hasChannel {
			continue
		}
		result.Peers = append(result.Peers, &PeerInfo{
			Alias: node.Alias,
			Connected: peer.Connected,
			HasChannel: hasChannel,
		})
	}

	return result, nil
}

func buildLookup(nodes []glightning.Node) map[string]*glightning.Node {
	result := make(map[string]*glightning.Node)

	for i, node := range nodes {
		result[node.Id] = &nodes[i]
	}
	return result
}

type PeerResult struct {
	Peers []*PeerInfo `json:"peers"`
}

type PeerInfo struct {
	Alias string `json:"alias"`
	Connected bool `json:"is_connected"`
	HasChannel bool `json:"has_channel"`
}

var lightning *glightning.Lightning

func onInit(p *glightning.Plugin, option map[string]string, config *glightning.Config) {
	log.Printf("'peers' plugin initialized")

	lightning = glightning.NewLightning()
	lightning.StartUp(config.RpcFile, config.LightningDir)
}

func main() {
	plugin := glightning.NewPlugin(onInit)

	// add a new RPC method 'peers'
	peerRpc := glightning.NewRpcMethod(&PeerRpc{}, "List peers with aliases")
	peerRpc.LongDesc = "More longer, help description"
	peerRpc.Category = "plugin"

	plugin.RegisterMethod(peerRpc)

	err := plugin.Start(os.Stdin, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
}
