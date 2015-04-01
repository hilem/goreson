// Routes
//    ** MISC **
//    GET     /ping                              StatusHandler
//
//    ** USERS **
//    GET     /api/:v/users                      IndexUsersHandler
//    POST    /api/:v/users                      CreateUserHandler
//    GET     /api/:v/users/:id                  ShowUserHandler
//    PUT     /api/:v/users/:id                  UpdateUserHandler
//    DELETE  /api/:v/users/:id                  DeleteUserHandler
//
//    ** EVENTS **
//    GET     /api/:v/users/:user_id/events      IndexUserEventsHandler
//    POST    /api/:v/users/:user_id/events      CreateUserEventHandler
//    PUT     /api/:v/users/:user_id/events/:id  UpdateUserEventHandler
//    DELETE  /api/:v/users/:user_id/events/:id  DeleteUserEventHandler
//    GET     /api/:v/events/:id                 ShowEventHandler
//    GET     /api/:v/events                     IndexEventsHandler
//
//    ** MESSAGES **
//    POST    /api/:v/events/:event_id/messages  CreateEventMessageHandler
//    GET     /api/:v/events/:event_id/messages  IndexEventMessagesHandler
//    DELETE  /api/:v/messages/:id               DeleteMessageHandler
//    PUT     /api/:v/messages/:id               UpdateMessageHandler
//    GET     /api/:v/messages                   IndexUserMessagesHandler
//
//     ** PARTICIPANTS **
//     POST   /api/:v/events/:event_id/participants  CreateEventParticipantHandler
//     GET    /api/:v/events/:event_id/participants  IndexEventParticipantsHandler
//     DELETE /api/:v/participants/:id               DeleteParticipantHandler
//     PUT    /api/:v/participants/:id               UpdateParticipantHandler
//     GET    /api/:v/participants                   IndexUserParticipantsHandler


package main

import (
  r "github.com/dancannon/gorethink"
  "github.com/bmizerany/pat"
  "log"
  "net/http"
)

const TimeFormat string = "Mon Jan 2 2006 15:04:05 MST-07:00"
const CurrVersion string = "v0.6.0"

var (
  router  *pat.PatternServeMux
  session *r.Session
)

func init() {
  var err error

  log.Println("Starting up")
  session, err = r.Connect(r.ConnectOpts{
    Address:  "db:28015",
    Database: "gadder",
    // AuthKey:  "6mg<MrNRz}M6n24$~1N:zLWt!F1]43",
  })
  if err != nil {
    log.Fatalln(err.Error())
  }
}

func NewServer(addr string) *http.Server {
  // Setup router
  router = initRouting()
  http.Handle("/", router)
  http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

  // Create and start server
  return &http.Server{
    Addr:    addr,
  }
}

func StartServer(server *http.Server) {
  err := server.ListenAndServe()
  if err != nil {
    log.Fatalln("Error: %v", err)
  }
}

// Notes: A trailing slash on index route will grab both index and show routes
func initRouting() *pat.PatternServeMux {
  // TODO: once we add better session authentication we can remove the user from the routes and use that
  m := pat.New()
  //// Misc
  m.Get("/ping",              http.HandlerFunc(StatusHandler))
  //// API
  // Authentication
  // m.Post("/api/:v/device",   http.HandlerFunc(createDeviceHandler))
  // m.Del("/api/:v/device",    http.HandlerFunc(deleteDeviceHandler))
  // Users
  m.Get("/api/:v/users",      http.HandlerFunc(IndexUsersHandler))
  m.Post("/api/:v/users",     http.HandlerFunc(CreateUserHandler))
  m.Get("/api/:v/users/:id",  http.HandlerFunc(ShowUserHandler))
  m.Put("/api/:v/users/:id",  http.HandlerFunc(UpdateUserHandler))
  m.Del("/api/:v/users/:id",  http.HandlerFunc(DeleteUserHandler))
  // Events
  m.Post("/api/:v/users/:user_id/events",    http.HandlerFunc(CreateUserEventHandler))
  m.Put("/api/:v/users/:user_id/events/:id", http.HandlerFunc(UpdateUserEventHandler))
  m.Del("/api/:v/users/:user_id/events/:id", http.HandlerFunc(DeleteUserEventHandler))
  m.Get("/api/:v/users/:user_id/events",     http.HandlerFunc(IndexUserEventsHandler))
  m.Get("/api/:v/events/:id", http.HandlerFunc(ShowEventHandler))
  m.Get("/api/:v/events",     http.HandlerFunc(IndexEventsHandler))
  // Messages
  m.Post("/api/:v/events/:event_id/messages", http.HandlerFunc(CreateEventMessageHandler))
  m.Get("/api/:v/events/:event_id/messages",  http.HandlerFunc(IndexEventMessagesHandler))
  m.Del("/api/:v/messages/:id",               http.HandlerFunc(DeleteMessageHandler))
  m.Put("/api/:v/messages/:id",               http.HandlerFunc(UpdateMessageHandler))
  m.Get("/api/:v/messages",                   http.HandlerFunc(IndexUserMessagesHandler))
  // Participants
  m.Post("/api/:v/events/:event_id/participants", http.HandlerFunc(CreateEventParticipantHandler))
  m.Get("/api/:v/events/:event_id/participants",  http.HandlerFunc(IndexEventParticipantsHandler))
  m.Del("/api/:v/participants/:id",               http.HandlerFunc(DeleteParticipantHandler))
  m.Put("/api/:v/participants/:id",               http.HandlerFunc(UpdateParticipantHandler))
  m.Get("/api/:v/participants",                   http.HandlerFunc(IndexUserParticipantsHandler))
  //
  log.Println("Creating Routes")
  return m
}
