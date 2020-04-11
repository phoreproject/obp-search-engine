package main

import (
	"database/sql"
	"flag"
	_ "github.com/go-sql-driver/mysql"
	"github.com/phoreproject/obp-search-engine/crawler/crawling"
	"github.com/phoreproject/obp-search-engine/crawler/db"
	"github.com/phoreproject/obp-search-engine/crawler/rpc"
	"github.com/phoreproject/obp-search-engine/crawler/server"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
)

func main() {
	// configure logger output format
	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true
	log.SetFormatter(customFormatter)

	// configure logger writers
	logFile, err := os.Create("crawler.log")
	if err != nil {
		log.Panic(err)
	}
	log.SetOutput(io.MultiWriter(os.Stderr, logFile)) // write to file and stderr

	// url format is user:password@protocol(address:port)/db_name
	databaseURL := flag.String("mysql", "root@tcp(127.0.0.1:3306)/obpsearch", "database url used to connect to MySQL database")
	rpcURL := flag.String("rpc", "127.0.0.1:5002", "rpc url used to connect to Phore Marketplace")
	skipMigration := flag.Bool("skipMigration", false, "skip database migration to the newest version on start")
	verbose := flag.Bool("verbose", false, "use more verbose logging")
	chunkSize := flag.Int("chunkSize", 100, "Maximum database select chunk size")
	maxParallelCorutine := flag.Int("maxCoroutine", 10, "Maximum number of parallel connections")
	httpClassifierUrl := flag.String("httpClassifierUrl", "", "Service url for classifying listings")
	startCrawlOnDemandServer := flag.Bool("server", false, "Start http server with on demand crawl API")
	flag.Parse()

	if *verbose {
		log.SetLevel(log.DebugLevel)
		log.Info("Using verbose logging!")
	}

	log.Debugf("Starting app with chunk size %d, and max parallel corutine cnt %d", *chunkSize, *maxParallelCorutine)

	database, err := sql.Open("mysql", *databaseURL+"?parseTime=true&interpolateParams=true")
	if err != nil {
		log.Panic(err)
	}

	d, err := db.NewSQLDatastore(database, !(*skipMigration))
	if err != nil {
		log.Panic(err)
	}

	r := rpc.NewRPC(*rpcURL)

	crawler := &crawling.Crawler{RPCInterface: r, DB: d, MaxCoroutineCnt: *maxParallelCorutine, ChunkSize: *chunkSize,
		HttpClassifierUrl: *httpClassifierUrl}

	config, err := r.GetConfig()
	if err != nil {
		log.Panic("You need to run marketplaced. Please check: https://github.com/phoreproject/pm-go")
	}

	profile, err := r.GetProfile(config.PeerID)
	if err != nil {
		log.Panic(err)
	}

	userAgent, err := crawler.RPCInterface.GetUserAgentFromIPNS(config.PeerID)
	if err != nil {
		log.Panic(err)
	}

	// add ourselves
	err = crawler.DB.SaveNodeUninitialized(crawling.Node{ID: config.PeerID, UserAgent: userAgent, Connections: []string{}, Profile: profile})
	if err != nil {
		log.Panic(err)
	}

	if *startCrawlOnDemandServer {
		serverInstance := server.NewCrawlServer(crawler)
		serverInstance.Serve()
	}

	crawler.MainLoop()
}
