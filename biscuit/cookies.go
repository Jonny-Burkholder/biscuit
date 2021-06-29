package biscuit

import (
	"fmt"
	"net/http"
	"time"
)

//SetSessionCookie sets a cookie in the browser containing the user's unique session ID
func (mng *sessionManager) SetSessionCookie(w http.ResponseWriter, id string) error { //I can't think of any errors to return, but I'm sure I need to return one
	sess := mng.sessions[id]
	cookie := http.Cookie{
		Name:     sessionCookieName + mng.id,
		Value:    sess.cookieID,
		MaxAge:   mng.sessionLength,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
	return nil
}

//SetPerformanceCookie adds a cookie to the browser that lives indefinitely
func (mng *sessionManager) SetPerformanceCookie(w http.ResponseWriter, data []byte) {
	cookie := http.Cookie{
		Name:     performanceCookieName + mng.id,
		Value:    fmt.Sprint(data),
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
}

//SetPreferencesCookie adds a cookie that stores user preferences in browser for JS to use.
//For simplicity's sake, this is passed as a slice of bytes that must be decoded in browser
func (mng *sessionManager) SetPreferencesCookie(w http.ResponseWriter, pref []byte) {
	cookie := http.Cookie{
		Name:     preferenceCookieName + mng.id,
		Value:    fmt.Sprint(pref),
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
}

//DeleteCookie sets a cookie to expire immediately. This is the function to be used for deleting
//all types of cookies in biscuit
func (mng *sessionManager) DeleteCookie(w http.ResponseWriter, c *http.Cookie) error { //see setSessionCookie
	c.Expires = time.Now()
	c.MaxAge = -1
	http.SetCookie(w, c)
	return nil
}

//SessionCookie returns the var sessionCookieName
func SessionCookie() string {
	return sessionCookieName
}

//PreferenceCookie returns the var prefCookieName
func PreferenceCookie() string {
	return preferenceCookieName
}

//PerformanceCookie returns the var perfCookieName
func PerformanceCookie() string {
	return performanceCookieName
}
