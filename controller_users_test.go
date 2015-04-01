package main

import (
  r "github.com/dancannon/gorethink"
  "log"
  "time"
  "encoding/json"

  "testing"
  "net/http"
  "net/http/httptest"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/suite"
)

type SuiteTester struct {
	suite.Suite
}

// Start Suite
func TestUsersControllerTestSuite(t *testing.T) {
  suite.Run(t, new(SuiteTester))
}

func (suite *SuiteTester) SetupSuite() {
  session = InitTestDB()
  _, err := r.Db("gadder_test").TableCreate("users").RunWrite(session)
  if err != nil {
    log.Println(err)
  }
}

// Wipe tables tested in suite and reload fixtures before each test
func (suite *SuiteTester) SetupTest() {
  r.Table("users").Delete().RunWrite(session)
  user_fixtures := make([]User, 4)
  user_fixtures[0] = User{
    FirstName:  "Tyrion",
    LastName:   "Lannister",
    Email:      "tyrion@lannister.com",
    Bio:        "Younger brother to Cersei and Jaime.",
    FacebookId: "0b8a2b98-f2c5-457a-adc0-34d10a6f3b5c",
    CreatedAt:  time.Date(2008, time.June, 13, 18, 30, 10, 0, time.UTC),
    UpdatedAt:  time.Date(2014, time.October, 5, 18, 30, 10, 0, time.UTC),
  }
  user_fixtures[1] = User{
    FirstName:  "Tywin",
    LastName:   "Lannister",
    Email:      "tywin@lannister.com",
    Bio:        "Lord of Casterly Rock, Shield of Lannisport and Warden of the West.",
    FacebookId: "bb2d8a7b-92e6-4baf-b4f7-b664bdeee25b",
    CreatedAt:  time.Date(1980, time.July, 14, 18, 30, 10, 0, time.UTC),
    UpdatedAt:  time.Date(2014, time.October, 6, 18, 30, 10, 0, time.UTC),
  }
  user_fixtures[2] = User{
    FirstName:  "Jaime",
    LastName:   "Lannister",
    Email:      "jaime@lannister.com",
    Bio:        "Nicknamed 'Kingslayer' for killing the previous King, Aerys II.",
    FacebookId: "d4c19866-eaff-4417-a1c1-93882162606d",
    CreatedAt:  time.Date(2000, time.September, 15, 18, 30, 10, 0, time.UTC),
    UpdatedAt:  time.Date(2014, time.October, 7, 18, 30, 10, 0, time.UTC),
  }
  user_fixtures[3] = User{
    FirstName:  "Cersei",
    LastName:   "Lannister",
    Email:      "cersei@lannister.com",
    Bio:        "Queen of the Seven Kingdoms of Westeros, is the wife of King Robert Baratheon.",
    FacebookId: "251d74d8-7462-4f2a-b132-6f7e429507e5",
    CreatedAt:  time.Date(2002, time.May, 12, 18, 30, 10, 0, time.UTC),
    UpdatedAt:  time.Date(2014, time.October, 8, 18, 30, 10, 0, time.UTC),
  }

  r.Table("users").Insert(user_fixtures).RunWrite(session)
}

func (suite *SuiteTester) TestUserIndexHandler() {
  resp := httptest.NewRecorder()
  req, err := http.NewRequest("GET", "/api/v1/users", nil)
  assert.Nil(suite.T(), err)
  IndexUsersHandler(resp, req)
  assert.Equal(suite.T(), 200, resp.Code)
  assert.Equal(suite.T(), "application/json", resp.Header().Get("Content-Type"))

  type Response struct {
    Users []User `json:"users"`
  }

  res := &Response{}
	err = json.Unmarshal([]byte(resp.Body.String()), &res)
  assert.Nil(suite.T(), err)
  assert.Equal(suite.T(), 4, len(res.Users))
}

func (suite *SuiteTester) TestUserCreateHandler() {
  // test various user models
  // test response code 200
  
  // user := User{ Name: "Brian Jones", Bio: "I got nothing" }
  // body, err = json.Marshal(user)
  // if err != nil {
  //   log.Println("Unable to marshal user")
  // }
  // req, err := http.NewRequest("POST", "/api/v1/users", bytes.NewReader(body))
}

func (suite *SuiteTester) TestUserDeleteHandler() {
  // test db only has 3 users remaining
  // test response code 200
}

func (suite *SuiteTester) TestUserShowHandler() {
  // test user is returned as expected
  // test response code 200
}

func (suite *SuiteTester) TestUserUpdateHandler() {
  // test user returned has updates includes
  // test updated field has changed
  // test response code 200
}
