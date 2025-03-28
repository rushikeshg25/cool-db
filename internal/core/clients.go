package core

import (
	"sync"
	"sync/atomic"
)

type clientManager struct {
	connectedClients sync.Map
	numOfClients     atomic.Int32
}

func newClientManager() *clientManager {
	return &clientManager{}
}

func (cm *clientManager) registerClient(clientID int, client interface{}) {
	cm.connectedClients.Store(clientID, client)
	cm.numOfClients.Add(1)
}

func (cm *clientManager) unregisterClient(clientID int) {
	cm.connectedClients.Delete(clientID)
	cm.numOfClients.Add(-1)
}
