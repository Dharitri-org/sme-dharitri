package connectionMonitor

import (
	"time"

	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/p2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"
)

// DurationBetweenReconnectAttempts is used as to not call reconnecter.ReconnectToNetwork() too often
// when there are a lot of peers disconnecting and reconnection to initial nodes succeeds
var DurationBetweenReconnectAttempts = time.Second * 5

type libp2pConnectionMonitorSimple struct {
	chDoReconnect              chan struct{}
	reconnecter                p2p.Reconnecter
	thresholdMinConnectedPeers int
	sharder                    Sharder
}

// NewLibp2pConnectionMonitorSimple creates a new connection monitor (version 2 that is more streamlined and does not care
// about pausing and resuming the discovery process)
func NewLibp2pConnectionMonitorSimple(
	reconnecter p2p.Reconnecter,
	thresholdMinConnectedPeers int,
	sharder Sharder,
) (*libp2pConnectionMonitorSimple, error) {
	if thresholdMinConnectedPeers < 0 {
		return nil, p2p.ErrInvalidValue
	}
	if check.IfNil(reconnecter) {
		return nil, p2p.ErrNilReconnecter
	}
	if check.IfNil(sharder) {
		return nil, p2p.ErrNilSharder
	}

	cm := &libp2pConnectionMonitorSimple{
		reconnecter:                reconnecter,
		chDoReconnect:              make(chan struct{}),
		thresholdMinConnectedPeers: thresholdMinConnectedPeers,
		sharder:                    sharder,
	}

	if reconnecter != nil {
		go cm.doReconnection()
	}

	return cm, nil
}

// Listen is called when network starts listening on an addr
func (lcms *libp2pConnectionMonitorSimple) Listen(network.Network, multiaddr.Multiaddr) {}

// ListenClose is called when network stops listening on an addr
func (lcms *libp2pConnectionMonitorSimple) ListenClose(network.Network, multiaddr.Multiaddr) {}

// Request a reconnect to initial list
func (lcms *libp2pConnectionMonitorSimple) doReconn() {
	select {
	case lcms.chDoReconnect <- struct{}{}:
	default:
	}
}

// Connected is called when a connection opened
func (lcms *libp2pConnectionMonitorSimple) Connected(netw network.Network, _ network.Conn) {
	allPeers := netw.Peers()

	evicted := lcms.sharder.ComputeEvictionList(allPeers)
	for _, pid := range evicted {
		_ = netw.ClosePeer(pid)
	}
}

// Disconnected is called when a connection closed
func (lcms *libp2pConnectionMonitorSimple) Disconnected(netw network.Network, _ network.Conn) {
	lcms.doReconnectionIfNeeded(netw)
}

func (lcms *libp2pConnectionMonitorSimple) doReconnectionIfNeeded(netw network.Network) {
	if !lcms.IsConnectedToTheNetwork(netw) {
		lcms.doReconn()
	}
}

// OpenedStream is called when a stream opened
func (lcms *libp2pConnectionMonitorSimple) OpenedStream(network.Network, network.Stream) {}

// ClosedStream is called when a stream closed
func (lcms *libp2pConnectionMonitorSimple) ClosedStream(network.Network, network.Stream) {}

func (lcms *libp2pConnectionMonitorSimple) doReconnection() {
	for {
		<-lcms.chDoReconnect
		<-lcms.reconnecter.ReconnectToNetwork()

		time.Sleep(DurationBetweenReconnectAttempts)
	}
}

// IsConnectedToTheNetwork returns true if the number of connected peer is at least equal with thresholdMinConnectedPeers
func (lcms *libp2pConnectionMonitorSimple) IsConnectedToTheNetwork(netw network.Network) bool {
	return len(netw.Peers()) >= lcms.thresholdMinConnectedPeers
}

// SetThresholdMinConnectedPeers sets the minimum connected peers number when the node is considered connected on the network
func (lcms *libp2pConnectionMonitorSimple) SetThresholdMinConnectedPeers(thresholdMinConnectedPeers int, netw network.Network) {
	if check.IfNilReflect(netw) {
		return
	}
	lcms.thresholdMinConnectedPeers = thresholdMinConnectedPeers
	lcms.doReconnectionIfNeeded(netw)
}

// ThresholdMinConnectedPeers returns the minimum connected peers number when the node is considered connected on the network
func (lcms *libp2pConnectionMonitorSimple) ThresholdMinConnectedPeers() int {
	return lcms.thresholdMinConnectedPeers
}

// IsInterfaceNil returns true if there is no value under the interface
func (lcms *libp2pConnectionMonitorSimple) IsInterfaceNil() bool {
	return lcms == nil
}
