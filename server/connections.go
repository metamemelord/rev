package server

import (
	"context"
	"net"
	"sync"

	"github.com/PlanckProject/go-commons/logger"
)

type ServiceConnection struct {
	Ctx        context.Context
	CancelFunc context.CancelFunc
	Conn       net.TCPConn
}

var connection_store map[string]*ServiceConnection = make(map[string]*ServiceConnection)
var rwLock sync.RWMutex

func AddConnection(serviceConnKey string, serviceConn *ServiceConnection) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
		}
	}()
	logger.Info("Adding a new connection: ", serviceConnKey)
	rwLock.Lock()
	defer rwLock.Unlock()
	if existingConn, ok := connection_store[serviceConnKey]; ok {
		logger.Warnf("Re-registration tried for %s. Closing old connection and replacing with new one.", serviceConnKey)
		existingConn.CancelFunc()
		existingConn.Conn.Close()
	}
	connection_store[serviceConnKey] = serviceConn
}

func GetConnection(serviceConnKey string) *ServiceConnection {
	rwLock.RLock()
	defer rwLock.RUnlock()
	return connection_store[serviceConnKey]
}

func ShutdownConnectionStore() {
	rwLock.Lock()
	defer rwLock.Unlock()
	for _, conn := range connection_store {
		conn.CancelFunc()
		conn.Conn.Close()
	}
}
