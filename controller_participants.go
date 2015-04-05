package main

import (
  r "github.com/dancannon/gorethink"
  "net/http"
  "log"
  "fmt"
  "time"
  "strconv"
)

// CreateEventParticipantHandler creates a participant request object by the session User for a specific event
//
// Returns: the new message object
//
// Required: <EVENT_ID> && sid
//
// Example:
//   Request:
//     curl -X POST
//          -DATA '{ "sid": "05914cc6-be5d-438d-9e42-b2520f0d6146" } \
//          <HOST_DOMAIN:PORT>/api/:v/events/:event_id/participants
//   Response:
//     {
//         "participant": {
//             "created_at": "2015-04-01T03:11:12Z",
//             "event_id": "2c4cf357-d7a7-438d-bb1f-599c48be3209",
//             "id": "31a7c3b3-a0be-4666-b4b5-bf23ce0ba77c",
//             "request_status": "requested",
//             "response_status": "pending",
//             "updated_at": "2015-04-01T03:11:12Z",
//             "user_id": "20f38193-b9d2-40d4-b60b-6c3cacc2d2e9"
//         }
//     }
func CreateEventParticipantHandler(w http.ResponseWriter, req *http.Request) {
  event_id := req.URL.Query().Get(":event_id")
  fmt.Println("")
  log.Println("Attempting to Create Participant for Event#" + event_id)

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
  participant := &ParticipantWrite{
    UserId:         user.Id,
    EventId:        event.Id,
    RequestStatus:  "requested",
    ResponseStatus: "pending",
    CreatedAt:      t,
    UpdatedAt:      t,
  }

  res, err := r.Table("participants").Insert(participant, r.InsertOpts{ReturnChanges: true}).RunWrite(session)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  sendJson(map[string]interface{}{"participant": res.Changes[0].NewValue}, w)
}

// DeleteParticipantHandler deletes a Participant request object owned by a User
//
// Returns: boolean "result" indicating result of operation
//
// Required: <PARTICIPANT_ID>
//
// Example:
//  Request:
//     curl -X DELETE <HOST_DOMAIN:PORT>/api/:v/participants/:id
//  Response:
//     {
//         "result": true
//     }
func DeleteParticipantHandler(w http.ResponseWriter, req *http.Request) {
  id := req.URL.Query().Get(":id")
  fmt.Println("")
  log.Println("Attempting to Delete Participant#" + id)

  var rawParams RawParams
  if ok := readBody(&rawParams, w, req); !ok {
    return
  }
  log.Printf("\n\n%+v\n\n", rawParams)

  participant := ParticipantWrite{}
  if ok := findParticipant(id, &participant, w, req); !ok {
    return
  }

  user := User{}
  if ok := fetchUserFromSession(&user, rawParams.Sid, participant.UserId, w, req); !ok {
    return
  }

  if participant.UserId != user.Id {
    http.Error(w, "Improper parameters!!!", http.StatusBadRequest)
		return
  }

	_, err := r.Table("participants").Get(id).Delete().RunWrite(session)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

  sendJson(map[string]bool{"result": true}, w)
}

// UpdateParticipantHandler updates a participant object owned by either the participant or event owner
//
// Returns: updated participant object and boolean, "changed" indicating if things were changed
//
// Required: <PARTICIPANT_ID> && sid
//
// Example:
//  Request:
//     curl -X PUT 
//          -DATA '{ "participant" {                  \
//                      "response_status": "accepted" \
//                  },                                \
//                  "sid": "05914cc6-be5d-438d-9e42-b2520f0d6146"}'
//          <HOST_DOMAIN:PORT>/api/:v/participants/:id
//  Response:
//     {
//         "changed": true,
//         "participant": {
//             "id": "31a7c3b3-a0be-4666-b4b5-bf23ce0ba77c",
//             "event_id": "2c4cf357-d7a7-438d-bb1f-599c48be3209",
//             "user_id": "20f38193-b9d2-40d4-b60b-6c3cacc2d2e9",
//             "request_status": "requested",
//             "response_status": "accepted",
//             "created_at": "2015-04-01T03:11:12Z",
//             "updated_at": "2015-03-31T23:15:17.269611648-04:00"
//         }
//     }
func UpdateParticipantHandler(w http.ResponseWriter, req *http.Request) {
  id := req.URL.Query().Get(":id")
  fmt.Println("")
  log.Println("Attempting to Update Participant#" + id)

  participant := ParticipantWrite{}
  if ok := findParticipant(id, &participant, w, req); !ok {
    return
  }

  var rawParams RawParams
  if ok := readBody(&rawParams, w, req); !ok {
    return
  }
  log.Printf("\n\n%+v\n\n", rawParams)

  user := User{}
  if ok := fetchUserFromSession(&user, rawParams.Sid, "", w, req); !ok {
    return
  }

  event := Event{}
  if ok := findEvent(participant.EventId, &event, w, req); !ok {
    return
  }

  if(!((user.Id == participant.UserId) || (user.Id == event.UserId))) {
    http.Error(w, "User does not have permission to update", http.StatusBadRequest)
    return
  }

  changed := false
  if content, ok := rawParams.Participant["request_status"]; ok {
    participant.RequestStatus, changed = content, true
  }
  if ref, ok := rawParams.Participant["response_status"]; ok {
    participant.ResponseStatus, changed = ref, true
  }

  if changed {
    participant.UpdatedAt = time.Now()

  	_, err := r.Table("participants").Get(id).Update(participant, r.UpdateOpts{ReturnChanges: true}).RunWrite(session)
  	if err != nil {
  		http.Error(w, err.Error(), http.StatusInternalServerError)
  		return
  	}
  }

  sendJson(map[string]interface{}{"participant": participant, "changed": changed}, w)
}

// Name/Desc: IndexUserParticipantsHandler shows requested participations 
//              with embedded events for the current user
//
// Returns: Paginated list of participations for the current user
//
// Required: sid
//
// Optional URL Params: page=<Integer> && per=<Integer>
//
// Example:
//   Request:
//     curl -X GET <HOST_DOMAIN:PORT>/api/v1/participants
//   Response:
//     {
//         "page": "1",
//         "participants": [
//             {
//                 "id": "31a7c3b3-a0be-4666-b4b5-bf23ce0ba77c",
//                 "event_id": "2c4cf357-d7a7-438d-bb1f-599c48be3209",
//                 "event": {
//                     "id": "2c4cf357-d7a7-438d-bb1f-599c48be3209",
//                     "user_id": "20f38193-b9d2-40d4-b60b-6c3cacc2d2e9",
//                     "picture_url": "",
//                     "lon": "",
//                     "lat": "",
//                     "location": {
//                         "Lon": 0,
//                         "Lat": 0
//                     },
//                     "title": "SXSW",
//                     "description": "Conference",
//                     "privacy_level": 0,
//                     "start_date": "0001-01-01T00:00:00Z",
//                     "end_date": "0001-01-01T00:00:00Z",
//                     "created_at": "2015-04-01T02:55:21Z",
//                     "updated_at": "2015-04-01T02:55:21Z"
//                 },
//                 "user_id": "20f38193-b9d2-40d4-b60b-6c3cacc2d2e9",
//                 "user": {
//                     "id": "20f38193-b9d2-40d4-b60b-6c3cacc2d2e9",
//                     "first_name": "Joe",
//                     "last_name": "Montana",
//                     "email": "",
//                     "avatar": "",
//                     "bio": "blah blah blah",
//                     "created_at": "2015-04-01T02:49:48Z",
//                     "updated_at": "2015-04-01T02:49:48Z"
//                 },
//                 "request_status": "requested",
//                 "response_status": "accepted",
//                 "created_at": "2015-04-01T03:11:12Z",
//                 "updated_at": "2015-04-01T03:15:17Z"
//             }
//         ],
//         "per": "20"
//     }
func IndexUserParticipantsHandler(w http.ResponseWriter, req *http.Request) {
  sid := req.URL.Query().Get("sid")
  fmt.Println("")
  log.Println("Attempting to list requested participations/events for Current User")

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

  participants := []Participant{}
  res, err := r.Table("participants").Filter(r.Row.Field("user_id").Eq(user.Id)).OrderBy(r.Asc("CreatedAt")).Slice((page-1)*per, page*per).Merge(
    map[string]interface{}{"User": r.Table("users").Get(r.Row.Field("user_id")), "Event": r.Table("events").Get(r.Row.Field("event_id")) }).Run(session)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  err = res.All(&participants)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  sendJson(map[string]interface{}{
    "participants": participants,
    "page": strconv.Itoa(page),
    "per": strconv.Itoa(per),
  }, w)
}

// Name/Desc: IndexEventParticipantsHandler returns a paginated list of participations belonging to an event
//
// Required: <EVENT_ID> && sid
//
// Optional URL Params: page=<Integer> && per=<Integer>
//
// Example:
//   Request:
//     curl -X GET <HOST_DOMAIN:PORT/api/:v/events/:event_id/participants?sid=<SID>
//   Response:
//     {
//         "page": "1",
//         "participants": [
//             {
//                 "id": "31a7c3b3-a0be-4666-b4b5-bf23ce0ba77c",
//                 "event_id": "2c4cf357-d7a7-438d-bb1f-599c48be3209",
//                 "event": {
//                     "id": "",
//                     "user_id": "",
//                     "picture_url": "",
//                     "lon": "",
//                     "lat": "",
//                     "location": {
//                         "Lon": 0,
//                         "Lat": 0
//                     },
//                     "title": "",
//                     "description": "",
//                     "privacy_level": 0,
//                     "start_date": "0001-01-01T00:00:00Z",
//                     "end_date": "0001-01-01T00:00:00Z",
//                     "created_at": "0001-01-01T00:00:00Z",
//                     "updated_at": "0001-01-01T00:00:00Z"
//                 },
//                 "user_id": "20f38193-b9d2-40d4-b60b-6c3cacc2d2e9",
//                 "user": {
//                     "id": "",
//                     "first_name": "",
//                     "last_name": "",
//                     "email": "",
//                     "avatar": "",
//                     "bio": "",
//                     "created_at": "0001-01-01T00:00:00Z",
//                     "updated_at": "0001-01-01T00:00:00Z"
//                 },
//                 "request_status": "requested",
//                 "response_status": "accepted",
//                 "created_at": "2015-04-01T03:11:12Z",
//                 "updated_at": "2015-04-01T03:15:17Z"
//             }
//         ],
//         "per": "20"
//     }
func IndexEventParticipantsHandler(w http.ResponseWriter, req *http.Request) {
  event_id := req.URL.Query().Get(":event_id")
  sid := req.URL.Query().Get("sid")
  fmt.Println("")
  log.Println("Attempting to list Participants for Event#" + event_id)

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

  participants := []Participant{}
  res, err := r.Table("participants").Filter(r.Row.Field("event_id").Eq(event.Id)).OrderBy(r.Asc("CreatedAt")).Slice((page-1)*per, page*per).Merge(
    map[string]interface{}{"User": r.Table("users").Get(r.Row.Field("user_id")), "Event": r.Table("events").Get(r.Row.Field("event_id")) }).Run(session)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  err = res.All(&participants)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  sendJson(map[string]interface{}{
    "participants": participants,
    "page": strconv.Itoa(page),
    "per": strconv.Itoa(per),
  }, w)
}

func findParticipant(id string, p *ParticipantWrite, w http.ResponseWriter, req *http.Request) bool {
	res, err := r.Table("participants").Get(id).Run(session)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return false
	}

	if res.IsNil() {
		http.NotFound(w, req)
		return false
	}

  res.One(&p)
  return true
}
