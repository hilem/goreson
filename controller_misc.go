package main

import (
	// r "github.com/dancannon/gorethink"
	"log"
	"net/http"
)

// StatusHandler is used to quickly test if the server is up and responding
// Example:
//   Request:
//     curl -X GET localhost:3000/ping
//   Response:
//     {
//         "pong": true
//     }
func StatusHandler(w http.ResponseWriter, req *http.Request) {
	log.Println("PING STATUS HANDLER")

	sendJson(map[string]interface{}{
		"pong":           "true",
		"version":        CurrVersion,
		"released_on":    "05/04/2015",
		"recent_changes": `- [bug] Participants returning Joined User/Event wasn't working properly`,
	}, w)
}
