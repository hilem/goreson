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
    "pong": "true",
    "version": CurrVersion,
    "released_on": "31/03/2015",
    "recent_changes": `- Added Participants table`,
  }, w)
}

//// OLD
// func activeIndexHandler(w http.ResponseWriter, req *http.Request) {
//   items := []TodoItem{}
//
//   // Fetch all the items from the database
//   query := r.Table("items").Filter(r.Row.Field("Status").Eq("active"))
//   query = query.OrderBy(r.Asc("Created"))
//   res, err := query.Run(session)
//   if err != nil {
//     http.Error(w, err.Error(), http.StatusInternalServerError)
//     return
//   }
//   err = res.All(&items)
//   if err != nil {
//     http.Error(w, err.Error(), http.StatusInternalServerError)
//     return
//   }
//
//   renderTemplate(w, "index", items)
// }
//
// func completedIndexHandler(w http.ResponseWriter, req *http.Request) {
//   items := []TodoItem{}
//
//   // Fetch all the items from the database
//   query := r.Table("items").Filter(r.Row.Field("Status").Eq("complete"))
//   query = query.OrderBy(r.Asc("Created"))
//   res, err := query.Run(session)
//   if err != nil {
//     http.Error(w, err.Error(), http.StatusInternalServerError)
//     return
//   }
//   err = res.All(&items)
//   if err != nil {
//     http.Error(w, err.Error(), http.StatusInternalServerError)
//     return
//   }
//
//   renderTemplate(w, "index", items)
// }
//
// func toggleHandler(w http.ResponseWriter, req *http.Request) {
//   // vars := mux.Vars(req)
//   // id := vars["id"]
//   id := req.URL.Query().Get(":id")
//   if id == "" {
//     http.NotFound(w, req)
//     return
//   }
//
//   // Check that the item exists
//   res, err := r.Table("items").Get(id).Run(session)
//   if err != nil {
//     http.Error(w, err.Error(), http.StatusInternalServerError)
//     return
//   }
//
//   if res.IsNil() {
//     http.NotFound(w, req)
//     return
//   }
//
//   // Toggle the item
//   _, err = r.Table("items").Get(id).Update(map[string]interface{}{"Status": r.Branch(
//     r.Row.Field("Status").Eq("active"),
//     "complete",
//     "active",
//   )}).RunWrite(session)
//   if err != nil {
//     http.Error(w, err.Error(), http.StatusInternalServerError)
//     return
//   }
//
//   http.Redirect(w, req, "/", http.StatusFound)
// }
//
// func deleteHandler(w http.ResponseWriter, req *http.Request) {
//   // vars := mux.Vars(req)
//   // id := vars["id"]
//   id := req.URL.Query().Get(":id")
//   if id == "" {
//     http.NotFound(w, req)
//     return
//   }
//
//   // Check that the item exists
//   res, err := r.Table("items").Get(id).Run(session)
//   if err != nil {
//     http.Error(w, err.Error(), http.StatusInternalServerError)
//     return
//   }
//
//   if res.IsNil() {
//     http.NotFound(w, req)
//     return
//   }
//
//   // Delete the item
//   _, err = r.Table("items").Get(id).Delete().RunWrite(session)
//   if err != nil {
//     http.Error(w, err.Error(), http.StatusInternalServerError)
//     return
//   }
//
//   http.Redirect(w, req, "/", http.StatusFound)
// }
//
// func clearHandler(w http.ResponseWriter, req *http.Request) {
//   // Delete all completed items
//   _, err := r.Table("items").Filter(
//     r.Row.Field("Status").Eq("complete"),
//   ).Delete().RunWrite(session)
//   if err != nil {
//     http.Error(w, err.Error(), http.StatusInternalServerError)
//     return
//   }
//
//   http.Redirect(w, req, "/", http.StatusFound)
// }
