package types

// I hate this. I don't want a types package. TODO make this not be needed

var PossibleVotes = []string{"0.5", "1", "2", "3", "5", "8", "13", "21", "100", "?"}
var SSETypeRoomUpdate = []byte("room")
var SSETypeClear = []byte("clear")
