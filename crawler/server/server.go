package server

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/phoreproject/obp-search-engine/crawler/crawling"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

type CrawlServer struct {
	crawlerInstance *crawling.Crawler
	serverPort      int
}

func NewCrawlServer(crawlerInstance *crawling.Crawler, serverPort int) *CrawlServer {
	return &CrawlServer{crawlerInstance: crawlerInstance,
		serverPort: serverPort}
}

func (c CrawlServer) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Health check."))
}

func (c CrawlServer) CrawlHandler(w http.ResponseWriter, r *http.Request) {
	nodeId := mux.Vars(r)["nodeID"]
	c.crawlerInstance.ProcessOneNodeSync(nodeId)
	val := fmt.Sprintf("Crawling node %s.", nodeId)
	_, err := w.Write([]byte(val))
	if err != nil {
		fmt.Printf("Failed to write response.")
	}
}

func (c CrawlServer) Serve() {
	log.Info("Starting crawling server.")
	rtr := mux.NewRouter()
	rtr.HandleFunc("/crawl/{nodeID:[a-zA-Z0-9]{46}}", c.CrawlHandler)
	rtr.HandleFunc("/", c.HealthCheck)

	http.Handle("/", rtr)
	err := http.ListenAndServe(strconv.Itoa(c.serverPort), nil)
	if err != nil {
		log.Error("Server serving error")
		log.Error(err)
	}
}
