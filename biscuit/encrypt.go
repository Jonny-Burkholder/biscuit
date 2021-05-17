package biscuit

import (
	"crypto/sha512"
	"fmt"
	"net/http"
)

/*This part of the package is under construction. Note that stronger security features are on the way*/

//this file is for security measures like encrypting cookies and checking ip addresses

type errorUnknownIP struct {
	IP string
}

//Error returns the content body of the errorUnknownIP error type
func (err errorUnknownIP) Error() string {
	return "Unknown IP address: " + err.IP + " does not match user address."
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
	for _, ip := range sess.ipAddress {
		if rAddress == ip {
			return nil
		}
	}
	return errorUnknownIP{IP: rAddress}
}

//Hash returns a hash from an input string based on the session manager's encryption type
func (mng *sessionManager) Hash(s string) ([64]byte, error) {
	switch mng.encryptionType {
	case "sha512":
		return sha512.Sum512([]byte(s)), nil
	default:
		return [64]byte{}, fmt.Errorf("Error hashing data: %q not supported.", s)
	}
}
