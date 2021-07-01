package biscuit

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

var sessionCookieName string = "SESSbsct"

var preferenceCookieName string = "PREFbsct"

var performanceCookieName string = "PERFbsct" //I realize the similarity between "preference" and "performance" is confusing, I'll try to come up with better terms

var defaultSessionLength int //should be set by user for each session manager, but if not, sessions will by default end when the browser is closed

var defautlLockoutTime int = 60 * 5 //by default locks user out for 5 minutes

var defaultMaxLoginAttempts int = 5

var defaultEncryptionType string = "bcrypt"

var defaultHashStrength int = 5

var availableEncryptionTypes = []string{"bcrypt", "sha512"} //I'll add more later. I have to decide which ones I'll allow

var overseer map[string]*sessionManager

//user is a generic interface to interact with 3rd party user types
type user interface {
	CreatePassword(string) error //a user must have a method to create a password that returns an error
}

/*EXAMPLE OF HOW TO IMPLEMENT USER INTERFACE
type myUser struct{
	username string
	password string
}

func (user *myUser) CheckPassword(pswd string) error{
	if user.password != pswd{
		return fmt.Errorf("Passwords do not match")
	}
}
*/

//Session holds information about a user session
type session struct {
	mux       *sync.Mutex
	username  string
	cookieID  string
	ipAddress map[string]bool //false is blocked, while true is allowed
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
	hashStrength         int
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
		hashStrength:         defaultHashStrength,
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

//SetSettionLength determines how long a session lasts in the session manager. The session manager
//will use the same length of time for session cookies as well as the in-memory session handler
func (mng *sessionManager) SetSessionLength(i int) {
	mng.sessionLength = i
}

//SetMaxAttempts determines the maximum number of incorrect password attempts a user has before being
//locked out of their account
func (mng *sessionManager) SetMaxAttempts(i int) {
	mng.maxUserLoginAttempts = i
}

//SetLockoutTime determines, in seconds, how long a user will be locked out of their account for
//reaching the maximimum number of login attempts
func (mng *sessionManager) SetLockoutTime(i int) {
	mng.userLockoutTime = i
}

//SetEncryptionType sets which encryption type will be used by default in the session manager for
//hashing passwords and other data
func (mng *sessionManager) SetEncryptionType(s string) error {
	for _, val := range availableEncryptionTypes {
		if s == val {
			mng.encryptionType = val
		}
	}
	return fmt.Errorf("Error: encryption type %q not supported", s)
}

//SetHashStrength changes the default hash strength for passwords hashed by the session manager.
//Hash strenght must be between 1 and 10
func (mng *sessionManager) SetHashStrength(i int) error {
	if i < 1 {
		return fmt.Errorf("Error: hash strength must be greater than 0.")
	} else if i > 10 {
		i := i % 10
		mng.hashStrength = i
		return fmt.Errorf("Maximum hash strength is 10. Hash strength set to %v.", i)
	}
	mng.hashStrength = i
	return nil
}

//NewSession generates a new session and adds it to the manager
//In next update, session number should be hashed in the session manager, and the unhashed
//session number should be returned
func (mng *sessionManager) NewSession(user string, r *http.Request) (string, error) {
	_, ok := mng.users[user]
	if ok != false {
		return "", fmt.Errorf("Invalid: user session %q already in progress", user)
	}
	var mux *sync.Mutex
	id := mng.newSessionID()
	ip := getIP(r)
	ipMap := make(map[string]bool)
	ipMap[ip] = true
	mng.sessions[id] = &session{
		mux:       mux,
		username:  user,
		cookieID:  id,
		ipAddress: ipMap,
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

//addIP adds a new IP address to the session's ipAddress map, and sets its state to false
func (mng *sessionManager) addIP(sess *session, ip string) {
	sess.ipAddress[ip] = false
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

//allowIP sets the state of a given IP address to true
func (mng *sessionManager) allowIP(sess *session, ip string) {
	sess.ipAddress[ip] = true
}

//BlockIP sets the state of a given IP address to false
func (mng *sessionManager) BlockIP(sess *session, ip string) {
	sess.ipAddress[ip] = false
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

//GetSession allows a user to retrieve a session
func (mng *sessionManager) GetSession(id string) (*session, error) {
	sess, ok := mng.sessions[id]
	if ok != true {
		return &session{}, fmt.Errorf("Session %q not found", id)
	}
	return sess, nil
}

//GetNameFromID takes as input a session ID from the session cookie, and returns the name from
//the user session
func (mng *sessionManager) GetNameFromID(id string) (string, error) {
	sess, err := mng.GetSession(id)
	if err != nil {
		return "", err
	}
	return sess.username, nil
}
