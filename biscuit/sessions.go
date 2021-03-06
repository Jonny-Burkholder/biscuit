package biscuit

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
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

var availableEncryptionTypes = []string{"bcrypt", "sha512", "md5"} //I'll add more later. I have to decide which ones I'll allow

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

//Session holds information about a user session. May depricate certain
//fields in the future, instead using something like data interface{} for
//user or other data
type session struct {
	mux       *sync.Mutex
	username  string //not every session needs a user, need to update this
	role      string
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

//sessionManager is an in-memory struct that keeps track of
//session data
type sessionManager struct {
	mux                  *sync.Mutex
	id                   string //the id is for if you're using multiple session managers, which I don't recommend. I might remove
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

//NewSessionManager is the basis of the user API. It takes no arguments, and
//returns a pointer to a session manager struct. The session manager is automatically
//running on creation
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

//LoadSessionManager takes a string argument "path", which points
//to a json serialized session manager on disk. It then loads this
//session manager into an in-memory struct, returning its pointer.
//Eventually we'll do a better job sanatizing the path string, but
//for now I just want it in place
func LoadSessionManager(path string) (*sessionManager, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	//will have to update this when we move things to
	//
	mng := &sessionManager{}
	err = json.Unmarshal(f, mng)
	if err != nil {
		return nil, err
	}
	return mng, err
}

//Marshal takes a session manager and marshals it to json for DB storage, if you're
//into that sort of thing
func (mng *sessionManager) Marshal() ([]byte, error) {
	//we'll need to manually change this to fill in the non-exported
	//fields for the session manager
	return json.Marshal(mng)
}

//run() allows the session manager to listen asyncronously
//on its various channels and perform tasks with them
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
//This eventually needs to be split into hashing and encryption as separate ideas within
//the session manager, as we'll want to have both available. Also maybe I'll consider using
//an int instead of a string, like with Rainman
func (mng *sessionManager) SetEncryptionType(s string) error {
	for _, val := range availableEncryptionTypes {
		if s == val {
			mng.encryptionType = val
		}
	}
	return fmt.Errorf("Error: encryption type %q not supported", s)
}

//SetHashStrength changes the default hash strength for passwords hashed by the session manager.
//Hash strength must be between 1 and 10. Why 10? Completely arbitrary. I'll update that with
//an informed decision eventually
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
//session number should be returned. Ok I know I wrote that part of the comment, but I'm not
//sure it makes any sense, I don't think that provides any security and it's probably just
//a waste of time
func (mng *sessionManager) NewSession(user string, r *http.Request, role ...string) (string, error) {
	_, ok := mng.users[user]
	if ok != false {
		return "", fmt.Errorf("Invalid: user session %q already in progress", user)
	}

	var userRole string

	if len(role) > 0 {
		userRole = role[0]
	}
	var mux *sync.Mutex
	id := mng.newSessionID()
	ip := getIP(r)
	ipMap := make(map[string]bool)
	ipMap[ip] = true
	mng.sessions[id] = &session{
		mux:       mux,
		username:  user,
		role:      userRole,
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

//newSessionID generates a new random ID for a session
//Currently this is not using what is considered a cryptographically
//secure randomizer. I'll switch this to crypto.Rand()
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

//addIP adds a new IP address to the session's ipAddress map, and sets its state
//to true, indicating that it has not been blocked
func (mng *sessionManager) addIP(sess *session, ip string) {
	sess.ipAddress[ip] = true
}

//Login changes the bool in a user session so that the manager views the session as being "alive", or active
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

//allowIP sets the state of a given IP address to true, indicating
//that is is allowed
func (mng *sessionManager) allowIP(sess *session, ip string) {
	sess.ipAddress[ip] = true
}

//BlockIP sets the state of a given IP address to false, indicating
//that it has been blocked
func (mng *sessionManager) BlockIP(sess *session, ip string) {
	sess.ipAddress[ip] = false
}

//Logout changes the session "alive" bool to false, so that the session
//manager no longer considers the session to be active
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

//CountUp increments the number of login attempts for a session, and locks the
//session if the attempts reaches the maximum attempts allowed by the session manager
func (mng *sessionManager) CountUp(sess *session) error {
	sess.counter.attempts++
	if sess.counter.attempts == mng.maxUserLoginAttempts {
		sess.locked = true
		mng.lockout(sess)
		return fmt.Errorf("Max attempts reached, locked out for %v minutes", mng.userLockoutTime/60)
	}
	return nil
}

//Lockout locks out a user session for the time indicated by the session manager
func (mng *sessionManager) lockout(sess *session) {
	go func() {
		//should probably update this to a ticker
		time.Sleep(time.Second * time.Duration(mng.userLockoutTime))
		mng.unlockChan <- sess.cookieID
	}()
}

//GetSession takes a session id and returns a pointer to a session and an error.
//If the session is not found, a non-nil error will be returned. Typically the user
//is retreiving this session ID from a request cookie value
func (mng *sessionManager) GetSession(id string) (*session, error) {
	sess, ok := mng.sessions[id]
	if ok != true {
		return &session{}, fmt.Errorf("Session %q not found", id)
	}
	return sess, nil
}

//GetNameFromID takes as input a session ID from the session cookie, and returns the name from
//the user session. Should be updated to more generically return user data from the session
func (mng *sessionManager) GetNameFromID(id string) (string, error) {
	sess, err := mng.GetSession(id)
	if err != nil {
		return "", err
	}
	return sess.username, nil
}

//GetRole takes a session ID string argument, and returns the user's
//role if found. If not found, it returns an empty string and a non-
//nil error
func (mng *sessionManager) GetRole(id string) (string, error) {
	user, err := mng.GetSession(id)
	if err != nil {
		return "", err
	}
	return user.role, nil
}

//CheckRole takes a slice of strings, "roles", and a session ID. It checks the session
//to see if the session role matches any of the roles in the slice. This should be updated
//to take a hash map for constant lookup time
func (mng *sessionManager) CheckRole(roles []string, id string) error {
	userRole, err := mng.GetRole(id)
	if err != nil {
		return err
	}
	for _, role := range roles {
		if userRole == role {
			return nil
		}
	}
	return fmt.Errorf("Roles do not matched. Wanted %v, got %v", roles, userRole)
}

//VerifySession takes a session id as an input and returns a non-nil
//error if the session does not exist, or if the session's loggedIn
//field is set to false
func (mng *sessionManager) VerifySession(id string) error {
	sess, err := mng.GetSession(id)
	if err != nil {
		return err
	}
	if sess.alive != true {
		return fmt.Errorf("User %v has a session, but is inactive", id)
	}
	return nil
}

//VerifySessionWithIP, like VerifySession, takes a session ID as input
//and returns a non-nil error if the session does not exist, or if the
//session's loggedIn bool is set to false. In addition, it returns a non-
//nil error if the request IP address has been blocked by the session manager
func (mng *sessionManager) VerifySessionWithIP(id string, r *http.Request) error {
	sess, err := mng.GetSession(id)
	if err != nil {
		return err
	}
	if sess.alive != true {
		return fmt.Errorf("User %v has a session, but is inactive", id)
	}
	return mng.ValidateIP(r, sess)
}
