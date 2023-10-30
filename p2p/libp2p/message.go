package libp2p

import (
	"fmt"

	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/p2p"
	"github.com/Dharitri-org/sme-dharitri/p2p/data"
	"github.com/Dharitri-org/sme-dharitri/p2p/message"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// NewMessage returns a new instance of a Message object
func NewMessage(msg *pubsub.Message, marshalizer p2p.Marshalizer) (*message.Message, error) {
	if check.IfNil(marshalizer) {
		return nil, p2p.ErrNilMarshalizer
	}

	newMsg := &message.Message{
		FromField:      msg.From,
		PayloadField:   msg.Data,
		SeqNoField:     msg.Seqno,
		TopicsField:    msg.TopicIDs,
		SignatureField: msg.Signature,
		KeyField:       msg.Key,
	}

	topicMessage := &data.TopicMessage{}
	err := marshalizer.Unmarshal(topicMessage, msg.Data)
	if err != nil {
		return nil, fmt.Errorf("%w error: %s", p2p.ErrMessageUnmarshalError, err.Error())
	}

	newMsg.DataField = topicMessage.Payload
	newMsg.TimestampField = topicMessage.Timestamp

	id, err := peer.IDFromBytes(newMsg.From())
	if err != nil {
		return nil, err
	}

	newMsg.PeerField = core.PeerID(id)
	return newMsg, nil
}
