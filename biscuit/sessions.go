package biscuit

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

var sessionCookieName string = "SESSbsct"

var sessionLength int //should be set by user, but if not, sessions will by default end when the browser is closed

var overseer map[string]*sessionManager

//user is a generic interface to interact with 3rd party user types
type user interface {
	Save() error //a user must have a method to be saved to disk
}

//Session holds information about a user session
type session struct {
	Mux       *sync.Mutex
	LockChan  chan bool
	Username  string
	CookieID  string
	IPAddress string
	Alive     bool
	locked    bool
	Counter   *counter
}

//counter keeps track of login attempts and locks the user out if there are too many attempts
type counter struct {
	Mux         *sync.Mutex
	MaxAttempts int
	Attempts    int
	ResetTime   int
}

//session manager keeps sessions alive in main
type sessionManager struct {
	Mux        *sync.Mutex
	ID         string
	UnlockChan chan string
	KillChan   chan bool
	Sessions   map[string]*session
	Users      map[string]user
	Data       map[string]interface{} //for any data a program might need beyond sessions and users
}

func NewSessionManager() *sessionManager {
	var mux *sync.Mutex
	id := newMngID()
	c := make(chan string)
	mng := &sessionManager{
		Mux:        mux,
		ID:         id,
		UnlockChan: c,
		Sessions:   make(map[string]*session),
		Users:      make(map[string]user),
		Data:       make(map[string]interface{}),
	}
	mng.run()
	return mng
}

//run() allows the session manager to listen on its channels
func (mng *sessionManager) Run() {
	go func() {
		defer close(mng.UnlockChan)
		defer close(mng.KillChan)
		for {
			select {
			case id := <-mng.UnlockChan:
				mng.Sessions[id].locked = false
			case <-mng.KillChan:
				return
			}
		}
	}()
}

//NewSession generates a new session and adds it to the manager
func (mng *sessionManager) NewSession(user string, r *http.Request) error {
	_, ok := mng.Users[user]
	if ok != false {
		return fmt.Errorf("Invalid: user session %q already in progress", user)
	}
	var mux *sync.Mutex
	id := mng.newSessionID()
	mng.Sessions[id] = &session{
		Mux:       mux,
		Username:  user,
		CookieID:  id,
		IPAddress: getIP(r),
		Alive:     false, //default to false, in case of something like
	}
	return nil
}

//generates a new, unique ID for a session manager
func newMngID() string {
	for {
		rand.Seed(time.Now().Unix())
		id := fmt.Sprint(rand.Uint64())
		_, ok := overseer[id]
		if ok != true {
			return id
		}
	}
}

func (mng *sessionManager) newSessionID() string {
	for {
		rand.Seed(time.Now().Unix())
		id := fmt.Sprint(rand.Uint64())
		_, ok := mng.Sessions[id]
		if ok != true {
			return id
		}
	}
}

//Login changes the bool in a user session so that the manager views the sessin as being "alive", or active
func (mng *sessionManager) Login(id string) error {
	sess, ok := mng.Sessions[id]
	if ok != true {
		return fmt.Errorf("Error: session ID not found\n%q", id)
	}
	if sess.Alive != false {
		return fmt.Errorf("Error: user %q already logged in.", sess.Username)
	}
	sess.Alive = true
	return nil
}

//Logout changes the session "alive" bool to false
func (mng *sessionManager) Logout(id string) error {
	sess, ok := mng.Sessions[id]
	if ok != true {
		return fmt.Errorf("Error: session ID not found\n%q", id)
	}
	if sess.Alive != true {
		return fmt.Errorf("Error: user %q is not logged in.", sess.Username)
	}
	sess.Alive = false
	return nil
}

func (mng *sessionManager) CountUp(sess *session) error {
	sess.Counter.Attempts++
	if sess.Counter.Attempts == sess.Counter.MaxAttempts {
		sess.locked = true
		mng.lockout(sess)
		return fmt.Errorf("Max attempts reached, locked out for %v minutes", sess.Counter.ResetTime/60)
	}
	return nil
}

func (mng *sessionManager) lockout(sess *session) {
	go func() {
		time.Sleep(time.Second * time.Duration(sess.Counter.ResetTime))
		mng.UnlockChan <- sess.CookieID
	}()
}
