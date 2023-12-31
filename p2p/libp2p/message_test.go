package libp2p_test

import (
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"testing"
	"time"

	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/p2p"
	"github.com/Dharitri-org/sme-dharitri/p2p/data"
	"github.com/Dharitri-org/sme-dharitri/p2p/libp2p"
	"github.com/Dharitri-org/sme-dharitri/testscommon"
	"github.com/btcsuite/btcd/btcec"
	libp2pCrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pubsubpb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/stretchr/testify/assert"
)

func getRandomID() []byte {
	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	sk := (*libp2pCrypto.Secp256k1PrivateKey)(prvKey)
	id, _ := peer.IDFromPublicKey(sk.GetPublic())

	return []byte(id)
}

func TestMessage_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	pMes := &pubsub.Message{}
	m, err := libp2p.NewMessage(pMes, nil)

	assert.True(t, check.IfNil(m))
	assert.True(t, errors.Is(err, p2p.ErrNilMarshalizer))
}

func TestMessage_ShouldErrBecauseOfFromField(t *testing.T) {
	t.Parallel()

	from := []byte("dummy from")
	marshalizer := &testscommon.ProtoMarshalizerMock{}

	topicMessage := &data.TopicMessage{
		Version:   libp2p.CurrentTopicMessageVersion,
		Timestamp: time.Now().Unix(),
		Payload:   []byte("data"),
	}
	buff, _ := marshalizer.Marshal(topicMessage)
	mes := &pubsubpb.Message{
		From: from,
		Data: buff,
	}
	pMes := &pubsub.Message{Message: mes}
	m, err := libp2p.NewMessage(pMes, marshalizer)

	assert.True(t, check.IfNil(m))
	assert.NotNil(t, err)
}

func TestMessage_ShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &testscommon.ProtoMarshalizerMock{}
	topicMessage := &data.TopicMessage{
		Version:   libp2p.CurrentTopicMessageVersion,
		Timestamp: time.Now().Unix(),
		Payload:   []byte("data"),
	}
	buff, _ := marshalizer.Marshal(topicMessage)
	mes := &pubsubpb.Message{
		From: getRandomID(),
		Data: buff,
	}

	pMes := &pubsub.Message{Message: mes}
	m, err := libp2p.NewMessage(pMes, marshalizer)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(m))
}

func TestMessage_From(t *testing.T) {
	t.Parallel()

	from := getRandomID()
	marshalizer := &testscommon.ProtoMarshalizerMock{}
	topicMessage := &data.TopicMessage{
		Version:   libp2p.CurrentTopicMessageVersion,
		Timestamp: time.Now().Unix(),
		Payload:   []byte("data"),
	}
	buff, _ := marshalizer.Marshal(topicMessage)
	mes := &pubsubpb.Message{
		From: from,
		Data: buff,
	}
	pMes := &pubsub.Message{Message: mes}
	m, err := libp2p.NewMessage(pMes, marshalizer)

	assert.Nil(t, err)
	assert.Equal(t, m.From(), from)
}

func TestMessage_Peer(t *testing.T) {
	t.Parallel()

	id := getRandomID()
	marshalizer := &testscommon.ProtoMarshalizerMock{}

	topicMessage := &data.TopicMessage{
		Version:   libp2p.CurrentTopicMessageVersion,
		Timestamp: time.Now().Unix(),
		Payload:   []byte("data"),
	}
	buff, _ := marshalizer.Marshal(topicMessage)
	mes := &pubsubpb.Message{
		From: id,
		Data: buff,
	}
	pMes := &pubsub.Message{Message: mes}
	m, _ := libp2p.NewMessage(pMes, marshalizer)

	assert.Equal(t, core.PeerID(id), m.Peer())
}

func TestMessage_WrongVersionShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &testscommon.ProtoMarshalizerMock{}

	topicMessage := &data.TopicMessage{
		Version:   libp2p.CurrentTopicMessageVersion + 1,
		Timestamp: time.Now().Unix(),
		Payload:   []byte("data"),
	}
	buff, _ := marshalizer.Marshal(topicMessage)
	mes := &pubsubpb.Message{
		From: getRandomID(),
		Data: buff,
	}

	pMes := &pubsub.Message{Message: mes}
	m, err := libp2p.NewMessage(pMes, marshalizer)

	assert.True(t, check.IfNil(m))
	assert.True(t, errors.Is(err, p2p.ErrUnsupportedMessageVersion))
}

func TestMessage_PopulatedPkFieldShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &testscommon.ProtoMarshalizerMock{}

	topicMessage := &data.TopicMessage{
		Version:   libp2p.CurrentTopicMessageVersion,
		Timestamp: time.Now().Unix(),
		Payload:   []byte("data"),
		Pk:        []byte("p"),
	}
	buff, _ := marshalizer.Marshal(topicMessage)
	mes := &pubsubpb.Message{
		From: getRandomID(),
		Data: buff,
	}

	pMes := &pubsub.Message{Message: mes}
	m, err := libp2p.NewMessage(pMes, marshalizer)

	assert.True(t, check.IfNil(m))
	assert.True(t, errors.Is(err, p2p.ErrUnsupportedFields))
}

func TestMessage_PopulatedSigFieldShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &testscommon.ProtoMarshalizerMock{}

	topicMessage := &data.TopicMessage{
		Version:        libp2p.CurrentTopicMessageVersion,
		Timestamp:      time.Now().Unix(),
		Payload:        []byte("data"),
		SignatureOnPid: []byte("s"),
	}
	buff, _ := marshalizer.Marshal(topicMessage)
	mes := &pubsubpb.Message{
		From: getRandomID(),
		Data: buff,
	}

	pMes := &pubsub.Message{Message: mes}
	m, err := libp2p.NewMessage(pMes, marshalizer)

	assert.True(t, check.IfNil(m))
	assert.True(t, errors.Is(err, p2p.ErrUnsupportedFields))
}
