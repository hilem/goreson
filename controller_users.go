package main

import (
  r "github.com/dancannon/gorethink"
  "github.com/dancannon/gorethink/encoding"
  "net/http"
  "fmt"
  "log"
  "strings"
  "time"
  "strconv"
)

// Name/Desc: IndexUsersHandler - returns paginated list of users
//
// Optional URL Params: page=<Integer> && per=<Integer>
//
// Example:
//   Request:
//     curl -X GET <HOST_DOMAIN:PORT>/api/v1/users
//   Response:
//   {
//     "page": "1",
//     "per": "20",
//     "users": [
//       {
//         "id": "2d6a3836-5535-4f8e-8e77-7eff45561984",
//         "first_name": "bobby",
//         "last_name": "",
//         "email": "",
//         "avatar": "",
//         "bio": "",
//         "created_at": "2014-11-10T18:19:58Z",
//         "updated_at": "2014-11-10T21:02:21Z"
//       },
//       {
//         "id": "3619f498-bcaf-44e5-9b80-02a07566f177",
//         "first_name": "",
//         "last_name": "",
//         "email": "",
//         "avatar": "",
//         "bio": "Mercury!!!!",
//         "created_at": "0001-01-01T00:00:00Z",
//         "updated_at": "0001-01-01T00:00:00Z"
//       },
//       ...
//     ]
//   }
func IndexUsersHandler(w http.ResponseWriter, req *http.Request) {
  fmt.Println("")
  log.Println("Listing Users...")
  users := []User{}

  //// Pagination
  page := 1
  per  := 20
  if ok := readIntFromUrlParam(req.URL.Query().Get("page"), &page, w, req); !ok {
    return
  }
  if ok := readIntFromUrlParam(req.URL.Query().Get("per"), &per, w, req); !ok {
    return
  }
  log.Printf("(after) page = %+v | per = %+v\n", page, per)

  res, err := r.Table("users").OrderBy(r.Asc("CreatedAt")).Slice((page-1)*per, page*per).Run(session)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  err = res.All(&users)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  sendJson(map[string]interface{}{
    "users": users,
    "page": strconv.Itoa(page),
    "per": strconv.Itoa(per),
  }, w)
}

// Name/Desc: CreateUserHandler - creates a new user
//
// Required:  user hash & facebook_id
//
// Returns:   user object
//            session_id (sid)
//
// Notes: If user w/ facebook_id already exists, that user is returned instead of creating
//
// Example:
//  Request:
//     curl -X POST 
//          -DATA '{"user": {                           \
//                    "bio": "blah blah blah",          \
//                    "i_am_not_getting_saved": "test", \
//                    "first_name": "Reggie",           \
//                    "last_name": "Bush"               \
//                  },                                  \
//                  "facebook_id": "1jj32-sdfs2-sds" }'
//          <HOST_DOMAIN:PORT>/api/v1/users
//  Response:
//     {
//       "user": {
//         "id": "c8a668fe-2574-47a6-a61e-dcd4a35fde54",
//         "first_name": "Reggie",
//         "last_name": "Bush",
//         "email": "",
//         "avatar": "",
//         "bio": "blah blah blah",
//         "created_at": "2014-11-18T04:38:20Z",
//         "updated_at": "2014-11-18T04:38:20Z"
//       },
//       "sid": "a4f326bb-5cb8-4f1c-be14-70feb21bd00a"
//     }
func CreateUserHandler(w http.ResponseWriter, req *http.Request) {
  fmt.Println("")
  log.Println("Attempting to create User")
  var params Params
  if ok := readBody(&params, w, req); !ok {
    return
  }

  t := time.Now()
  // If no facebook_id is sent, blow up
  fb_id := strings.TrimSpace(params.FacebookId)
  if len([]rune(fb_id)) == 0 {
    http.Error(w, "Improper parameters", http.StatusBadRequest)
    return
  }
  params.User.FacebookId = fb_id

  user := User{}

  // If User already exists, return current user
  if ok := findUserByFacebookId(fb_id, &user, w, req); !ok {
    params.User.CreatedAt = t
    params.User.UpdatedAt = t

    res, err := r.Table("users").Insert(params.User, r.InsertOpts{ReturnChanges: true}).RunWrite(session)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }

    encoding.Decode(&user, res.Changes[0].NewValue) // using reflection
  }

  log.Printf("user.Id = %v\n", user.Id)
  s, err := fetchSessionByUser(user.Id)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
  log.Printf("session = %+v\n", s)

  resp := struct {
    User User   `json:"user"`
    Sid  string `json:"sid"`
  }{ user, s.Id }

  sendJson(resp, w)
}

// ShowUserHandler show a new user
//
// Returns: user object thats saved for that id
//
// Required: id
//
// Example:
//  Request:
//     curl -X GET <HOST_DOMAIN:PORT>/api/v1/users/c8a668fe-2574-47a6-a61e-dcd4a35fde54
//  Response:
//     {
//         "user": {
//             "id": "c8a668fe-2574-47a6-a61e-dcd4a35fde54",
//             "first_name": "Reggie",
//             "last_name": "Bush",
//             "email": "",
//             "avatar": "",
//             "bio": "blah blah blah",
//             "created_at": "2014-11-18T04:38:20Z",
//             "updated_at": "2014-11-18T04:38:20Z"
//         }
//     }
func ShowUserHandler(w http.ResponseWriter, req *http.Request) {
  id := req.URL.Query().Get(":id")
  log.Println("Attempting to show User..." + id)

  user := User{}
  if ok := findUser(id, &user, w, req); !ok {
    return
  }
  sendJson(map[string]interface{}{"user": user}, w)
}

// UpdateUserHandler updates a persisted user object
//
// Returns: updated user object and boolean, "changed" indicating if things were changed
//
// Required: id & sid
//
// Example:
//  Request:
//     curl -X PUT 
//          -DATA '{"user": {                          \
//                    "i_am_not_getting_saved": "test" \
//                    "first_name": "Joe",             \
//                    "last_name": "Montana" } }'
//          <HOST_DOMAIN:PORT>/api/v1/users/c8a668fe-2574-47a6-a61e-dcd4a35fde54
//  Response:
//     {
//         "changed": true,
//         "user": {
//             "avatar": "",
//             "bio": "blah blah blah",
//             "created_at": "2014-11-18T04:38:20Z",
//             "email": "",
//             "facebook_id": "1jj32-sdfs2-sds",
//             "first_name": "Joe",
//             "id": "c8a668fe-2574-47a6-a61e-dcd4a35fde54",
//             "last_name": "Montana",
//             "updated_at": "2014-11-18T17:25:51Z"
//         }
//     }
func UpdateUserHandler(w http.ResponseWriter, req *http.Request) {
  id := req.URL.Query().Get(":id")
  fmt.Println("")
  log.Println("Attempting to update User#" + id)

  var rawParams RawParams
  if ok := readBody(&rawParams, w, req); !ok {
    return
  }

  log.Printf("\n\n%+v\n\n", rawParams)

  user := User{}
  if ok := fetchUserFromSession(&user, rawParams.Sid, id, w, req); !ok {
    return
  }

  changed := false
  if first_name, ok := rawParams.User["first_name"]; ok {
    user.FirstName, changed = first_name, true
  }
  if last_name, ok := rawParams.User["last_name"]; ok {
    user.LastName, changed = last_name, true
  }
  if email, ok := rawParams.User["email"]; ok {
    user.Email, changed = email, true
  }
  if avatar, ok := rawParams.User["avatar"]; ok {
    user.Avatar, changed = avatar, true
  }
  if bio, ok := rawParams.User["bio"]; ok {
    user.Bio, changed = bio, true
  }
  
  if changed {
    user.UpdatedAt = time.Now()

  	_, err := r.Table("users").Get(id).Update(user, r.UpdateOpts{ReturnChanges: true}).RunWrite(session)
  	if err != nil {
  		http.Error(w, err.Error(), http.StatusInternalServerError)
  		return
  	}
  }

  sendJson(map[string]interface{}{"user": user, "changed": changed}, w)
}

// DeleteUserHandler deletes a persisted user object
//
// Returns: boolean "result" indicating result of operation
//
// Required: id & sid
//
// Example:
//  Request:
//     curl -X DELETE <HOST_DOMAIN:PORT>/api/v1/users/c8a668fe-2574-47a6-a61e-dcd4a35fde54
//  Response:
//     {
//         "result": true
//     }
func DeleteUserHandler(w http.ResponseWriter, req *http.Request) {
  id := req.URL.Query().Get(":id")
  log.Println("Attempting to delete User..." + id)

  var rawParams RawParams
  if ok := readBody(&rawParams, w, req); !ok {
    return
  }

  user := User{}
  if ok := fetchUserFromSession(&user, rawParams.Sid, id, w, req); !ok {
    return
  }

	_, err := r.Table("users").Get(id).Delete().RunWrite(session)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = r.Table("events").GetAllByIndex("user_id", id).Delete().RunWrite(session)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

  _, err = r.Table("messages").GetAllByIndex("user_id", id).Delete().RunWrite(session)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

  sendJson(map[string]bool{"result": true}, w)
}

func findUser(id string, u *User, w http.ResponseWriter, req *http.Request) bool {
	res, err := r.Table("users").Get(id).Run(session)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return false
	}

	if res.IsNil() {
		http.NotFound(w, req)
		return false
	}

  res.One(&u)
  return true
}

func findUserByFacebookId(f_id string, u *User, w http.ResponseWriter, req *http.Request) bool {
  res, err := r.Table("users").GetAllByIndex("facebook_id", f_id).Run(session)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return false
	}

	if res.IsNil() {
		return false
	}

  res.One(&u)
  return true
}
