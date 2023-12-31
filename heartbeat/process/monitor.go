package process

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/heartbeat"
	"github.com/Dharitri-org/sme-dharitri/heartbeat/data"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/p2p"
	"github.com/Dharitri-org/sme-dharitri/statusHandler"
	logger "github.com/Dharitri-org/sme-logger"
)

var log = logger.GetOrCreate("heartbeat/process")

// ArgHeartbeatMonitor represents the arguments for the heartbeat monitor
type ArgHeartbeatMonitor struct {
	Marshalizer                        marshal.Marshalizer
	MaxDurationPeerUnresponsive        time.Duration
	PubKeysMap                         map[uint32][]string
	GenesisTime                        time.Time
	MessageHandler                     heartbeat.MessageHandler
	Storer                             heartbeat.HeartbeatStorageHandler
	PeerTypeProvider                   heartbeat.PeerTypeProviderHandler
	Timer                              heartbeat.Timer
	AntifloodHandler                   heartbeat.P2PAntifloodHandler
	HardforkTrigger                    heartbeat.HardforkTrigger
	ValidatorPubkeyConverter           core.PubkeyConverter
	HeartbeatRefreshIntervalInSec      uint32
	HideInactiveValidatorIntervalInSec uint32
}

// Monitor represents the heartbeat component that processes received heartbeat messages
type Monitor struct {
	maxDurationPeerUnresponsive        time.Duration
	marshalizer                        marshal.Marshalizer
	peerTypeProvider                   heartbeat.PeerTypeProviderHandler
	mutHeartbeatMessages               sync.RWMutex
	mutAppStatusHandler                sync.Mutex
	heartbeatMessages                  map[string]*heartbeatMessageInfo
	pubKeysMap                         map[uint32][]string
	mutFullPeersSlice                  sync.RWMutex
	fullPeersSlice                     [][]byte
	appStatusHandler                   core.AppStatusHandler
	genesisTime                        time.Time
	messageHandler                     heartbeat.MessageHandler
	storer                             heartbeat.HeartbeatStorageHandler
	timer                              heartbeat.Timer
	antifloodHandler                   heartbeat.P2PAntifloodHandler
	hardforkTrigger                    heartbeat.HardforkTrigger
	validatorPubkeyConverter           core.PubkeyConverter
	heartbeatRefreshIntervalInSec      uint32
	hideInactiveValidatorIntervalInSec uint32
}

// NewMonitor returns a new monitor instance
func NewMonitor(arg ArgHeartbeatMonitor) (*Monitor, error) {
	if check.IfNil(arg.Marshalizer) {
		return nil, heartbeat.ErrNilMarshalizer
	}
	if check.IfNil(arg.PeerTypeProvider) {
		return nil, heartbeat.ErrNilPeerTypeProvider
	}
	if arg.PubKeysMap == nil {
		return nil, heartbeat.ErrNilPublicKeysMap
	}
	if check.IfNil(arg.MessageHandler) {
		return nil, heartbeat.ErrNilMessageHandler
	}
	if check.IfNil(arg.Storer) {
		return nil, heartbeat.ErrNilHeartbeatStorer
	}
	if check.IfNil(arg.Timer) {
		return nil, heartbeat.ErrNilTimer
	}
	if check.IfNil(arg.AntifloodHandler) {
		return nil, heartbeat.ErrNilAntifloodHandler
	}
	if check.IfNil(arg.HardforkTrigger) {
		return nil, heartbeat.ErrNilHardforkTrigger
	}
	if check.IfNil(arg.ValidatorPubkeyConverter) {
		return nil, heartbeat.ErrNilPubkeyConverter
	}
	if arg.HeartbeatRefreshIntervalInSec == 0 {
		return nil, heartbeat.ErrZeroHeartbeatRefreshIntervalInSec
	}
	if arg.HideInactiveValidatorIntervalInSec == 0 {
		return nil, heartbeat.ErrZeroHideInactiveValidatorIntervalInSec
	}

	mon := &Monitor{
		marshalizer:                        arg.Marshalizer,
		heartbeatMessages:                  make(map[string]*heartbeatMessageInfo),
		peerTypeProvider:                   arg.PeerTypeProvider,
		maxDurationPeerUnresponsive:        arg.MaxDurationPeerUnresponsive,
		appStatusHandler:                   &statusHandler.NilStatusHandler{},
		genesisTime:                        arg.GenesisTime,
		messageHandler:                     arg.MessageHandler,
		storer:                             arg.Storer,
		timer:                              arg.Timer,
		antifloodHandler:                   arg.AntifloodHandler,
		hardforkTrigger:                    arg.HardforkTrigger,
		validatorPubkeyConverter:           arg.ValidatorPubkeyConverter,
		heartbeatRefreshIntervalInSec:      arg.HeartbeatRefreshIntervalInSec,
		hideInactiveValidatorIntervalInSec: arg.HideInactiveValidatorIntervalInSec,
	}

	err := mon.storer.UpdateGenesisTime(arg.GenesisTime)
	if err != nil {
		return nil, err
	}

	err = mon.initializeHeartbeatMessagesInfo(arg.PubKeysMap)
	if err != nil {
		return nil, err
	}

	err = mon.loadRestOfPubKeysFromStorage()
	if err != nil {
		log.Debug("heartbeat can't load public keys from storage", "error", err.Error())
	}

	mon.startValidatorProcessing()

	return mon, nil
}

func (m *Monitor) initializeHeartbeatMessagesInfo(pubKeysMap map[uint32][]string) error {
	pubKeysMapCopy := make(map[uint32][]string)
	pubKeysToSave := make(map[string]*heartbeatMessageInfo)
	for shardId, pubKeys := range pubKeysMap {
		for _, pubkey := range pubKeys {
			e := m.initializeHeartBeatForPK(pubkey, shardId, pubKeysToSave, pubKeysMapCopy)
			if e != nil {
				return e
			}
		}
	}

	go m.SaveMultipleHeartbeatMessageInfos(pubKeysToSave)

	m.pubKeysMap = pubKeysMapCopy
	return nil
}

func (m *Monitor) initializeHeartBeatForPK(
	pubkey string,
	shardId uint32,
	pubKeysToSave map[string]*heartbeatMessageInfo,
	pubKeysMapCopy map[uint32][]string,
) error {
	hbmi, err := m.loadHeartbeatsFromStorer(pubkey)
	if err != nil { // if pubKey not found in DB, create a new instance
		peerType, shardID := m.computePeerTypeAndShardID([]byte(pubkey))
		hbmi, err = newHeartbeatMessageInfo(m.maxDurationPeerUnresponsive, peerType, m.genesisTime, m.timer)
		if err != nil {
			return err
		}

		hbmi.genesisTime = m.genesisTime
		hbmi.computedShardID = shardID
		pubKeysToSave[pubkey] = hbmi
	}
	m.heartbeatMessages[pubkey] = hbmi
	pubKeysMapCopy[shardId] = append(pubKeysMapCopy[shardId], pubkey)
	return nil
}

// SaveMultipleHeartbeatMessageInfos stores all heartbeatMessageInfos to the storer
func (m *Monitor) SaveMultipleHeartbeatMessageInfos(pubKeysToSave map[string]*heartbeatMessageInfo) {
	m.mutHeartbeatMessages.RLock()
	defer m.mutHeartbeatMessages.RUnlock()

	for key, hmbi := range pubKeysToSave {
		hbDTO := m.convertToExportedStruct(hmbi)
		err := m.storer.SavePubkeyData([]byte(key), &hbDTO)
		if err != nil {
			log.Debug("cannot save heartbeat to db", "error", err.Error())
		}
	}
}

func (m *Monitor) loadRestOfPubKeysFromStorage() error {
	peersSlice, err := m.storer.LoadKeys()
	if err != nil {
		return err
	}

	for _, peer := range peersSlice {
		pubKey := string(peer)
		_, ok := m.heartbeatMessages[pubKey]
		if !ok { // peer not in nodes map
			hbmi, err1 := m.loadHeartbeatsFromStorer(pubKey)
			if err1 != nil {
				continue
			}
			m.heartbeatMessages[pubKey] = hbmi
		}
	}

	return nil
}

func (m *Monitor) loadHeartbeatsFromStorer(pubKey string) (*heartbeatMessageInfo, error) {
	heartbeatDTO, err := m.storer.LoadHeartBeatDTO(pubKey)
	if err != nil {
		return nil, err
	}

	receivedHbmi := m.convertFromExportedStruct(*heartbeatDTO, m.maxDurationPeerUnresponsive)
	receivedHbmi.getTimeHandler = m.timer.Now
	crtTime := m.timer.Now()
	crtDuration := crtTime.Sub(receivedHbmi.lastUptimeDowntime)
	crtDuration = maxDuration(0, crtDuration)
	if receivedHbmi.isActive {
		receivedHbmi.totalUpTime += crtDuration
		receivedHbmi.timeStamp = crtTime
	} else {
		receivedHbmi.totalDownTime += crtDuration
	}
	receivedHbmi.lastUptimeDowntime = crtTime
	receivedHbmi.genesisTime = m.genesisTime

	return receivedHbmi, nil
}

// SetAppStatusHandler will set the AppStatusHandler which will be used for monitoring
func (m *Monitor) SetAppStatusHandler(ash core.AppStatusHandler) error {
	if check.IfNil(ash) {
		return heartbeat.ErrNilAppStatusHandler
	}

	m.mutAppStatusHandler.Lock()
	m.appStatusHandler = ash
	m.mutAppStatusHandler.Unlock()
	return nil
}

// ProcessReceivedMessage satisfies the p2p.MessageProcessor interface so it can be called
// by the p2p subsystem each time a new heartbeat message arrives
func (m *Monitor) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
	if check.IfNil(message) {
		return heartbeat.ErrNilMessage
	}
	if message.Data() == nil {
		return heartbeat.ErrNilDataToProcess
	}

	err := m.antifloodHandler.CanProcessMessage(message, fromConnectedPeer)
	if err != nil {
		return err
	}
	err = m.antifloodHandler.CanProcessMessagesOnTopic(fromConnectedPeer, core.HeartbeatTopic, 1, uint64(len(message.Data())), message.SeqNo())
	if err != nil {
		return err
	}

	hbRecv, err := m.messageHandler.CreateHeartbeatFromP2PMessage(message)
	if err != nil {
		//this situation is so severe that we have to black list both the message originator and the connected peer
		//that disseminated this message.

		reason := "blacklisted due to invalid heartbeat message"
		m.antifloodHandler.BlacklistPeer(message.Peer(), reason, core.InvalidMessageBlacklistDuration)
		m.antifloodHandler.BlacklistPeer(fromConnectedPeer, reason, core.InvalidMessageBlacklistDuration)

		return err
	}

	isHardforkTrigger, err := m.hardforkTrigger.TriggerReceived(message.Data(), hbRecv.Payload, hbRecv.Pubkey)
	if isHardforkTrigger {
		return err
	}

	if !bytes.Equal(hbRecv.Pid, message.Peer().Bytes()) {
		//this situation is so severe that we have to black list both the message originator and the connected peer
		//that disseminated this message.
		reason := "blacklisted due to inconsistent heartbeat message"
		m.antifloodHandler.BlacklistPeer(message.Peer(), reason, core.InvalidMessageBlacklistDuration)
		m.antifloodHandler.BlacklistPeer(fromConnectedPeer, reason, core.InvalidMessageBlacklistDuration)

		return fmt.Errorf("%w heartbeat pid %s, message pid %s",
			heartbeat.ErrHeartbeatPidMismatch,
			p2p.PeerIdToShortString(core.PeerID(hbRecv.Pid)),
			p2p.PeerIdToShortString(message.Peer()),
		)
	}

	//message is validated, process should be done async, method can return nil
	go m.addHeartbeatMessageToMap(hbRecv)

	go m.computeAllHeartbeatMessages()

	return nil
}

func (m *Monitor) addHeartbeatMessageToMap(hb *data.Heartbeat) {
	pubKeyStr := string(hb.Pubkey)
	m.mutHeartbeatMessages.Lock()
	hbmi, ok := m.heartbeatMessages[pubKeyStr]
	if hbmi == nil || !ok {
		var err error
		peerType, _ := m.computePeerTypeAndShardID(hb.Pubkey)
		hbmi, err = newHeartbeatMessageInfo(m.maxDurationPeerUnresponsive, peerType, m.genesisTime, m.timer)
		if err != nil {
			log.Debug("error creating heartbeat message info", "error", err.Error())
			m.mutHeartbeatMessages.Unlock()
			return
		}
		m.heartbeatMessages[pubKeyStr] = hbmi
	}
	m.mutHeartbeatMessages.Unlock()

	peerType, computedShardID := m.computePeerTypeAndShardID(hb.Pubkey)

	hbmi.HeartbeatReceived(computedShardID, hb.ShardID, hb.VersionNumber, hb.NodeDisplayName, hb.Identity, peerType)
	hbDTO := m.convertToExportedStruct(hbmi)

	err := m.storer.SavePubkeyData(hb.Pubkey, &hbDTO)
	if err != nil {
		log.Debug("cannot save heartbeat to db", "error", err.Error())
	}
	m.addPeerToFullPeersSlice(hb.Pubkey)
}

func (m *Monitor) addPeerToFullPeersSlice(pubKey []byte) {
	m.mutFullPeersSlice.Lock()
	defer m.mutFullPeersSlice.Unlock()
	if !m.isPeerInFullPeersSlice(pubKey) {
		m.fullPeersSlice = append(m.fullPeersSlice, pubKey)
		err := m.storer.SaveKeys(m.fullPeersSlice)
		if err != nil {
			log.Debug("can't store the keys slice", "error", err.Error())
		}
	}
}

func (m *Monitor) isPeerInFullPeersSlice(pubKey []byte) bool {
	for _, peer := range m.fullPeersSlice {
		if bytes.Equal(peer, pubKey) {
			return true
		}
	}

	return false
}

func (m *Monitor) computePeerTypeAndShardID(pubkey []byte) (string, uint32) {
	peerType, shardID, err := m.peerTypeProvider.ComputeForPubKey(pubkey)
	if err != nil {
		log.Warn("monitor: compute peer type and shard", "error", err)
		return string(core.ObserverList), 0
	}

	return string(peerType), shardID
}

func (m *Monitor) computeAllHeartbeatMessages() {
	m.mutHeartbeatMessages.Lock()
	counterActiveValidators := 0
	counterConnectedNodes := 0
	hbChangedStateToInactiveMap := make(map[string]*heartbeatMessageInfo)
	for key, v := range m.heartbeatMessages {
		previousActive := v.GetIsActive()
		v.ComputeActive(m.timer.Now())
		isActive := v.GetIsActive()

		if isActive {
			counterConnectedNodes++

			if v.GetIsValidator() {
				counterActiveValidators++
			}
		}
		changedStateToInactive := previousActive && !isActive
		if changedStateToInactive {
			hbChangedStateToInactiveMap[key] = v
		}
	}

	m.mutHeartbeatMessages.Unlock()
	go m.SaveMultipleHeartbeatMessageInfos(hbChangedStateToInactiveMap)

	m.mutAppStatusHandler.Lock()
	m.appStatusHandler.SetUInt64Value(core.MetricLiveValidatorNodes, uint64(counterActiveValidators))
	m.appStatusHandler.SetUInt64Value(core.MetricConnectedNodes, uint64(counterConnectedNodes))
	m.mutAppStatusHandler.Unlock()
}

func (m *Monitor) computeInactiveHeartbeatMessages() {
	m.mutHeartbeatMessages.Lock()
	inactiveHbChangedMap := make(map[string]*heartbeatMessageInfo)
	for key, v := range m.heartbeatMessages {
		isActive := v.GetIsActive()
		if isActive {
			continue
		}

		peerType, shardId := m.computePeerTypeAndShardID([]byte(key))
		if v.peerType != peerType || v.computedShardID != shardId {
			v.UpdateShardAndPeerType(shardId, peerType)
			inactiveHbChangedMap[key] = v
		}
	}

	peerTypeInfos := m.peerTypeProvider.GetAllPeerTypeInfos()
	for _, peerTypeInfo := range peerTypeInfos {
		if m.heartbeatMessages[peerTypeInfo.PublicKey] == nil {
			hbmi, err := newHeartbeatMessageInfo(m.maxDurationPeerUnresponsive, peerTypeInfo.PeerType, m.genesisTime, m.timer)
			if err != nil {
				log.Debug("could not create hbmi ", "err", err)
				continue
			}
			hbmi.computedShardID = peerTypeInfo.ShardId
			m.heartbeatMessages[peerTypeInfo.PublicKey] = hbmi
		}
	}

	m.mutHeartbeatMessages.Unlock()
	go m.SaveMultipleHeartbeatMessageInfos(inactiveHbChangedMap)
}

// GetHeartbeats returns the heartbeat status
func (m *Monitor) GetHeartbeats() []data.PubKeyHeartbeat {
	m.mutHeartbeatMessages.Lock()
	status := make([]data.PubKeyHeartbeat, 0, len(m.heartbeatMessages))
	for k, v := range m.heartbeatMessages {
		if m.shouldSkipValidator(v) {
			delete(m.heartbeatMessages, k)
			continue
		}

		tmp := data.PubKeyHeartbeat{
			PublicKey: m.validatorPubkeyConverter.Encode([]byte(k)),
			TimeStamp: v.timeStamp,
			MaxInactiveTime: data.Duration{
				Duration: v.maxInactiveTime,
			},
			IsActive:        v.isActive,
			ReceivedShardID: v.receivedShardID,
			ComputedShardID: v.computedShardID,
			TotalUpTime:     int64(v.totalUpTime.Seconds()),
			TotalDownTime:   int64(v.totalDownTime.Seconds()),
			VersionNumber:   v.versionNumber,
			NodeDisplayName: v.nodeDisplayName,
			Identity:        v.identity,
			PeerType:        v.peerType,
		}
		status = append(status, tmp)
	}
	m.mutHeartbeatMessages.Unlock()

	sort.Slice(status, func(i, j int) bool {
		return strings.Compare(status[i].PublicKey, status[j].PublicKey) < 0
	})

	return status
}

func (m *Monitor) shouldSkipValidator(v *heartbeatMessageInfo) bool {
	isInactiveObserver := !v.GetIsActive() &&
		(v.peerType != string(core.EligibleList) &&
			v.peerType != string(core.WaitingList))
	if isInactiveObserver {
		lastInactiveInterval := m.timer.Now().Sub(v.timeStamp)
		if lastInactiveInterval.Seconds() > float64(m.hideInactiveValidatorIntervalInSec) {
			return true
		}
	}

	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (m *Monitor) IsInterfaceNil() bool {
	return m == nil
}

func (m *Monitor) convertToExportedStruct(v *heartbeatMessageInfo) data.HeartbeatDTO {
	v.updateMutex.Lock()
	defer v.updateMutex.Unlock()
	ret := data.HeartbeatDTO{
		IsActive:        v.isActive,
		ReceivedShardID: v.receivedShardID,
		ComputedShardID: v.computedShardID,
		VersionNumber:   v.versionNumber,
		NodeDisplayName: v.nodeDisplayName,
		Identity:        v.identity,
		PeerType:        v.peerType,
	}

	ret.TimeStamp = v.timeStamp.UnixNano()
	ret.MaxInactiveTime = v.maxInactiveTime.Nanoseconds()
	ret.TotalUpTime = v.totalUpTime.Nanoseconds()
	ret.TotalDownTime = v.totalDownTime.Nanoseconds()
	ret.LastUptimeDowntime = v.lastUptimeDowntime.UnixNano()
	ret.GenesisTime = v.genesisTime.UnixNano()

	return ret
}

func (m *Monitor) convertFromExportedStruct(hbDTO data.HeartbeatDTO, maxDuration time.Duration) *heartbeatMessageInfo {
	hbmi := &heartbeatMessageInfo{
		maxDurationPeerUnresponsive: maxDuration,
		isActive:                    hbDTO.IsActive,
		receivedShardID:             hbDTO.ReceivedShardID,
		computedShardID:             hbDTO.ComputedShardID,
		versionNumber:               hbDTO.VersionNumber,
		nodeDisplayName:             hbDTO.NodeDisplayName,
		identity:                    hbDTO.Identity,
		peerType:                    hbDTO.PeerType,
	}

	hbmi.maxInactiveTime = time.Duration(hbDTO.MaxInactiveTime)
	hbmi.timeStamp = time.Unix(0, hbDTO.TimeStamp)
	hbmi.totalUpTime = time.Duration(hbDTO.TotalUpTime)
	hbmi.totalDownTime = time.Duration(hbDTO.TotalDownTime)
	hbmi.lastUptimeDowntime = time.Unix(0, hbDTO.LastUptimeDowntime)
	hbmi.genesisTime = time.Unix(0, hbDTO.GenesisTime)

	return hbmi
}

// startValidatorProcessing will start the updating of the information about the nodes
func (m *Monitor) startValidatorProcessing() {
	go func() {
		refreshInterval := time.Duration(m.heartbeatRefreshIntervalInSec) * time.Second

		for {
			m.refreshHeartbeatMessageInfo()
			time.Sleep(refreshInterval)
		}
	}()
}

func (m *Monitor) refreshHeartbeatMessageInfo() {
	m.computeAllHeartbeatMessages()
	m.computeInactiveHeartbeatMessages()
}
