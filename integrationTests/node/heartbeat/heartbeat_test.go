package heartbeat

import (
	"errors"
	"fmt"
	"testing"
	"time"

	mock2 "github.com/Dharitri-org/sme-dharitri/heartbeat/mock"

	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/crypto"
	"github.com/Dharitri-org/sme-dharitri/crypto/signing"
	"github.com/Dharitri-org/sme-dharitri/crypto/signing/mcl"
	mclsig "github.com/Dharitri-org/sme-dharitri/crypto/signing/mcl/singlesig"
	"github.com/Dharitri-org/sme-dharitri/heartbeat/data"
	"github.com/Dharitri-org/sme-dharitri/heartbeat/process"
	"github.com/Dharitri-org/sme-dharitri/integrationTests"
	"github.com/Dharitri-org/sme-dharitri/integrationTests/mock"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/p2p"
	"github.com/Dharitri-org/sme-dharitri/sharding"
	logger "github.com/Dharitri-org/sme-logger"
	"github.com/stretchr/testify/assert"
)

var stepDelay = time.Second
var log = logger.GetOrCreate("integrationtests/node")

// TestHeartbeatMonitorWillUpdateAnInactivePeer test what happen if a peer out of 2 stops being responsive on heartbeat status
// The active monitor should change it's active flag to false when a new heartbeat message has arrived.
func TestHeartbeatMonitorWillUpdateAnInactivePeer(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap()
	advertiserAddr := integrationTests.GetConnectableAddress(advertiser)
	maxUnresposiveTime := time.Second * 10

	monitor := createMonitor(maxUnresposiveTime)
	nodes, senders, pks := prepareNodes(advertiserAddr, monitor, 3, "nodeName")

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Close()
		}
	}()

	fmt.Println("Delaying for node bootstrap and topic announcement...")
	time.Sleep(integrationTests.P2pBootstrapDelay)

	fmt.Println("Sending first messages from both public keys...")
	err := senders[0].SendHeartbeat()
	log.LogIfError(err)

	err = senders[1].SendHeartbeat()
	log.LogIfError(err)

	time.Sleep(stepDelay)

	fmt.Println("Checking both public keys are active...")
	checkReceivedMessages(t, monitor, pks, []int{0, 1})

	fmt.Println("Waiting for max unresponsive time...")
	time.Sleep(maxUnresposiveTime)

	fmt.Println("Only first pk will send another message...")
	_ = senders[0].SendHeartbeat()

	time.Sleep(stepDelay)

	fmt.Println("Checking only first pk is active...")
	checkReceivedMessages(t, monitor, pks, []int{0})
}

func TestHeartbeatMonitorWillNotUpdateTooLongHeartbeatMessages(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap()
	advertiserAddr := integrationTests.GetConnectableAddress(advertiser)
	maxUnresposiveTime := time.Second * 10

	length := 129
	buff := make([]byte, length)

	for i := 0; i < length; i++ {
		buff[i] = byte(97)
	}
	bigNodeName := string(buff)

	monitor := createMonitor(maxUnresposiveTime)
	nodes, senders, pks := prepareNodes(advertiserAddr, monitor, 3, bigNodeName)

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Close()
		}
	}()

	fmt.Println("Delaying for node bootstrap and topic announcement...")
	time.Sleep(integrationTests.P2pBootstrapDelay)

	fmt.Println("Sending first messages with long name...")
	_ = senders[1].SendHeartbeat()

	time.Sleep(stepDelay)

	secondPK := pks[1]

	pkHeartBeats := monitor.GetHeartbeats()

	assert.True(t, isPkActive(pkHeartBeats, secondPK))
	expectedLen := 128
	assert.True(t, isMessageCorrectLen(pkHeartBeats, secondPK, expectedLen))
}

func prepareNodes(
	advertiserAddr string,
	monitor *process.Monitor,
	interactingNodes int,
	defaultNodeName string,
) ([]p2p.Messenger, []*process.Sender, []crypto.PublicKey) {

	senderIdxs := []int{0, 1}
	nodes := make([]p2p.Messenger, interactingNodes)
	topicHeartbeat := "topic"
	senders := make([]*process.Sender, 0)
	pks := make([]crypto.PublicKey, 0)

	for i := 0; i < interactingNodes; i++ {
		nodes[i] = integrationTests.CreateMessengerWithKadDht(advertiserAddr)
		_ = nodes[i].CreateTopic(topicHeartbeat, true)

		isSender := integrationTests.IsIntInSlice(i, senderIdxs)
		if isSender {
			sender, pk := createSenderWithName(nodes[i], topicHeartbeat, defaultNodeName)
			senders = append(senders, sender)
			pks = append(pks, pk)
		} else {
			_ = nodes[i].RegisterMessageProcessor(topicHeartbeat, monitor)
		}

		_ = nodes[i].Bootstrap()
	}

	return nodes, senders, pks
}

func checkReceivedMessages(t *testing.T, monitor *process.Monitor, pks []crypto.PublicKey, activeIdxs []int) {
	pkHeartBeats := monitor.GetHeartbeats()

	extraPkInMonitor := 1
	assert.Equal(t, len(pks), len(pkHeartBeats)-extraPkInMonitor)

	for idx, pk := range pks {
		pkShouldBeActive := integrationTests.IsIntInSlice(idx, activeIdxs)
		assert.Equal(t, pkShouldBeActive, isPkActive(pkHeartBeats, pk))
		assert.True(t, isMessageReceived(pkHeartBeats, pk))
	}
}

func isMessageReceived(heartbeats []data.PubKeyHeartbeat, pk crypto.PublicKey) bool {
	pkBytes, _ := pk.ToByteArray()

	for _, hb := range heartbeats {
		isPkMatching := hb.PublicKey == integrationTests.TestValidatorPubkeyConverter.Encode(pkBytes)
		if isPkMatching {
			return true
		}
	}

	return false
}

func isPkActive(heartbeats []data.PubKeyHeartbeat, pk crypto.PublicKey) bool {
	pkBytes, _ := pk.ToByteArray()

	for _, hb := range heartbeats {
		isPkMatchingAndActve := hb.PublicKey == integrationTests.TestValidatorPubkeyConverter.Encode(pkBytes) && hb.IsActive
		if isPkMatchingAndActve {
			return true
		}
	}

	return false
}

func isMessageCorrectLen(heartbeats []data.PubKeyHeartbeat, pk crypto.PublicKey, expectedLen int) bool {
	pkBytes, _ := pk.ToByteArray()

	for _, hb := range heartbeats {
		isPkMatching := hb.PublicKey == integrationTests.TestValidatorPubkeyConverter.Encode(pkBytes)
		if isPkMatching {
			return len(hb.NodeDisplayName) == expectedLen
		}
	}

	return false
}

func createSenderWithName(messenger p2p.Messenger, topic string, nodeName string) (*process.Sender, crypto.PublicKey) {
	suite := mcl.NewSuiteBLS12()
	signer := &mclsig.BlsSingleSigner{}
	keyGen := signing.NewKeyGenerator(suite)
	sk, pk := keyGen.GeneratePair()
	version := "v01"

	argSender := process.ArgHeartbeatSender{
		PeerMessenger:        messenger,
		PeerSignatureHandler: &mock2.PeerSignatureHandler{Signer: signer},
		PrivKey:              sk,
		Marshalizer:          integrationTests.TestMarshalizer,
		Topic:                topic,
		ShardCoordinator:     &sharding.OneShardCoordinator{},
		PeerTypeProvider:     &mock.PeerTypeProviderStub{},
		StatusHandler:        &mock.AppStatusHandlerStub{},
		VersionNumber:        version,
		NodeDisplayName:      nodeName,
		HardforkTrigger:      &mock.HardforkTriggerStub{},
	}

	sender, _ := process.NewSender(argSender)
	return sender, pk
}

func createMonitor(maxDurationPeerUnresponsive time.Duration) *process.Monitor {
	suite := mcl.NewSuiteBLS12()
	singlesigner := &mclsig.BlsSingleSigner{}
	keyGen := signing.NewKeyGenerator(suite)
	marshalizer := &marshal.GogoProtoMarshalizer{}

	mp, _ := process.NewMessageProcessor(
		&mock2.PeerSignatureHandler{Signer: singlesigner, KeyGen: keyGen},
		marshalizer,
		&mock.NetworkShardingCollectorStub{
			UpdatePeerIdPublicKeyCalled: func(pid core.PeerID, pk []byte) {},
			UpdatePeerIdShardIdCalled:   func(pid core.PeerID, shardId uint32) {},
		})

	argMonitor := process.ArgHeartbeatMonitor{
		Marshalizer:                 integrationTests.TestMarshalizer,
		MaxDurationPeerUnresponsive: maxDurationPeerUnresponsive,
		PubKeysMap:                  map[uint32][]string{0: {""}},
		GenesisTime:                 time.Now(),
		MessageHandler:              mp,
		Storer: &mock.HeartbeatStorerStub{
			UpdateGenesisTimeCalled: func(genesisTime time.Time) error {
				return nil
			},
			LoadHeartBeatDTOCalled: func(pubKey string) (*data.HeartbeatDTO, error) {
				return nil, errors.New("not found")
			},
			LoadKeysCalled: func() ([][]byte, error) {
				return nil, nil
			},
			SavePubkeyDataCalled: func(pubkey []byte, heartbeat *data.HeartbeatDTO) error {
				return nil
			},
			SaveKeysCalled: func(peersSlice [][]byte) error {
				return nil
			},
		},
		PeerTypeProvider: &mock.PeerTypeProviderStub{},
		Timer:            &process.RealTimer{},
		AntifloodHandler: &mock.P2PAntifloodHandlerStub{
			CanProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
				return nil
			},
		},
		HardforkTrigger:                    &mock.HardforkTriggerStub{},
		ValidatorPubkeyConverter:           integrationTests.TestValidatorPubkeyConverter,
		HeartbeatRefreshIntervalInSec:      1,
		HideInactiveValidatorIntervalInSec: 600,
	}

	monitor, _ := process.NewMonitor(argMonitor)

	return monitor
}
