package db

// TODO lock this shit down
type voteDB map[RoomId]Room

type Room struct {
	VotesVisible bool
	SessionUserMap
}

type SessionUserMap map[SessionId]User

type User struct {
	Name string
	Vote string
}

type RoomId string
type SessionId string

var VoteStore voteDB

func init() {
	VoteStore = make(voteDB)
}

func (db voteDB) GetRoom(room RoomId) Room {
	if db == nil {
		return Room{}
	} else if _, exists := db[room]; !exists {
		return Room{}
	}
	return db[room]
}

func (db voteDB) GetVoteBySession(room RoomId, sessId SessionId) string {
	if db == nil {
		return ""
	} else if _, exists := db[room]; !exists {
		return ""
	} else {
		return db[room].SessionUserMap[sessId].Vote
	}
}

func (db voteDB) ClearRoomVotes(room RoomId) {
	if db == nil {
		return
	}
	if _, exists := db[room]; !exists {
		return
	}
	sessVotes := db[room].SessionUserMap
	for sessId, _ := range sessVotes {
		s := sessVotes[sessId]
		s.Vote = ""
		sessVotes[sessId] = s
	}
	r := db[room]
	r.SessionUserMap = sessVotes
	db[room] = r
}

func (db voteDB) SetRoomVoteVisibility(room RoomId, show bool) {
	if db == nil {
		db = make(voteDB)
	}
	if _, exists := db[room]; !exists {
		db[room] = Room{
			VotesVisible:   show,
			SessionUserMap: make(map[SessionId]User),
		}
		return
	}
	r := db[room]
	r.VotesVisible = show
	db[room] = r
}

func (db voteDB) SetVoteBySession(room RoomId, sessId SessionId, vote string) {
	if db == nil {
		db = make(voteDB)
	}
	if _, exists := db[room]; !exists {
		db[room] = Room{
			VotesVisible:   false,
			SessionUserMap: make(map[SessionId]User),
		}
	}
	user := db[room].SessionUserMap[sessId]
	user.Vote = vote

	db[room].SessionUserMap[sessId] = user
}

// TODO I need to combine this and setVote, too much overlap
func (db voteDB) SetUsernameBySession(room RoomId, sessId SessionId, name string) {
	if db == nil {
		db = make(voteDB)
	}
	if _, exists := db[room]; !exists {
		db[room] = Room{
			VotesVisible:   false,
			SessionUserMap: make(map[SessionId]User),
		}
	}
	user := db[room].SessionUserMap[sessId]
	user.Name = name

	db[room].SessionUserMap[sessId] = user
}
