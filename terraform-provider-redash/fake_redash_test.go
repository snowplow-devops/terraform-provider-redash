//
// Copyright (c) 2020-2026 Snowplow Analytics Ltd. All rights reserved.
//
// This program is licensed to you under the Apache License Version 2.0,
// and you may not use this file except in compliance with the Apache License Version 2.0.
// You may obtain a copy of the Apache License Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0.
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the Apache License Version 2.0 is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the Apache License Version 2.0 for the specific language governing permissions and limitations there under.
//
package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/snowplow-devops/redash-client-go/redash"
)

// fakeDefaultGroupID is the group Redash puts every new user and data source in.
const fakeDefaultGroupID = 2

const fakeTimestamp = "2026-01-02T03:04:05Z"

// fakeRedash is an in-memory stand-in for the parts of the Redash API the
// provider uses, so resources can be tested against the real client.
type fakeRedash struct {
	mu sync.Mutex

	// failWith makes every request fail with this status code; failOn does
	// the same for single requests, keyed by "METHOD /path".
	failWith int
	failOn   map[string]int
	requests []string

	destinations      map[int]map[string]interface{}
	nextDestinationID int

	users      map[int]*fakeUser
	nextUserID int

	groups      map[int]*fakeGroup
	nextGroupID int

	dataSources      map[int]*fakeDataSource
	nextDataSourceID int
}

type fakeUser struct {
	ID                  int         `json:"id"`
	Name                string      `json:"name"`
	Email               string      `json:"email"`
	Groups              []int       `json:"groups"`
	AuthType            string      `json:"auth_type"`
	IsDisabled          bool        `json:"is_disabled"`
	IsInvitationPending bool        `json:"is_invitation_pending"`
	IsEmailVerified     bool        `json:"is_email_verified"`
	ProfileImageURL     string      `json:"profile_image_url"`
	CreatedAt           string      `json:"created_at"`
	UpdatedAt           string      `json:"updated_at"`
	ActiveAt            string      `json:"active_at"`
	DisabledAt          interface{} `json:"disabled_at"`
}

type fakeGroup struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Permissions []string `json:"permissions"`
	CreatedAt   string   `json:"created_at"`
}

type fakeDataSource struct {
	ID                 int                    `json:"id"`
	Name               string                 `json:"name"`
	Type               string                 `json:"type"`
	Syntax             string                 `json:"syntax"`
	Paused             int                    `json:"paused"`
	PauseReason        string                 `json:"pause_reason"`
	QueueName          string                 `json:"queue_name"`
	ScheduledQueueName string                 `json:"scheduled_queue_name"`
	Options            map[string]interface{} `json:"options"`
	Groups             map[int]bool           `json:"groups"`
}

// fakeDataSourceTypes is served from /api/data_sources/types. Only "pg" is
// listed, so the client validates options for pg and passes any other type
// through untouched.
const fakeDataSourceTypes = `[
  {
    "type": "pg",
    "name": "PostgreSQL",
    "configuration_schema": {
      "type": "object",
      "required": ["dbname"],
      "secret": ["password"],
      "properties": {
        "host": {"type": "string"},
        "port": {"type": "number"},
        "user": {"type": "string"},
        "password": {"type": "string"},
        "dbname": {"type": "string"},
        "sslmode": {"type": "string"}
      }
    }
  }
]`

func newFakeRedash(t *testing.T) (*fakeRedash, *redash.Client) {
	t.Helper()

	f := &fakeRedash{
		failOn:            map[string]int{},
		destinations:      map[int]map[string]interface{}{},
		nextDestinationID: 1,
		users:             map[int]*fakeUser{},
		nextUserID:        1,
		groups:            map[int]*fakeGroup{},
		nextGroupID:       1,
		dataSources:       map[int]*fakeDataSource{},
		nextDataSourceID:  1,
	}
	server := httptest.NewServer(f)
	t.Cleanup(server.Close)

	c, err := redash.NewClient(&redash.Config{RedashURI: server.URL, APIKey: "test-key"})
	if err != nil {
		t.Fatalf("creating client: %s", err)
	}

	return f, c
}

func (f *fakeRedash) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := r.Method + " " + r.URL.Path

	f.mu.Lock()
	f.requests = append(f.requests, key)
	status := f.failOn[key]
	if status == 0 {
		status = f.failWith
	}
	f.mu.Unlock()

	if status != 0 {
		http.Error(w, `{"message": "boom"}`, status)
		return
	}

	path := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/"), "/")
	switch path[0] {
	case "destinations":
		f.handleDestinations(w, r)
	case "users":
		f.handleUsers(w, r, path[1:])
	case "groups":
		f.handleGroups(w, r, path[1:])
	case "data_sources":
		f.handleDataSources(w, r, path[1:])
	default:
		http.NotFound(w, r)
	}
}

// requestsMatching returns the requests made so far that start with prefix,
// e.g. "POST /api/users".
func (f *fakeRedash) requestsMatching(prefix string) []string {
	f.mu.Lock()
	defer f.mu.Unlock()

	var matched []string
	for _, r := range f.requests {
		if strings.HasPrefix(r, prefix) {
			matched = append(matched, r)
		}
	}
	return matched
}

func decodeBody(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	body, _ := io.ReadAll(r.Body)
	if err := json.Unmarshal(body, v); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	_ = json.NewEncoder(w).Encode(v)
}

func notFound(w http.ResponseWriter) {
	http.Error(w, `{"message": "Not found"}`, http.StatusNotFound)
}

// pathID parses the ID segment of a path, e.g. "3" in /api/users/3.
func pathID(path []string, i int) (int, bool) {
	if len(path) <= i {
		return 0, false
	}
	id, err := strconv.Atoi(path[i])
	return id, err == nil
}

// Users

func (f *fakeRedash) addUser(u *fakeUser) *fakeUser {
	f.mu.Lock()
	defer f.mu.Unlock()

	u.ID = f.nextUserID
	f.nextUserID++
	if u.Groups == nil {
		u.Groups = []int{fakeDefaultGroupID}
	}
	if u.AuthType == "" {
		u.AuthType = "password"
	}
	u.CreatedAt, u.UpdatedAt, u.ActiveAt = fakeTimestamp, fakeTimestamp, fakeTimestamp
	f.users[u.ID] = u
	return u
}

func (f *fakeRedash) user(id int) *fakeUser {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.users[id]
}

func (f *fakeRedash) handleUsers(w http.ResponseWriter, r *http.Request, path []string) {
	if len(path) == 0 || path[0] == "" {
		switch r.Method {
		case http.MethodGet:
			f.searchUsers(w, r.URL.Query().Get("q"))
		case http.MethodPost:
			var payload struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			}
			if !decodeBody(w, r, &payload) {
				return
			}
			writeJSON(w, f.addUser(&fakeUser{Name: payload.Name, Email: payload.Email, IsInvitationPending: true}))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	id, ok := pathID(path, 0)
	u := f.user(id)
	if !ok || u == nil {
		notFound(w)
		return
	}

	switch {
	case len(path) == 2 && path[1] == "disable" && r.Method == http.MethodPost:
		f.mu.Lock()
		u.IsDisabled = true
		u.DisabledAt = fakeTimestamp
		f.mu.Unlock()
		writeJSON(w, u)
	case len(path) == 1 && r.Method == http.MethodGet:
		writeJSON(w, u)
	case len(path) == 1 && r.Method == http.MethodPost:
		var payload struct {
			Name     *string `json:"name"`
			Email    *string `json:"email"`
			GroupIDs []int   `json:"group_ids"`
		}
		if !decodeBody(w, r, &payload) {
			return
		}
		f.mu.Lock()
		if payload.Name != nil {
			u.Name = *payload.Name
		}
		if payload.Email != nil {
			u.Email = *payload.Email
		}
		if payload.GroupIDs != nil {
			u.Groups = payload.GroupIDs
		}
		f.mu.Unlock()
		writeJSON(w, u)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// searchUsers mimics /api/users?q=, which matches on name or email and
// returns groups as objects rather than IDs.
func (f *fakeRedash) searchUsers(w http.ResponseWriter, q string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	type groupRef struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	type result struct {
		fakeUser
		Groups []groupRef `json:"groups"`
	}

	results := []result{}
	for id := 1; id < f.nextUserID; id++ {
		u, ok := f.users[id]
		if !ok || !(strings.Contains(u.Email, q) || strings.Contains(u.Name, q)) {
			continue
		}
		res := result{fakeUser: *u, Groups: []groupRef{}}
		for _, g := range u.Groups {
			res.Groups = append(res.Groups, groupRef{ID: g, Name: "group " + strconv.Itoa(g)})
		}
		results = append(results, res)
	}

	writeJSON(w, map[string]interface{}{"count": len(results), "page": 1, "page_size": 25, "results": results})
}

// Groups

func (f *fakeRedash) addGroup(name string) *fakeGroup {
	f.mu.Lock()
	defer f.mu.Unlock()

	g := &fakeGroup{
		ID:          f.nextGroupID,
		Name:        name,
		Type:        "regular",
		Permissions: []string{"create_query", "view_query"},
		CreatedAt:   fakeTimestamp,
	}
	f.nextGroupID++
	f.groups[g.ID] = g
	return g
}

func (f *fakeRedash) group(id int) *fakeGroup {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.groups[id]
}

func (f *fakeRedash) handleGroups(w http.ResponseWriter, r *http.Request, path []string) {
	if len(path) == 0 || path[0] == "" {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var payload struct {
			Name string `json:"name"`
		}
		if !decodeBody(w, r, &payload) {
			return
		}
		writeJSON(w, f.addGroup(payload.Name))
		return
	}

	id, ok := pathID(path, 0)
	g := f.group(id)
	if !ok || g == nil {
		notFound(w)
		return
	}

	if len(path) > 1 && path[1] == "data_sources" {
		f.handleGroupDataSources(w, r, id, path[2:])
		return
	}

	switch r.Method {
	case http.MethodGet:
		writeJSON(w, g)
	case http.MethodPost:
		var payload struct {
			Name string `json:"name"`
		}
		if !decodeBody(w, r, &payload) {
			return
		}
		f.mu.Lock()
		g.Name = payload.Name
		f.mu.Unlock()
		writeJSON(w, g)
	case http.MethodDelete:
		f.mu.Lock()
		delete(f.groups, id)
		f.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (f *fakeRedash) handleGroupDataSources(w http.ResponseWriter, r *http.Request, groupID int, path []string) {
	switch {
	case len(path) == 0 && r.Method == http.MethodPost:
		var payload struct {
			DataSourceID int `json:"data_source_id"`
		}
		if !decodeBody(w, r, &payload) {
			return
		}
		ds := f.dataSource(payload.DataSourceID)
		if ds == nil {
			notFound(w)
			return
		}
		f.mu.Lock()
		ds.Groups[groupID] = false
		f.mu.Unlock()
		writeJSON(w, ds)
	case len(path) == 1 && r.Method == http.MethodDelete:
		dsID, _ := pathID(path, 0)
		ds := f.dataSource(dsID)
		if ds == nil {
			notFound(w)
			return
		}
		f.mu.Lock()
		delete(ds.Groups, groupID)
		f.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// Data sources

func (f *fakeRedash) addDataSource(ds *fakeDataSource) *fakeDataSource {
	f.mu.Lock()
	defer f.mu.Unlock()

	ds.ID = f.nextDataSourceID
	f.nextDataSourceID++
	if ds.Groups == nil {
		ds.Groups = map[int]bool{fakeDefaultGroupID: false}
	}
	if ds.Syntax == "" {
		ds.Syntax = "sql"
	}
	f.dataSources[ds.ID] = ds
	return ds
}

func (f *fakeRedash) dataSource(id int) *fakeDataSource {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.dataSources[id]
}

// writeDataSource masks secret options, as Redash does.
func (f *fakeRedash) writeDataSource(w http.ResponseWriter, ds *fakeDataSource) {
	f.mu.Lock()
	masked := *ds
	masked.Options = map[string]interface{}{}
	for k, v := range ds.Options {
		if k == "password" {
			v = redashSecretPlaceholder
		}
		masked.Options[k] = v
	}
	f.mu.Unlock()

	writeJSON(w, masked)
}

func (f *fakeRedash) handleDataSources(w http.ResponseWriter, r *http.Request, path []string) {
	if len(path) == 1 && path[0] == "types" && r.Method == http.MethodGet {
		_, _ = io.WriteString(w, fakeDataSourceTypes)
		return
	}

	if len(path) == 0 || path[0] == "" {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		ds := &fakeDataSource{}
		if !decodeBody(w, r, ds) {
			return
		}
		f.writeDataSource(w, f.addDataSource(ds))
		return
	}

	id, ok := pathID(path, 0)
	ds := f.dataSource(id)
	if !ok || ds == nil {
		notFound(w)
		return
	}

	switch r.Method {
	case http.MethodGet:
		f.writeDataSource(w, ds)
	case http.MethodPost:
		update := &fakeDataSource{}
		if !decodeBody(w, r, update) {
			return
		}
		f.mu.Lock()
		update.ID, update.Groups = ds.ID, ds.Groups
		if update.Syntax == "" {
			update.Syntax = ds.Syntax
		}
		f.dataSources[id] = update
		f.mu.Unlock()
		f.writeDataSource(w, update)
	case http.MethodDelete:
		f.mu.Lock()
		delete(f.dataSources, id)
		f.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
