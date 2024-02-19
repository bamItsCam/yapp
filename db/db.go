package db

import (
	"sort"
	"sync"
	"time"
)

type VoteDB struct {
	db    map[RoomId]Room
	mutex *sync.RWMutex
}

type Room struct {
	VotesVisible bool
	sessionUserMap
	lastUsed time.Time
}

func (r Room) Users() []User {
	// good enough
	keys := make([]string, 0, len(r.sessionUserMap))
	users := make([]User, 0, len(sessionUserMap{}))
	for k := range r.sessionUserMap {
		keys = append(keys, string(k))
	}

	sort.Strings(keys)

	for _, k := range keys {
		users = append(users, r.sessionUserMap[SessionId(k)])
	}

	return users
}

func NewRoom(votes bool, lastUsed time.Time) Room {
	return Room{
		VotesVisible:   votes,
		sessionUserMap: make(map[SessionId]User),
		lastUsed:       lastUsed,
	}
}

type sessionUserMap map[SessionId]User

type User struct {
	Name string
	Vote string
}

type RoomId string
type SessionId string

var VoteStore VoteDB

func init() {
	VoteStore = VoteDB{
		db:    make(map[RoomId]Room),
		mutex: &sync.RWMutex{},
	}
}

func (vdb VoteDB) GetRoom(room RoomId) Room {
	vdb.mutex.Lock()
	defer vdb.mutex.Unlock()

	if r, exists := vdb.db[room]; !exists {
		return Room{}
	} else if r.lastUsed.Before(time.Now().Add(-1 * time.Hour)) {
		vdb.db[room] = NewRoom(false, time.Now())
	}

	r := vdb.db[room]
	r.lastUsed = time.Now()
	vdb.db[room] = r
	return r
}

func (vdb VoteDB) GetVoteBySession(room RoomId, sessId SessionId) string {
	vdb.mutex.RLock()
	defer vdb.mutex.RUnlock()

	if _, exists := vdb.db[room]; !exists {
		return ""
	} else {
		return vdb.db[room].sessionUserMap[sessId].Vote
	}
}

func (vdb VoteDB) ClearRoomVotes(room RoomId) {
	vdb.mutex.Lock()
	defer vdb.mutex.Unlock()

	if _, exists := vdb.db[room]; !exists {
		return
	}
	sessVotes := vdb.db[room].sessionUserMap
	for sessId, _ := range sessVotes {
		s := sessVotes[sessId]
		s.Vote = ""
		sessVotes[sessId] = s
	}
	r := vdb.db[room]
	r.sessionUserMap = sessVotes
	vdb.db[room] = r
}

func (vdb VoteDB) SetRoomVoteVisibility(room RoomId, show bool) {
	vdb.mutex.Lock()
	defer vdb.mutex.Unlock()

	if _, exists := vdb.db[room]; !exists {
		vdb.db[room] = NewRoom(show, time.Now())
		return
	}
	r := vdb.db[room]
	r.VotesVisible = show
	vdb.db[room] = r
}

func (vdb VoteDB) SetVoteBySession(room RoomId, sessId SessionId, vote string) {
	vdb.mutex.Lock()
	defer vdb.mutex.Unlock()

	if _, exists := vdb.db[room]; !exists {
		vdb.db[room] = NewRoom(false, time.Now())
	}
	user := vdb.db[room].sessionUserMap[sessId]
	user.Vote = vote

	vdb.db[room].sessionUserMap[sessId] = user
}

// TODO I need to combine this and setVote, too much overlap
func (vdb VoteDB) SetUsernameBySession(room RoomId, sessId SessionId, name string) {
	vdb.mutex.Lock()
	defer vdb.mutex.Unlock()

	if _, exists := vdb.db[room]; !exists {
		vdb.db[room] = NewRoom(false, time.Now())
	}
	user := vdb.db[room].sessionUserMap[sessId]
	user.Name = name

	vdb.db[room].sessionUserMap[sessId] = user
}
