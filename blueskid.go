package main

import (
	blueskidgo "blueskidgo/lib"
	"flag"
	"fmt"
	"log"
	"net/http"
)

// contains a web server which exhibits blueskid exercising parts of the "@bluesky identity" protocol.

func main() {
	port := flag.Int("port", 8123, "port number")
	portArg := fmt.Sprintf(":%d", *port)

	http.HandleFunc("/grant-assertions", blueskidgo.GrantAssertionsHandler)
	http.HandleFunc("/claim-assertion", blueskidgo.ClaimAssertionsHandler)
	http.HandleFunc("/unclaim-assertion", blueskidgo.UnclaimAssertionsHandler)
	http.HandleFunc("/claim-bid", blueskidgo.ClaimBIDHandler)
	http.HandleFunc("/grant-bid", blueskidgo.GrantBIDHandler)
	http.HandleFunc("/unclaim-bid", blueskidgo.UnclaimBIDHandler)
	http.HandleFunc("/pid-group", blueskidgo.GetPIDGroupHandler)
	http.HandleFunc("/pids-for-bid", blueskidgo.GetPIDsForBIDHandler)
	http.HandleFunc("/bids-for-pid", blueskidgo.GetBIDsforPIDHandler)
	http.HandleFunc("/ledger", blueskidgo.LedgerHandler)

	err := http.ListenAndServe(portArg, nil)
	if err != nil {
		log.Fatalln(err)
	}
}
