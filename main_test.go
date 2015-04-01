package main

import (
  r "github.com/dancannon/gorethink"
  "fmt"
  "log"
  "net/http"
  "net/http/httptest"
  "github.com/stretchr/testify/assert"
  "testing"
)

func TestUsersIndex (t *testing.T) {
  recorder := httptest.NewRecorder()
  req, err := http.NewRequest("GET", "/api/v1/users", nil)
  assert.Nil(t, err)
  IndexUsersHandler(recorder, req)
  assert.Equal(t, 200, recorder.Code)
  assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
}

func TestUsersIndexXXX (t *testing.T) {
  server := httptest.NewServer(http.HandlerFunc(IndexUsersHandler))
  defer server.Close()

  // Pretend this is some sort of Go client...
  url := fmt.Sprintf("%s?say=Nothing", server.URL)

  resp, err := http.DefaultClient.Get(url)
  assert.Nil(t, err)
  assert.Equal(t, 200, resp.StatusCode)
}

func InitTestDB() *r.Session {
  session, err := r.Connect(r.ConnectOpts{
    Address: "localhost:28015",
    Database: "gadder_test",
  })

  if err != nil {
    log.Println(err)
  }
  r.DbDrop("gadder_test").Exec(session)
  err = r.DbCreate("gadder_test").Exec(session)
  if err != nil {
    log.Println(err)
  }

  return session
}
