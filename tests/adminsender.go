package tests

import (
	"sync"

	"github.com/moira-alert/notifier"
)

type adminSender struct {
	mutex      sync.Mutex
	lastEvents notifier.EventsData
}

func (sender *adminSender) Init(senderSettings map[string]string, logger notifier.Logger) error {
	return nil
}

func (sender *adminSender) SendEvents(events notifier.EventsData, contact notifier.ContactData, trigger notifier.TriggerData, throttled bool) error {
	sender.mutex.Lock()
	sender.lastEvents = events
	sender.mutex.Unlock()
	return nil
}
