package user

import (
	"context"
	"errors"
	"log"
	"math/rand"

	"github.com/Nahemah1022/singsphere-backend/playlist"
	"github.com/Nahemah1022/singsphere-backend/stereo"
	"github.com/pion/webrtc/v2"
)

type broadcastMsg struct {
	data []byte
	user *User // message will be broadcasted to everyone, except this user
}

// Room maintains the set of active clients and broadcasts messages to the
// clients.
type Room struct {
	Name            string
	users           map[string]*User
	broadcast       chan broadcastMsg
	join            chan *User // Register requests from the clients.
	leave           chan *User // Unregister requests from clients.
	stereoTrack     *webrtc.Track
	stereoPlaying   bool
	requests        chan string       // Songs requested from users
	transcodedSongs chan *stereo.Song // Songs transcoded into OPUS after user requests
	playlist        *playlist.Playlist
}

// RoomWrap is a public representation of a room
type RoomWrap struct {
	Users  []*UserWrap `json:"users"`
	Name   string      `json:"name"`
	Online int         `json:"online"`
}

// Wrap returns public version of room
func (r *Room) Wrap(me *User) *RoomWrap {
	usersWrap := []*UserWrap{}
	for _, user := range r.GetUsers() {
		if me != nil {
			// do not add current user to room
			if me.ID == user.ID {
				continue
			}
		}
		usersWrap = append(usersWrap, user.Wrap())
	}

	return &RoomWrap{
		Users:  usersWrap,
		Name:   r.Name,
		Online: len(usersWrap),
	}
}

// NewRoom creates new room
func NewRoom(name string) *Room {
	// mediaEngine := webrtc.MediaEngine{}
	// mediaEngine.RegisterCodec(webrtc.NewRTPOpusCodec(webrtc.DefaultPayloadTypeOpus, 48000))
	// iceConnectedCtx, iceConnectedCtxCancel := context.WithCancel(context.Background())

	requests := make(chan string)

	pl := playlist.New()
	pl.Subscribe(name, requests)

	audioTrack, addTrackErr := webrtc.NewTrack(
		webrtc.DefaultPayloadTypeOpus,
		rand.Uint32(),
		"stereo_audio",
		"pion",
		webrtc.NewRTPOpusCodec(webrtc.DefaultPayloadTypeOpus, 48000),
	)
	if addTrackErr != nil {
		panic(addTrackErr)
	}

	return &Room{
		broadcast:       make(chan broadcastMsg),
		join:            make(chan *User),
		leave:           make(chan *User),
		users:           make(map[string]*User),
		Name:            name,
		stereoTrack:     audioTrack,
		stereoPlaying:   false,
		requests:        requests,
		transcodedSongs: make(chan *stereo.Song),
		playlist:        pl,
	}
}

func (r *Room) StereoPlay() {
	// ensure this goroutine invoked only once when the first user comes in
	if r.stereoPlaying {
		return
	}
	r.stereoPlaying = true

	log.Println("Stereo Start Playing")
	for song := range r.transcodedSongs {
		log.Printf("Playing Song: %s\n", song.Name)
		r.BroadcastSong("next_song", song)
		ctx, ctxCancel := context.WithCancel(context.Background())
		stereo.Play(song.Path, r.stereoTrack, ctxCancel)
		<-ctx.Done()
	}
}

// GetUsers converts map[int64]*User to list
func (r *Room) GetUsers() []*User {
	users := []*User{}
	for _, user := range r.users {
		users = append(users, user)
	}
	return users
}

// GetOtherUsers returns other users of room except current
func (r *Room) GetOtherUsers(user *User) []*User {
	users := []*User{}
	for _, userCandidate := range r.users {
		if user.ID == userCandidate.ID {
			continue
		}
		users = append(users, userCandidate)
	}
	return users
}

// Join connects user and room
func (r *Room) Join(user *User) {
	r.join <- user
}

// Leave disconnects user and room
func (r *Room) Leave(user *User) {
	r.leave <- user
}

// Broadcast sends message to everyone except user (if passed)
func (r *Room) Broadcast(data []byte, user *User) {
	message := broadcastMsg{data: data, user: user}
	r.broadcast <- message
}

// GetUsersCount return users count in the room
func (r *Room) GetUsersCount() int {
	return len(r.GetUsers())
}

// Broadcast enqueue/nextsong event to room's users
func (r *Room) BroadcastSong(operation string, song *stereo.Song) {
	for _, u := range r.users {
		if err := u.SendEvent(Event{Type: operation, Song: song}); err != nil {
			panic(err)
		}
	}
}

func (r *Room) run() {
	for {
		select {
		case user := <-r.join:
			r.users[user.ID] = user
			go user.BroadcastEventJoin()
		case user := <-r.leave:
			if _, ok := r.users[user.ID]; ok {
				delete(r.users, user.ID)
				close(user.send)
			}
			go user.BroadcastEventLeave()
		case message := <-r.broadcast:
			for _, user := range r.users {
				// message will be broadcasted to everyone, except this user
				if message.user != nil && user.ID == message.user.ID {
					continue
				}
				select {
				case user.send <- message.data:
				default:
					close(user.send)
					delete(r.users, user.ID)
				}
			}
		case encoded := <-r.requests:
			// the transcoded songs won't be played immediately, but we still need
			// an immediate notification from transcoder about whether the song is
			// successfully enqueued (might fail to transcode or doesn't exist)
			// therefore, here we use this additional channel to receive the notification
			successCh := make(chan *stereo.Song)
			go stereo.Trans(encoded, r.transcodedSongs, successCh)
			song := <-successCh
			if song != nil {
				r.BroadcastSong("enqueue", song)
			}
			close(successCh)
		}
	}
}

// Rooms is a set of rooms
type Rooms struct {
	rooms map[string]*Room
}

var ErrNotFound = errors.New("not found")

// Get room by room id
func (r *Rooms) Get(roomID string) (*Room, error) {
	if room, exists := r.rooms[roomID]; exists {
		return room, nil
	}
	return nil, ErrNotFound
}

// GetOrCreate creates room if it does not exist
func (r *Rooms) GetOrCreate(roomID string) *Room {
	room, err := r.Get(roomID)
	if err == nil {
		return room
	}
	newRoom := NewRoom(roomID)
	r.AddRoom(roomID, newRoom)
	go newRoom.run()
	return newRoom

}

// AddRoom adds room to rooms list
func (r *Rooms) AddRoom(roomID string, room *Room) error {
	if _, exists := r.rooms[roomID]; exists {
		return errors.New("room with id " + roomID + " already exists")
	}
	r.rooms[roomID] = room
	return nil
}

// RemoveRoom remove room from rooms list
func (r *Rooms) RemoveRoom(roomID string) error {
	if _, exists := r.rooms[roomID]; exists {
		delete(r.rooms, roomID)
		return nil
	}
	return nil
}

// RoomsStats is an app global statistics
type RoomsStats struct {
	Online int         `json:"online"`
	Rooms  []*RoomWrap `json:"rooms"`
}

// GetStats get app statistics
func (r *Rooms) GetStats() RoomsStats {
	stats := RoomsStats{
		Rooms: []*RoomWrap{},
	}
	for _, room := range r.rooms {
		stats.Online += room.GetUsersCount()
		stats.Rooms = append(stats.Rooms, room.Wrap(nil))
	}
	return stats
}

// NewRooms creates rooms instance
func NewRooms() *Rooms {
	return &Rooms{
		rooms: make(map[string]*Room, 100),
	}
}
