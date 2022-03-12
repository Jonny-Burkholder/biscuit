package biscuit

import (
	"crypto/md5"
	"crypto/sha512"
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

/*This part of the package is under construction. Note that stronger security features and more robust
errors are on the way*/

//this file is for security measures like encrypting cookies and checking ip addresses

type errorUnauthorizedIP struct {
	IP string
}

//Error returns the content body of the errorUnauthorizedIP error type
func (err errorUnauthorizedIP) Error() string {
	return "Unauthorized IP address: " + err.IP + " does not match user address."
}

//getIP returns the client's IP address
func getIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

//ValidateIP returns an error if a session cookie comes from an IP address other than
//what is stored in the session manager
func (mng *sessionManager) ValidateIP(r *http.Request, sess *session) error {
	rAddress := getIP(r)
	ip, ok := sess.ipAddress[rAddress]
	if ok != true {
		mng.addIP(sess, rAddress)
	}
	if ip != true {
		return errorUnauthorizedIP{IP: rAddress}
	}
	return nil
}

//Hash returns a hash from an input string based on the session manager's encryption type
func (mng *sessionManager) Hash(s string) ([]byte, error) {
	switch mng.encryptionType {
	case "bcrypt":
		hash, err := bcrypt.GenerateFromPassword([]byte(s), mng.hashStrength)
		if err != nil {
			return []byte{}, err
		}
		return hash, nil
	case "md5": //non b-crypt hashing will be updated in the future to incorporate hashing rounds as well as salting
		//initial hash
		hash := md5.Sum([]byte(s))

		//re-hash according to manager hash strength
		for i := 1; i < mng.hashStrength; i++ {
			hash = md5.Sum(hash[:])
		}

		return hash[:], nil
	case "sha512":
		//initial hash
		hash := sha512.Sum512([]byte(s))

		//re-hash for strength and stuff
		for i := 1; i < mng.hashStrength; i++ {
			hash = sha512.Sum512(hash[:])
		}

		return hash[:], nil
	default:
		return []byte{}, fmt.Errorf("Error hashing data: %q not supported.", s)
	}
}

//CheckPassword takes a password string and compares it to a hash to see if they match.
//The function returns an error if they do not match, and nil if they do match
func (mng *sessionManager) CheckPassword(pswd string, hash []byte) error {
	if mng.encryptionType == "bcrypt" {
		return bcrypt.CompareHashAndPassword(hash, []byte(pswd))
	} else {
		return fmt.Errorf("Sorry, still working on anything that's not bcrypt")
	}
}
