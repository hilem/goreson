package main

import (
  r "github.com/dancannon/gorethink"
  "github.com/dancannon/gorethink/types"
  "net/http"
  "log"
  "fmt"
  "strconv"
  "time"
)

// Name/Desc: IndexUserEventsHandler - returns a paginated list of events owned by the session user
//
// Required: <USER_ID> && sid
//
// Optional URL Params: page=<Integer> && per=<Integer>
//
// Example:
//   Request:
//     curl -X GET <HOST_DOMAIN:PORT>/api/v1/users/92b4fbfc-77ef-4c12-917c-913394ce6767/events?sid=05914cc6-be5d-438d-9e42-b2520f0d6146
//   Response:
//     {
//         "page": "1",
//         "per": "20",
//         "events": [
//             {
//                 "id": "b135d900-638b-47be-9aa5-5bf21218083b",
//                 "user_id": "82e196a0-554b-487c-b24b-0e1714da00a6",
//                 "picture_url": "",
//                 "lon": "",
//                 "lat": "",
//                 "location": {
//                     "Lon": -75.1641667,
//                     "Lat": 39.9522222
//                 },
//                 "title": "Test",
//                 "description": "Conference",
//                 "privacy_level": 0,
//                 "start_date": "0001-01-01T00:00:00Z",
//                 "end_date": "0001-01-01T00:00:00Z",
//                 "created_at": "2014-11-21T03:07:42Z",
//                 "updated_at": "2014-11-21T03:10:03Z"
//             }
//         ]
//     }
func IndexUserEventsHandler(w http.ResponseWriter, req *http.Request) {
  id := req.URL.Query().Get(":user_id")
  sid := req.URL.Query().Get("sid")
  fmt.Println("")
  log.Println("Attempting to list Events for User#" + id)

  user := User{}
  if ok := fetchUserFromSession(&user, sid, id, w, req); !ok {
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

  events := []Event{}
  res, err := r.Table("events").Filter(r.Row.Field("user_id").Eq(user.Id)).OrderBy(r.Asc("CreatedAt")).Slice((page-1)*per, page*per).Run(session)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  err = res.All(&events)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  sendJson(map[string]interface{}{
    "events": events,
    "page": strconv.Itoa(page),
    "per": strconv.Itoa(per),
  }, w)
}

// Name/Desc: IndexEventsHandler returns a paginated list of events that are centered around a location
//
// TODO: this should be geo-based, if no lat/lon passed in, it should pick a random location
//
// Optional URL Params: page=<Integer> && per=<Integer>
//
// Example:
//   Request:
//     curl -X GET <HOST_DOMAIN:PORT>/api/v1/events
//   Response:
//     {
//         "page": "1",
//         "per": "20",
//         "events": [
//             {
//                 "Id": "7700de02-214e-4367-8402-c30581c83c37",
//                 "UserId": "92b4fbfc-77ef-4c12-917c-913394ce6767",
//                 "Location": {
//                     "Lon": -73.1641667,
//                     "Lat": 41.9522222
//                 },
//                 "Title": "All The Drones",
//                 "Description": "Robot Wars",
//                 "PrivacyLevel": 0,
//                 "StartDate": "0001-01-01T00:00:00Z",
//                 "EndDate": "0001-01-01T00:00:00Z",
//                 "CreatedAt": "2014-10-24T02:17:10Z",
//                 "UpdatedAt": "2014-10-24T02:17:10Z"
//             },
//             {
//                 "Id": "d77ee502-1911-4ef9-8afa-b5cd90912441",
//                 "UserId": "92b4fbfc-77ef-4c12-917c-913394ce6767",
//                 "Location": {
//                     "Lon": -75.1641667,
//                     "Lat": 39.9522222
//                 },
//                 "Title": "SXSW",
//                 "Description": "Conference",
//                 "PrivacyLevel": 0,
//                 "StartDate": "0001-01-01T00:00:00Z",
//                 "EndDate": "0001-01-01T00:00:00Z",
//                 "CreatedAt": "2014-10-24T01:51:21Z",
//                 "UpdatedAt": "2014-10-24T01:51:21Z"
//             }
//         ]
//     }
func IndexEventsHandler(w http.ResponseWriter, req *http.Request) {
  fmt.Println("")
  log.Println("Attempting to list Events")
  events := []Event{}

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

  res, err := r.Table("events").OrderBy(r.Asc("CreatedAt")).Slice((page-1)*per, page*per).Run(session)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  err = res.All(&events)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  sendJson(map[string]interface{}{
    "events": events,
    "page": strconv.Itoa(page),
    "per": strconv.Itoa(per),
  }, w)
}

// CreateUserEventHandler creates an event for the session User
//
// Returns: the new event object
//
// Required: <USER_ID> && sid
//
// Example:
//   Request:
//     curl -X POST
//          -DATA '{"event": {                                      \
//                    "title": "SXSW",                              \
//                    "description": "Conference",                  \
//                    "lon": "-75.1641667",                         \
//                    "lat": "39.9522222"                           \
//                  },                                              \
//                  "sid": "05914cc6-be5d-438d-9e42-b2520f0d6146" } \
//          <HOST_DOMAIN:PORT>/api/v1/users/82e196a0-554b-487c-b24b-0e1714da00a6/events
//   Response:
//     {
//         "event": {
//             "created_at": "2014-11-21T03:07:42Z",
//             "description": "Conference",
//             "end_date": "0001-01-01T00:00:00Z",
//             "id": "b135d900-638b-47be-9aa5-5bf21218083b",
//             "location": {
//                 "Lat": 39.9522222,
//                 "Lon": -75.1641667
//             },
//             "picture_url": "",
//             "privacy_level": 0,
//             "start_date": "0001-01-01T00:00:00Z",
//             "title": "SXSW",
//             "updated_at": "2014-11-21T03:07:42Z",
//             "user_id": "82e196a0-554b-487c-b24b-0e1714da00a6"
//         }
//     }
func CreateUserEventHandler(w http.ResponseWriter, req *http.Request) {
  id := req.URL.Query().Get(":user_id")
  fmt.Println("")
  log.Println("Attempting to Create Event for User#" + id)

  var rawParams RawParams
  if ok := readBody(&rawParams, w, req); !ok {
    return
  }
  log.Printf("\n\n%+v\n\n", rawParams)

  user := User{}
  if ok := fetchUserFromSession(&user, rawParams.Sid, id, w, req); !ok {
    return
  }

  t := time.Now()
  event := &Event{
    UserId:       user.Id,
    Title:        rawParams.Event["title"],
    Description:  rawParams.Event["description"],
    CreatedAt:    t,
    UpdatedAt:    t,
  }
  if pict, ok := rawParams.Event["picture_url"]; ok {
    event.PictureUrl = pict
  }
  if priv_lvl, ok := rawParams.Event["privacy_level"]; ok {
    event.PrivacyLevel, _ = strconv.Atoi(priv_lvl)
  }
  if _, ok := rawParams.Event["lon"]; ok {
    if _, ok := rawParams.Event["lat"]; ok {
      lon, _ := strconv.ParseFloat(rawParams.Event["lon"], 64)
      lat, _ := strconv.ParseFloat(rawParams.Event["lat"], 64)
      event.Location = types.Point{ Lon: lon, Lat: lat, }
    }
  }
  if s_date, ok := rawParams.Event["start_date"]; ok {
    event.StartDate, _ = time.Parse(TimeFormat, s_date)
  }
  if e_date, ok := rawParams.Event["end_date"]; ok {
    event.EndDate, _ = time.Parse(TimeFormat, e_date)
  }

  res, err := r.Table("events").Insert(event, r.InsertOpts{ReturnChanges: true}).RunWrite(session)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  sendJson(map[string]interface{}{"event": res.Changes[0].NewValue}, w)
}

// UpdateUserEventHandler updates a persisted event object owned by a User
//
// Returns: updated event object and boolean, "changed" indicating if things were changed
//
// Required: <USER_ID> && <EVENT_ID> && sid
//
// Example:
//  Request:
//     curl -X PUT 
//          -DATA '{ "event" { "title": "Test"}, "sid": "05914cc6-be5d-438d-9e42-b2520f0d6146"}'
//          <HOST_DOMAIN:PORT>/api/v1/users/<USER_ID>/events/<EVENT_ID>
//  Response:
//     {
//         "changed": true,
//         "event": {
//             "id": "b135d900-638b-47be-9aa5-5bf21218083b",
//             "user_id": "82e196a0-554b-487c-b24b-0e1714da00a6",
//             "picture_url": "",
//             "lon": "",
//             "lat": "",
//             "location": {
//                 "Lon": -75.1641667,
//                 "Lat": 39.9522222
//             },
//             "title": "Test",
//             "description": "Conference",
//             "privacy_level": 0,
//             "start_date": "0001-01-01T00:00:00Z",
//             "end_date": "0001-01-01T00:00:00Z",
//             "created_at": "2014-11-21T03:07:42Z",
//             "updated_at": "2014-11-20T22:10:03.921444177-05:00"
//         }
//     }
func UpdateUserEventHandler(w http.ResponseWriter, req *http.Request) {
  user_id := req.URL.Query().Get(":user_id")
  id := req.URL.Query().Get(":id")
  fmt.Println("")
  log.Println("Attempting to Update Event#" + id + " by User#" + user_id)

  var rawParams RawParams
  if ok := readBody(&rawParams, w, req); !ok {
    return
  }
  log.Printf("\n\n%+v\n\n", rawParams)

  user := User{}
  if ok := fetchUserFromSession(&user, rawParams.Sid, user_id, w, req); !ok {
    return
  }

  event := Event{}
  if ok := findEvent(id, &event, w, req); !ok {
    return
  }

  if event.UserId != user.Id {
    http.Error(w, "Improper parameters!!!", http.StatusBadRequest)
		return
  }

  changed := false
  if title, ok := rawParams.Event["title"]; ok {
    event.Title, changed = title, true
  }
  if description, ok := rawParams.Event["description"]; ok {
    event.Description, changed = description, true
  }
  if priv_lvl, ok := rawParams.Event["privacy_level"]; ok {
    event.PrivacyLevel, _ = strconv.Atoi(priv_lvl)
    changed = true
  }
  if s_date, ok := rawParams.Event["start_date"]; ok {
    event.StartDate, _ = time.Parse(TimeFormat, s_date)
    changed = true
  }
  if e_date, ok := rawParams.Event["end_date"]; ok {
    event.EndDate, _ = time.Parse(TimeFormat, e_date)
    changed = true
  }
  if pict, ok := rawParams.Event["picture_url"]; ok {
    event.PictureUrl, changed = pict, true
  }
  if _, ok := rawParams.Event["lon"]; ok {
    if _, ok := rawParams.Event["lat"]; ok {
      lon, _ := strconv.ParseFloat(rawParams.Event["lon"], 64)
      lat, _ := strconv.ParseFloat(rawParams.Event["lat"], 64)
      event.Location = types.Point{ Lon: lon, Lat: lat, }
    }
  }

  if changed {
    event.UpdatedAt = time.Now()

  	_, err := r.Table("events").Get(id).Update(event, r.UpdateOpts{ReturnChanges: true}).RunWrite(session)
  	if err != nil {
  		http.Error(w, err.Error(), http.StatusInternalServerError)
  		return
  	}
  }

  sendJson(map[string]interface{}{"event": event, "changed": changed}, w)
}

// ShowEventHandler show a new user
//
// Required: <EVENT_ID>
//
// Returns: Event object corresponding to the ID passed in
//
// Example:
//  Request:
//     curl -X GET <HOST_DOMAIN:PORT>/api/v1/events/<EVENT_ID>
//  Response:
//     {
//         "event": {
//             "id": "b135d900-638b-47be-9aa5-5bf21218083b",
//             "user_id": "82e196a0-554b-487c-b24b-0e1714da00a6",
//             "picture_url": "",
//             "lon": "",
//             "lat": "",
//             "location": {
//                 "Lon": -75.1641667,
//                 "Lat": 39.9522222
//             },
//             "title": "Test",
//             "description": "Conference",
//             "privacy_level": 0,
//             "start_date": "0001-01-01T00:00:00Z",
//             "end_date": "0001-01-01T00:00:00Z",
//             "created_at": "2014-11-21T03:07:42Z",
//             "updated_at": "2014-11-21T03:10:03Z"
//         }
//     }
func ShowEventHandler(w http.ResponseWriter, req *http.Request) {
  id := req.URL.Query().Get(":id")
  fmt.Println("")
  log.Println("Attempting to show Event#" + id)

  event := Event{}
  if ok := findEvent(id, &event, w, req); !ok {
    return
  }
  sendJson(map[string]interface{}{"event": event}, w)
}

// DeleteUserEventHandler deletes a persisted Event object owned by a User
//
// Returns: boolean "result" indicating result of operation
//
// Required: <USER_ID> of owner && <EVENT_ID> of event to be deleted
//
// Example:
//  Request:
//     curl -X DELETE <HOST_DOMAIN:PORT>/api/v1/users/<USER_ID>/events/<EVENT_ID>
//  Response:
//     {
//         "result": true
//     }
func DeleteUserEventHandler(w http.ResponseWriter, req *http.Request) {
  user_id := req.URL.Query().Get(":user_id")
  id := req.URL.Query().Get(":id")
  fmt.Println("")
  log.Println("Attempting to Delete Event#" + id + " by User#" + user_id)

  var rawParams RawParams
  if ok := readBody(&rawParams, w, req); !ok {
    return
  }
  log.Printf("\n\n%+v\n\n", rawParams)

  user := User{}
  if ok := fetchUserFromSession(&user, rawParams.Sid, user_id, w, req); !ok {
    return
  }

  event := Event{}
  if ok := findEvent(id, &event, w, req); !ok {
    return
  }

  if event.UserId != user.Id {
    http.Error(w, "Improper parameters!!!", http.StatusBadRequest)
		return
  }

	_, err := r.Table("events").Get(id).Delete().RunWrite(session)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

  sendJson(map[string]bool{"result": true}, w)
}

func findEvent(id string, e *Event, w http.ResponseWriter, req *http.Request) bool {
	res, err := r.Table("events").Get(id).Run(session)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return false
	}

	if res.IsNil() {
		http.NotFound(w, req)
		return false
	}

  res.One(&e)
  return true
}
