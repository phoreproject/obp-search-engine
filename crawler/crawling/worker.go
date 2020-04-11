package crawling

import (
	"log"
	"time"
)

// Worker represents a single crawling worker.
type Worker struct {
	workerQueue  chan chan Node
	datastore    Datastore
	rpcInterface RPCInterface
	workQueue    chan Node
}

// NewWorker creates a new worker given certain interfaces.
func NewWorker(workerQueue chan chan Node, datastore Datastore, rpc RPCInterface) Worker {
	return Worker{workerQueue: workerQueue, datastore: datastore, workQueue: make(chan Node), rpcInterface: rpc}
}

// Work continuously receives work from the queue and updates it.
func (w *Worker) Work() {
	for {
		w.workerQueue <- w.workQueue
		work := <-w.workQueue

		work.LastCrawled = time.Now()

		connections, err := w.rpcInterface.GetConnections(work.ID)

		if err != nil {
			log.Fatal(err)
		}

		nodes := make([]Node, len(connections))
		for i := range connections {
			nodes[i] = Node{ID: connections[i], Connections: []string{}, LastCrawled: time.Date(2017, 12, 13, 0, 0, 0, 0, time.Local)}
		}

		work.Connections = connections

		w.datastore.AddUninitializedNodes(nodes)

		w.datastore.SaveNode(work)
	}
}
