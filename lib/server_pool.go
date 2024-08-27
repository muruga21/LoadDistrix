package lib

import (
	"log"
	"math"
	"net/url"
	"sync/atomic"
)

type ServerPool struct {
	Backends []*ServerNode
	Current  uint64
}

func (s *ServerPool) NextServerIndex() int {
	nxtIndex := atomic.AddUint64(&s.Current, uint64(1)) % uint64(len(s.Backends))
	if s.Current >= math.MaxUint64-1 {
		atomic.StoreUint64(&s.Current, 0)
	}
	return int(nxtIndex)
}

func (s *ServerPool) GetNextPeer() *ServerNode {
	log.Println("Getting next peer")
	nxtIndex := s.NextServerIndex()
	lenOfBackendArr := len(s.Backends)
	lengthNeedToTraverse := lenOfBackendArr + nxtIndex

	for i := nxtIndex; i < lengthNeedToTraverse; i++ {
		index := i % lenOfBackendArr
		log.Printf("Checking server at index %d, Alive: %v\n", index, s.Backends[index].IsAlive())
		if s.Backends[index].IsAlive() {
			atomic.StoreUint64(&s.Current, uint64(index))
			return s.Backends[index]
		}
	}

	return nil
}

func (s *ServerPool) MarkDownTheServer(backendUrl *url.URL, serverStatus bool) {
	for _, backend := range s.Backends {
		if backend.URL.String() == backendUrl.String() {
			backend.SetAlive(serverStatus)
			return
		}
	}
}
