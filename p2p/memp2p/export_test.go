package memp2p

import "github.com/Dharitri-org/sme-dharitri/p2p"

func (messenger *Messenger) TopicValidator(name string) p2p.MessageProcessor {
	messenger.topicsMutex.RLock()
	processor := messenger.topicValidators[name]
	messenger.topicsMutex.RUnlock()

	return processor
}
