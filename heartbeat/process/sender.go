package process

import (
	"fmt"
	"time"

	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/crypto"
	"github.com/Dharitri-org/sme-dharitri/heartbeat"
	"github.com/Dharitri-org/sme-dharitri/heartbeat/data"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/sharding"
)

// ArgHeartbeatSender represents the arguments for the heartbeat sender
type ArgHeartbeatSender struct {
	PeerMessenger        heartbeat.P2PMessenger
	PeerSignatureHandler crypto.PeerSignatureHandler
	PrivKey              crypto.PrivateKey
	Marshalizer          marshal.Marshalizer
	Topic                string
	ShardCoordinator     sharding.Coordinator
	PeerTypeProvider     heartbeat.PeerTypeProviderHandler
	StatusHandler        core.AppStatusHandler
	VersionNumber        string
	NodeDisplayName      string
	KeyBaseIdentity      string
	HardforkTrigger      heartbeat.HardforkTrigger
}

// Sender periodically sends heartbeat messages on a pubsub topic
type Sender struct {
	peerMessenger        heartbeat.P2PMessenger
	peerSignatureHandler crypto.PeerSignatureHandler
	privKey              crypto.PrivateKey
	marshalizer          marshal.Marshalizer
	shardCoordinator     sharding.Coordinator
	peerTypeProvider     heartbeat.PeerTypeProviderHandler
	statusHandler        core.AppStatusHandler
	topic                string
	versionNumber        string
	nodeDisplayName      string
	keyBaseIdentity      string
	hardforkTrigger      heartbeat.HardforkTrigger
}

// NewSender will create a new sender instance
func NewSender(arg ArgHeartbeatSender) (*Sender, error) {
	if check.IfNil(arg.PeerMessenger) {
		return nil, heartbeat.ErrNilMessenger
	}
	if check.IfNil(arg.PeerSignatureHandler) {
		return nil, heartbeat.ErrNilPeerSignatureHandler
	}
	if check.IfNil(arg.PrivKey) {
		return nil, heartbeat.ErrNilPrivateKey
	}
	if check.IfNil(arg.Marshalizer) {
		return nil, heartbeat.ErrNilMarshalizer
	}
	if check.IfNil(arg.ShardCoordinator) {
		return nil, heartbeat.ErrNilShardCoordinator
	}
	if check.IfNil(arg.PeerTypeProvider) {
		return nil, heartbeat.ErrNilPeerTypeProvider
	}
	if check.IfNil(arg.StatusHandler) {
		return nil, heartbeat.ErrNilAppStatusHandler
	}
	if check.IfNil(arg.HardforkTrigger) {
		return nil, heartbeat.ErrNilHardforkTrigger
	}
	err := VerifyHeartbeatProperyLen("application version string", []byte(arg.VersionNumber))
	if err != nil {
		return nil, err
	}

	sender := &Sender{
		peerMessenger:        arg.PeerMessenger,
		peerSignatureHandler: arg.PeerSignatureHandler,
		privKey:              arg.PrivKey,
		marshalizer:          arg.Marshalizer,
		topic:                arg.Topic,
		shardCoordinator:     arg.ShardCoordinator,
		peerTypeProvider:     arg.PeerTypeProvider,
		statusHandler:        arg.StatusHandler,
		versionNumber:        arg.VersionNumber,
		nodeDisplayName:      arg.NodeDisplayName,
		keyBaseIdentity:      arg.KeyBaseIdentity,
		hardforkTrigger:      arg.HardforkTrigger,
	}

	return sender, nil
}

// SendHeartbeat broadcasts a new heartbeat message
func (s *Sender) SendHeartbeat() error {
	hb := &data.Heartbeat{
		Payload:         []byte(fmt.Sprintf("%v", time.Now())),
		ShardID:         s.shardCoordinator.SelfId(),
		VersionNumber:   s.versionNumber,
		NodeDisplayName: s.nodeDisplayName,
		Identity:        s.keyBaseIdentity,
		Pid:             s.peerMessenger.ID().Bytes(),
	}

	triggerMessage, isHardforkTriggered := s.hardforkTrigger.RecordedTriggerMessage()
	if isHardforkTriggered {
		isPayloadRecorder := len(triggerMessage) != 0
		if isPayloadRecorder {
			//beside sending the regular heartbeat message, send also the initial payload hardfork trigger message
			// so that will be spread in an epidemic manner
			s.peerMessenger.Broadcast(s.topic, triggerMessage)
		} else {
			hb.Payload = s.hardforkTrigger.CreateData()
		}
	}

	var err error
	hb.Pubkey, err = s.privKey.GeneratePublic().ToByteArray()
	if err != nil {
		return err
	}

	s.updateMetrics(hb)

	err = verifyLengths(hb)
	if err != nil {
		log.Warn("verify hb length", "error", err.Error())
		trimLengths(hb)
	}

	hb.Signature, err = s.peerSignatureHandler.GetPeerSignature(s.privKey, hb.Pid)
	if err != nil {
		return err
	}

	buffToSend, err := s.marshalizer.Marshal(hb)
	if err != nil {
		return err
	}

	s.peerMessenger.Broadcast(s.topic, buffToSend)

	return nil
}

func (s *Sender) updateMetrics(hb *data.Heartbeat) {
	result := s.computePeerList(hb.Pubkey)

	nodeType := ""
	if result == string(core.ObserverList) {
		nodeType = string(core.NodeTypeObserver)
	} else {
		nodeType = string(core.NodeTypeValidator)
	}

	s.statusHandler.SetStringValue(core.MetricNodeType, nodeType)
	s.statusHandler.SetStringValue(core.MetricPeerType, result)
}

func (s *Sender) computePeerList(pubkey []byte) string {
	peerType, _, err := s.peerTypeProvider.ComputeForPubKey(pubkey)
	if err != nil {
		log.Warn("sender: compute peer type", "error", err)
		return string(core.ObserverList)
	}

	return string(peerType)
}
