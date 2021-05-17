package biscuit

import (
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

//DeleteSessionCookie sets a session cookie to expire immediately, thus ending the user's session
func (mng *sessionManager) DeleteSessionCookie(w http.ResponseWriter, r *http.Request) error { //see setSessionCookie
	c, err := r.Cookie(sessionCookieName + mng.id)
	if err != nil {
		return err
	}
	c.Expires = time.Now()
	c.MaxAge = -1
	http.SetCookie(w, c)
	return nil
}
