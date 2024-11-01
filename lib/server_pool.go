package lib

import (
	"container/heap"
	"log"
	"net/url"
	"sync"
)

type ServerNodeQueue []*ServerNode

func (pq ServerNodeQueue) Len() int {
	return len(pq)
}

func (pq ServerNodeQueue) Less(i, j int) bool {
	return (pq[i].Weight < pq[j].Weight)
}

func (pq ServerNodeQueue) Swap(i, j int) { pq[i], pq[j] = pq[j], pq[i] }

func (pq *ServerNodeQueue) Push(x interface{}) {
	*pq = append(*pq, x.(*ServerNode))
}

func (pq *ServerNodeQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

type ServerPool struct {
	Backends ServerNodeQueue
	mutex    sync.Mutex
}

func (s *ServerPool) GetNextPeer() *ServerNode {
	log.Println("Getting next peer.. !")
	s.mutex.Lock()
	defer s.mutex.Unlock()
	var server *ServerNode
	for s.Backends.Len() > 0 {
		server = heap.Pop(&s.Backends).(*ServerNode)
		if server.Alive {
			server.Mux.Lock()
			server.Weight++
			server.Mux.Unlock()
			heap.Push(&s.Backends, server)
			return server
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
