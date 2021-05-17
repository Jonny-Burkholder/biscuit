package biscuit

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

var sessionCookieName string = "SESSbsct"

var defaultSessionLength int //should be set by user for each session manager, but if not, sessions will by default end when the browser is closed

var defautlLockoutTime int = 60 * 5 //by default locks user out for 5 minutes

var defaultMaxLoginAttempts int = 5

var defaultEncryptionType string = "sha512"

var availableEncryptionTypes = []string{"sha512"} //I'll add more later. I have to decide which ones I'll allow

var overseer map[string]*sessionManager

//user is a generic interface to interact with 3rd party user types
type user interface {
	Save() error //a user must have a method to be saved to disk
}

//Session holds information about a user session
type session struct {
	mux       *sync.Mutex
	username  string
	cookieID  string
	ipAddress []string
	alive     bool
	locked    bool
	counter   *counter
}

//counter keeps track of login attempts and locks the user out if there are too many attempts
type counter struct {
	mux      *sync.Mutex
	attempts int
}

//session manager keeps sessions alive in main
type sessionManager struct {
	mux                  *sync.Mutex
	id                   string
	unlockChan           chan string
	killChan             chan bool
	sessions             map[string]*session
	users                map[string]user
	data                 map[string]interface{} //for any data a program might need beyond sessions and users
	sessionLength        int
	maxUserLoginAttempts int
	userLockoutTime      int
	encryptionType       string
}

func NewSessionManager() *sessionManager {
	var mux *sync.Mutex
	id := newMngID()
	c := make(chan string)
	mng := &sessionManager{
		mux:                  mux,
		id:                   id,
		unlockChan:           c,
		sessions:             make(map[string]*session),
		users:                make(map[string]user),
		data:                 make(map[string]interface{}),
		sessionLength:        defaultSessionLength,
		maxUserLoginAttempts: defaultMaxLoginAttempts,
		userLockoutTime:      defautlLockoutTime,
		encryptionType:       defaultEncryptionType,
	}
	mng.run()
	return mng
}

//run() allows the session manager to listen on its channels
func (mng *sessionManager) run() {
	go func() {
		defer close(mng.unlockChan)
		defer close(mng.killChan)
		for {
			select {
			case id := <-mng.unlockChan:
				mng.sessions[id].locked = false
			case <-mng.killChan:
				return
			}
		}
	}()
}

func (mng *sessionManager) SetSessionLength(i int) {
	mng.sessionLength = i
}

func (mng *sessionManager) SetMaxAttempts(i int) {
	mng.maxUserLoginAttempts = i
}

func (mng *sessionManager) SetLockoutTime(i int) {
	mng.userLockoutTime = i
}

//NewSession generates a new session and adds it to the manager
func (mng *sessionManager) NewSession(user string, r *http.Request) (string, error) {
	_, ok := mng.users[user]
	if ok != false {
		return "", fmt.Errorf("Invalid: user session %q already in progress", user)
	}
	var mux *sync.Mutex
	id := mng.newSessionID()
	mng.sessions[id] = &session{
		mux:       mux,
		username:  user,
		cookieID:  id,
		ipAddress: []string{getIP(r)},
		alive:     false, //default to false, mostly to keep track of invalid login attempts
		locked:    false,
		counter:   newCounter(),
	}
	return id, nil
}

func newCounter() *counter {
	var mux sync.Mutex
	return &counter{mux: &mux}
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
		_, ok := mng.sessions[id]
		if ok != true {
			return id
		}
	}
}

//Login changes the bool in a user session so that the manager views the sessin as being "alive", or active
func (mng *sessionManager) Login(id string) error {
	sess, ok := mng.sessions[id]
	if ok != true {
		return fmt.Errorf("Error: session ID not found\n%q", id)
	}
	if sess.alive != false {
		return fmt.Errorf("Error: user %q already logged in.", sess.username)
	}
	sess.alive = true
	return nil
}

//Logout changes the session "alive" bool to false
func (mng *sessionManager) Logout(id string) error {
	sess, ok := mng.sessions[id]
	if ok != true {
		return fmt.Errorf("Error: session ID not found\n%q", id)
	}
	if sess.alive != true {
		return fmt.Errorf("Error: user %q is not logged in.", sess.username)
	}
	sess.alive = false
	return nil
}

func (mng *sessionManager) CountUp(sess *session) error {
	sess.counter.attempts++
	if sess.counter.attempts == mng.maxUserLoginAttempts {
		sess.locked = true
		mng.lockout(sess)
		return fmt.Errorf("Max attempts reached, locked out for %v minutes", mng.userLockoutTime/60)
	}
	return nil
}

func (mng *sessionManager) lockout(sess *session) {
	go func() {
		time.Sleep(time.Second * time.Duration(mng.userLockoutTime))
		mng.unlockChan <- sess.cookieID
	}()
}
