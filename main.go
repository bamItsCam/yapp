package main

import (
	"bytes"
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	sse "github.com/r3labs/sse/v2"
	"log"
	"net/http"
	"yapp/components"
	"yapp/db"
)

var sessionCookie = "sess_id"

var eventServer *sse.Server

func init() {
	eventServer = sse.New()
}

func main() {
	mux := chi.NewRouter()
	mux.Use(middleware.Logger)
	mux.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := getSessionId(r)
			if id == "" {
				id = uuid.New().String()
				http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: id, Secure: true, HttpOnly: true})
			}
			next.ServeHTTP(w, r)
		})
	})

	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		go func() {
			// Received Browser Disconnection
			<-r.Context().Done()
			log.Println("The client is disconnected here")
			return
		}()

		eventServer.ServeHTTP(w, r)
	})

	mux.Group(func(html chi.Router) {
		html.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html")
				next.ServeHTTP(w, r)
			})
		})
		html.Post("/{room:[0-9]+}/point", setPoint)
		html.Post("/{room:[0-9]+}/user", setUser)
		html.Post("/{room:[0-9]+}/show", showVotes)
		html.Post("/{room:[0-9]+}/hide", hideVotes)
		html.Post("/{room:[0-9]+}/hide", clearVotes)

		html.Get("/{room:[0-9]+}", pageRoom)
	})

	http.ListenAndServe(":3000", mux)
}

func pageRoom(w http.ResponseWriter, r *http.Request) {
	room := chi.URLParam(r, "room")
	eventServer.CreateStream(room)
	ctx := context.WithValue(r.Context(), "room", room)
	if err := components.Room(db.VoteStore.GetVoteBySession(db.RoomId(room), db.SessionId(getSessionId(r))), db.VoteStore.GetRoom(db.RoomId(room))).Render(ctx, w); err != nil {
		log.Println(err)
	}
}

func setPoint(w http.ResponseWriter, r *http.Request) {
	var point string
	room := chi.URLParam(r, "room")
	ctx := context.WithValue(r.Context(), "room", room)

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else if points := r.Form["point"]; len(points) == 1 {
		point = points[0]
	} // todo validation

	db.VoteStore.SetVoteBySession(db.RoomId(room), db.SessionId(getSessionId(r)), point)

	publishVoteTableUpdateMsg(ctx, db.RoomId(room))

	if err := components.Pointer(point).Render(ctx, w); err != nil {
		log.Println(err)
	}
}

func setUser(w http.ResponseWriter, r *http.Request) {
	var username string
	room := chi.URLParam(r, "room")

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else if usernames := r.Form["username"]; len(usernames) == 1 {
		username = usernames[0]
	} //todo validation

	db.VoteStore.SetUsernameBySession(db.RoomId(room), db.SessionId(getSessionId(r)), username)
	// todo should clear user's prior vote?
	ctx := context.WithValue(r.Context(), "room", room)
	if err := components.UsernameDisplay(username).Render(ctx, w); err != nil {
		log.Println(err)
	}
	// todo setting name should also send sse to update point table
	if err := components.Pointer(db.VoteStore.GetVoteBySession(db.RoomId(room), db.SessionId(getSessionId(r)))).Render(ctx, w); err != nil {
		log.Println(err)
	}
}

func showVotes(w http.ResponseWriter, r *http.Request) {
	room := chi.URLParam(r, "room")
	ctx := context.WithValue(r.Context(), "room", room)

	db.VoteStore.SetVoteVisibility(db.RoomId(room), true)
	publishVoteTableUpdateMsg(ctx, db.RoomId(room))
}

func hideVotes(w http.ResponseWriter, r *http.Request) {
	room := chi.URLParam(r, "room")
	ctx := context.WithValue(r.Context(), "room", room)

	db.VoteStore.SetVoteVisibility(db.RoomId(room), false)
	publishVoteTableUpdateMsg(ctx, db.RoomId(room))
}

func clearVotes(w http.ResponseWriter, r *http.Request) {
	room := chi.URLParam(r, "room")
	ctx := context.WithValue(r.Context(), "room", room)

	db.VoteStore.SetVoteVisibility(db.RoomId(room), false)
	publishVoteTableUpdateMsg(ctx, db.RoomId(room))
}

func getSessionId(r *http.Request) (id string) {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil {
		return
	}
	return cookie.Value
}

func publishVoteTableUpdateMsg(ctx context.Context, room db.RoomId) {
	buf := new(bytes.Buffer)
	if err := components.RoomVotes(db.VoteStore.GetRoom(room)).Render(ctx, buf); err != nil {
		log.Println(err)
	}
	eventServer.Publish(string(room), &sse.Event{
		Data: buf.Bytes(),
	})
}
