package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/phoreproject/obp-search-engine/crawling"
	"github.com/phoreproject/obp-search-engine/db"
	"github.com/phoreproject/obp-search-engine/rpc"
	"time"
)

func testInterface() {
	// try to connect with marketplace with api
	rpcURL := flag.String("rpc", "127.0.0.1:5002", "rpc url used to connect to Phore Marketplace")
	r := rpc.NewRPC(*rpcURL)

	fmt.Print("Get config")
	config, err := r.GetConfig()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", config)

	peerId := config.PeerID

	fmt.Println("Get connections")
	connections, err := r.GetConnections(peerId)
	if err != nil {
		panic(err)
	}

	fmt.Println("Printing connections")
	for index := range connections {
		fmt.Printf("%+v\n", connections[index])
	}

	fmt.Println("Get items")
	items, err := r.GetItems(peerId)
	fmt.Println("Printing items")
	for index := range  items {
		fmt.Printf("%+v\n", items[index])
	}

	fmt.Println("Get profile")
	profile, err := r.GetProfile(peerId)
	fmt.Println("Print profile")
	fmt.Printf("%+v\n", profile)

	fmt.Println("Get User Agent")
	userAgent, err := r.GetUserAgent(peerId)
	fmt.Println("Print user agent")
	fmt.Printf("%+v\n", userAgent)
}

func testDB() {
	// try to connect marketplace and push results into db
	rpcURL := flag.String("rpc", "127.0.0.1:5002", "rpc url used to connect to Phore Marketplace")
	databaseURL := flag.String("mysql", "root:secret@tcp(127.0.0.1:3306)/obpsearch", "database url used to connect to MySQL database")
	database, err := sql.Open("mysql", *databaseURL+"?parseTime=true&interpolateParams=true")
	if err != nil {
		panic(err)
	}
	sqlDataStore, err := db.NewSQLDatastore(database)
	if err != nil {
		panic(err)
	}

	r := rpc.NewRPC(*rpcURL)
	config, err := r.GetConfig()
	if err != nil {
		panic(err)
	}

	peerId := config.PeerID
	connections, err := r.GetConnections(peerId)
	profile, err := r.GetProfile(peerId)
	userAgent, err := r.GetUserAgent(peerId)

	node := crawling.Node{peerId, userAgent, connections, time.Now(), profile}
	err = sqlDataStore.SaveNode(node)
	if err != nil {
		panic(err)
	}

	items, err := r.GetItems(peerId)
	if err != nil {
		panic(err)
	}

	if err = sqlDataStore.AddItemsForNode(peerId, items); err != nil {
		panic(err)
	}
}

func main() {
	testDB()
}