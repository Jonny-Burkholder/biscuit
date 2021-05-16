package biscuit

import (
	"net/http"
)

/*This part of the package is under construction. Note that stronger security features are on the way*/

//this file is for security measures like encrypting cookies and checking ip addresses

type errorUnknownIP struct {
	IP string
}

func (err errorUnknownIP) Error() string {
	return "Unknown IP address: " + err.IP + " does not match user address."
}

func getIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

func (mng *sessionManager) VallidateIP(r *http.Request, sess *session) error {
	rAddress := getIP(r)
	if rAddress != sess.IPAddress {
		return errorUnknownIP{IP: rAddress}
	}
	return nil
}
