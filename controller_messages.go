package main

import (
  r "github.com/dancannon/gorethink"
  "net/http"
  "log"
  "fmt"
  "time"
  "strconv"
)

// CreateEventMessageHandler creates a message by the session User for a specific event
//
// Returns: the new message object
//
// Required: <EVENT_ID> && sid
//
// Example:
//   Request:
//     curl -X POST
//          -DATA '{"message": {                                    \
//                    "content": "Let's get some coffee!"           \
//                  },                                              \
//                  "sid": "05914cc6-be5d-438d-9e42-b2520f0d6146" } \
//          <HOST_DOMAIN:PORT>/api/v1/events/:event_id/messages
//   Response:
//     {
//         "message": {
//             "content": "Let's get some coffee!",
//             "created_at": "2014-11-23T16:29:24Z",
//             "event_id": "d77bea99-856c-4189-8875-2df0cef55abe",
//             "id": "98cd27f1-350e-4657-943c-0c8e4d41f75d",
//             "references": "",
//             "updated_at": "2014-11-23T16:29:24Z",
//             "user_id": "82e196a0-554b-487c-b24b-0e1714da00a6"
//         }
//     }
func CreateEventMessageHandler(w http.ResponseWriter, req *http.Request) {
  event_id := req.URL.Query().Get(":event_id")
  fmt.Println("")
  log.Println("Attempting to Create Message for Event#" + event_id)

  var rawParams RawParams
  if ok := readBody(&rawParams, w, req); !ok {
    return
  }
  log.Printf("\n\n%+v\n\n", rawParams)

  event := Event{}
  if ok := findEvent(event_id, &event, w, req); !ok {
    return
  }

  user := User{}
  if ok := fetchUserFromSession(&user, rawParams.Sid, "", w, req); !ok {
    return
  }

  t := time.Now()
  message := &Message{
    UserId:       user.Id,
    EventId:      event.Id,
    Content:      rawParams.Message["content"],
    References:   rawParams.Message["references"],
    CreatedAt:    t,
    UpdatedAt:    t,
  }

  res, err := r.Table("messages").Insert(message, r.InsertOpts{ReturnChanges: true}).RunWrite(session)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  sendJson(map[string]interface{}{"message": res.Changes[0].NewValue}, w)
}

// DeleteMessageHandler deletes a persisted Message object owned by a User
//
// Returns: boolean "result" indicating result of operation
//
// Required: <MESSAGE_ID>
//
// Example:
//  Request:
//     curl -X DELETE <HOST_DOMAIN:PORT>/api/v1/messages/<MESSAGE_ID>
//  Response:
//     {
//         "result": true
//     }
func DeleteMessageHandler(w http.ResponseWriter, req *http.Request) {
  id := req.URL.Query().Get(":id")
  fmt.Println("")
  log.Println("Attempting to Delete Message#" + id)

  var rawParams RawParams
  if ok := readBody(&rawParams, w, req); !ok {
    return
  }
  log.Printf("\n\n%+v\n\n", rawParams)

  message := Message{}
  if ok := findMessage(id, &message, w, req); !ok {
    return
  }

  user := User{}
  if ok := fetchUserFromSession(&user, rawParams.Sid, message.UserId, w, req); !ok {
    return
  }

  if message.UserId != user.Id {
    http.Error(w, "Improper parameters!!!", http.StatusBadRequest)
		return
  }

	_, err := r.Table("messages").Get(id).Delete().RunWrite(session)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

  sendJson(map[string]bool{"result": true}, w)
}

// UpdateMessageHandler updates a persisted message object owned by the current session User
//
// Returns: updated message object and boolean, "changed" indicating if things were changed
//
// Required: <MESSAGE_ID> && sid
//
// Example:
//  Request:
//     curl -X PUT 
//          -DATA '{ "message" { "content": "No thanks."}, "sid": "05914cc6-be5d-438d-9e42-b2520f0d6146"}'
//          <HOST_DOMAIN:PORT>/api/v1/messages/<MESSAGE_ID>
//  Response:
//     {
//         "changed": true,
//         "message": {
//             "id": "98cd27f1-350e-4657-943c-0c8e4d41f75d",
//             "user_id": "82e196a0-554b-487c-b24b-0e1714da00a6",
//             "event_id": "d77bea99-856c-4189-8875-2df0cef55abe",
//             "references": "",
//             "content": "No thanks.",
//             "created_at": "2014-11-23T16:29:24Z",
//             "updated_at": "2014-11-23T11:32:29.575053997-05:00"
//         }
//     }
func UpdateMessageHandler(w http.ResponseWriter, req *http.Request) {
  id := req.URL.Query().Get(":id")
  fmt.Println("")
  log.Println("Attempting to Update Message#" + id)

  message := Message{}
  if ok := findMessage(id, &message, w, req); !ok {
    return
  }

  var rawParams RawParams
  if ok := readBody(&rawParams, w, req); !ok {
    return
  }
  log.Printf("\n\n%+v\n\n", rawParams)

  user := User{}
  if ok := fetchUserFromSession(&user, rawParams.Sid, message.UserId, w, req); !ok {
    return
  }

  changed := false
  if content, ok := rawParams.Message["content"]; ok {
    message.Content, changed = content, true
  }
  if ref, ok := rawParams.Message["references"]; ok {
    message.References, changed = ref, true
  }

  if changed {
    message.UpdatedAt = time.Now()

  	_, err := r.Table("messages").Get(id).Update(message, r.UpdateOpts{ReturnChanges: true}).RunWrite(session)
  	if err != nil {
  		http.Error(w, err.Error(), http.StatusInternalServerError)
  		return
  	}
  }

  sendJson(map[string]interface{}{"message": message, "changed": changed}, w)
}

// Name/Desc: IndexUserMessagesHandler shows linked messages for the current user
//
// Returns: Paginated list of messages for the current user
//
// Required: sid
//
// Optional URL Params: page=<Integer> && per=<Integer>
//
// Example:
//   Request:
//     curl -X GET <HOST_DOMAIN:PORT>/api/v1/messages
//   Response:
//     {
//         "page": "1",
//         "per": "20",
//         "messages": [
//             {
//                 "id": "98cd27f1-350e-4657-943c-0c8e4d41f75d",
//                 "user_id": "82e196a0-554b-487c-b24b-0e1714da00a6",
//                 "event_id": "d77bea99-856c-4189-8875-2df0cef55abe",
//                 "references": "",
//                 "content": "Let's get some coffee!",
//                 "created_at": "2014-11-23T16:29:24Z",
//                 "updated_at": "2014-11-23T16:29:24Z"
//             }
//         ]
//     }
func IndexUserMessagesHandler(w http.ResponseWriter, req *http.Request) {
  sid := req.URL.Query().Get("sid")
  fmt.Println("")
  log.Println("Attempting to list Messages for Current User")

  user := User{}
  if ok := fetchUserFromSession(&user, sid, "", w, req); !ok {
    return
  }

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

  messages := []Message{}
  res, err := r.Table("messages").Filter(r.Row.Field("user_id").Eq(user.Id)).OrderBy(r.Asc("CreatedAt")).Slice((page-1)*per, page*per).Run(session)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  err = res.All(&messages)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  sendJson(map[string]interface{}{
    "messages": messages,
    "page": strconv.Itoa(page),
    "per": strconv.Itoa(per),
  }, w)
}

// Name/Desc: IndexEventMessagesHandler returns a paginated list of messages belonging to an event
//
// Required: <EVENT_ID> && sid
//
// Optional URL Params: page=<Integer> && per=<Integer>
//
// Example:
//   Request:
//     curl -X GET <HOST_DOMAIN:PORT>/api/v1/events/<EVENT_ID>/messages?sid=05914cc6-be5d-438d-9e42-b2520f0d6146
//   Response:
//     {
//         "page": "1",
//         "per": "20",
//         "messages": [
//             {
//                 "id": "98cd27f1-350e-4657-943c-0c8e4d41f75d",
//                 "user_id": "82e196a0-554b-487c-b24b-0e1714da00a6",
//                 "event_id": "d77bea99-856c-4189-8875-2df0cef55abe",
//                 "references": "",
//                 "content": "No thanks.",
//                 "created_at": "2014-11-23T16:29:24Z",
//                 "updated_at": "2014-11-23T16:32:29Z"
//             }
//         ]
//     }
func IndexEventMessagesHandler(w http.ResponseWriter, req *http.Request) {
  event_id := req.URL.Query().Get(":event_id")
  sid := req.URL.Query().Get("sid")
  fmt.Println("")
  log.Println("Attempting to list Messages for Event#" + event_id)

  event := Event{}
  if ok := findEvent(event_id, &event, w, req); !ok {
    return
  }

  user := User{}
  if ok := fetchUserFromSession(&user, sid, "", w, req); !ok {
    return
  }

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

  messages := []Message{}
  res, err := r.Table("messages").Filter(r.Row.Field("event_id").Eq(event.Id)).OrderBy(r.Asc("CreatedAt")).Slice((page-1)*per, page*per).Run(session)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  err = res.All(&messages)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  sendJson(map[string]interface{}{
    "messages": messages,
    "page": strconv.Itoa(page),
    "per": strconv.Itoa(per),
  }, w)
}

func findMessage(id string, m *Message, w http.ResponseWriter, req *http.Request) bool {
	res, err := r.Table("messages").Get(id).Run(session)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return false
	}

	if res.IsNil() {
		http.NotFound(w, req)
		return false
	}

  res.One(&m)
  return true
}
