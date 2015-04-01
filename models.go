package main

import (
  "github.com/dancannon/gorethink/types"
	"time"
)

//// OLD
// type TodoItem struct {
//   Id      string `gorethink:"id,omitempty"`
//   Text    string
//   Status  string
//   Created time.Time
// }
//
// func (t *TodoItem) Completed() bool {
//   return t.Status == "complete"
// }
//
// func NewTodoItem(text string) *TodoItem {
//   return &TodoItem{
//     Text:   text,
//     Status: "active",
//   }
// }

//////

type User struct {
	Id         string        `gorethink:"id,omitempty"  json:"id"`
  FirstName  string        `gorethink:"first_name"    json:"first_name"`
  LastName   string        `gorethink:"last_name"     json:"last_name"`
  Email      string        `gorethink:"email"         json:"email"`
	Avatar     string        `gorethink:"avatar"        json:"avatar"`
  Bio        string        `gorethink:"bio"           json:"bio"`
  FacebookId string        `gorethink:"facebook_id"   json:"-"`
	CreatedAt  time.Time     `gorethink:"created_at"    json:"created_at"`
  UpdatedAt  time.Time     `gorethink:"updated_at"    json:"updated_at"`
}

type Event struct {
  Id           string      `gorethink:"id,omitempty"  json:"id"`
  UserId       string      `gorethink:"user_id"       json:"user_id"`
  PictureUrl   string      `gorethink:"picture_url"   json:"picture_url"`
  Lon          string      `gorethink:"-"             json:"lon"`
  Lat          string      `gorethink:"-"             json:"lat"`
  Location     types.Point `gorethink:"location"      json:"location"`
  Title        string      `gorethink:"title"         json:"title"`
  Description  string      `gorethink:"description"   json:"description"`
  PrivacyLevel int         `gorethink:"privacy_level" json:"privacy_level"`
  StartDate    time.Time   `gorethink:"start_date"    json:"start_date"`
  EndDate      time.Time   `gorethink:"end_date"      json:"end_date"`
  CreatedAt    time.Time   `gorethink:"created_at"    json:"created_at"`
  UpdatedAt    time.Time   `gorethink:"updated_at"    json:"updated_at"`
}

type Message struct {
  Id         string        `gorethink:"id,omitempty"  json:"id"`
  UserId     string        `gorethink:"user_id"       json:"user_id"`
  EventId    string        `gorethink:"event_id"      json:"event_id"`
  References string        `gorethink:"references"    json:"references"`
  Content    string        `gorethink:"content"       json:"content"`
  CreatedAt  time.Time     `gorethink:"created_at"    json:"created_at"`
  UpdatedAt  time.Time     `gorethink:"updated_at"    json:"updated_at"`
}

type Participant struct {
  Id             string    `gorethink:"id,omitempty"    json:"id"`
  EventId        string    `gorethink:"event_id"        json:"event_id"`
  Event          Event     `gorethink:"event"           json:"event"`
  UserId         string    `gorethink:"user_id"         json:"user_id"`
  User           User      `gorethink:"user"            json:"user"`
  RequestStatus  string    `gorethink:"request_status"  json:"request_status"`
  ResponseStatus string    `gorethink:"response_status" json:"response_status"`
  CreatedAt      time.Time `gorethink:"created_at"      json:"created_at"`
  UpdatedAt      time.Time `gorethink:"updated_at"      json:"updated_at"`
}

type UserSession struct {
  Id         string        `gorethink:"id,omitempty"  json:"id"`
  UserId     string        `gorethink:"user_id"       json:"user_id"`
  CreatedAt  time.Time     `gorethink:"created_at"    json:"created_at"`
  UpdatedAt  time.Time     `gorethink:"updated_at"    json:"updated_at"`
}

// Helper Structs

type Params struct {
  User        User        `json:"user"`
  Event       Event       `json:"event"`
  Message     Message     `json:"message"`
  Participant Participant `json:"participant"`
  FacebookId  string      `json:"facebook_id"`
  Sid         string      `json:"sid"`
}

type RawParams struct {
  User        map[string]string `json:"user"`
  Event       map[string]string `json:"event"`
  Message     map[string]string `json:"message"`
  Participant map[string]string `json:"participant"`
  FacebookId  string            `json:"facebook_id"`
  Sid         string            `json:"sid"`
}


// var geoJson = {
//     'type': 'Point',
//     'coordinates': [ -122.423246, 37.779388 ]
// };
// r.table('geo').insert({
//     id: 'sfo',
//     name: 'San Francisco',
//     location: r.geojson(geoJson)

// create_table "events", force: true do |t|
//   t.text     "description"
//   t.integer  "privacy_level"
//   t.spatial  "lonlat",        limit: {:srid=>4326, :type=>"point", :geographic=>true}
// end
//
// add_index "events", ["lonlat"], :name => "index_events_on_lonlat", :spatial => true
//
// // type Metric struct {
// //   Id         string `gorethink:"id,omitempty"`
// //   DeviceId   string
// //   UserId     string
// //   Created    time.Time
// // }
