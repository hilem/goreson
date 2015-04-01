package main

import (
  r "github.com/dancannon/gorethink"
  "github.com/dancannon/gorethink/encoding"
	"html/template"
	"log"
	"net/http"
  "encoding/json"
	"os"
	"path/filepath"
  "time"
  "strconv"
)

var templates *template.Template

func init() {
	filenames := []string{}
	err := filepath.Walk("templates", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(path) == ".gohtml" {
			filenames = append(filenames, path)
		}

		return nil
	})
	if err != nil {
		log.Fatalln(err)
	}

	if len(filenames) == 0 {
		return
	}

	templates, err = template.ParseFiles(filenames...)
	if err != nil {
		log.Fatalln(err)
	}
}

func renderTemplate(w http.ResponseWriter, tmpl string, vars interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+".gohtml", vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func sendJson(v interface{}, w http.ResponseWriter) {
  js, err := json.Marshal(v)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  w.Header().Set("Content-Type", "application/json")
  w.Write(js)
}

func readIntFromUrlParam(param string, val *int, w http.ResponseWriter, req *http.Request) bool {
  if len([]rune(param)) > 0 {
    tmpVal, err := strconv.Atoi(param)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return false
    }
    *val = tmpVal
    return true
  }
  return true
}

func readBody(p interface{}, w http.ResponseWriter, req *http.Request) bool {
  decoder := json.NewDecoder(req.Body)
  err := decoder.Decode(&p)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return false
  }
  log.Printf("Params:%+v\n", p)
  return true
}

func fetchSessionByUser(user_id string) (s UserSession, err error) {
  res, err := r.Table("sessions").GetAllByIndex("user_id", user_id).Run(session)
  if err != nil {
    return s, err
  }

  if res.IsNil() {
    log.Println("fetchSession: if res isNil")
    t := time.Now()
    tmp_s := UserSession{
      UserId: user_id,
      CreatedAt: t,
      UpdatedAt: t,
    }
    res2, _ := r.Table("sessions").Insert(tmp_s, r.InsertOpts{ReturnChanges: true}).RunWrite(session)
    encoding.Decode(&s, res2.Changes[0].NewValue) // using reflection
  } else {
    log.Println("fetchSession: else res is not nil")
    res.One(&s)
  }
  return s, err
}

func fetchUserFromSession(u *User, sid string, id string, w http.ResponseWriter, req *http.Request) bool {
  var userSession UserSession
  var sessionUser User

  // Ensure Request was passed a session ID
  if len([]rune(sid)) == 0 {
    http.Error(w, "Missing Session ID", http.StatusBadRequest)
    return false
  }

  // Lookup Session in DB
  res, err := r.Table("sessions").Get(sid).Run(session)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return false
  }
  if res.IsNil() {
    http.Error(w, "Couldn't find Session with ID: " + sid, http.StatusNotFound)
    return false
  }
  res.One(&userSession)

  // Lookup User in DB from Session
	res, err = r.Table("users").Get(userSession.UserId).Run(session)
	if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return false
  }
  if res.IsNil() {
    http.Error(w, "Couldn't find User matching Session", http.StatusNotFound)
    return false
	}
  res.One(&sessionUser)

  // if no id is passed just return the user in the session
  if(len(id) == 0) {
    *u = sessionUser
    return true
  }

  // Find User from ID passed in
  if ok := findUser(id, u, w, req); !ok {
    return false
  }

  // if User ID passed does not equal User ID in Session
  log.Printf("Session  User.Id: %+v\n", sessionUser.Id)
  log.Printf("Resource User.Id: %+v\n", u.Id)
  if sessionUser.Id != u.Id {
    http.Error(w, "Users are not equal", http.StatusBadRequest)
    return false
  }

  return true
}
