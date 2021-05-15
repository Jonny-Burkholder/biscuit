package biscuit

import (
	"net/http"
	"time"
)

func (mng *sessionManager) setSessionCookie(w http.ResponseWriter, sess *session) error { //I can't think of any errors to return, but I'm sure I need to return one
	cookie := http.Cookie{
		Name:     sessionCookieName + mng.ID,
		Value:    sess.CookieID,
		MaxAge:   sessionLength,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
	return nil
}

func (mng *sessionManager) deleteSessionCookie(w http.ResponseWriter, c *http.Cookie) error { //see setSessionCookie
	c.Expires = time.Now()
	c.MaxAge = -1
	http.SetCookie(w, c)
	return nil
}
