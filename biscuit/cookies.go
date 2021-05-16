package biscuit

import (
	"net/http"
	"time"
)

func (mng *sessionManager) SetSessionCookie(w http.ResponseWriter, id string) error { //I can't think of any errors to return, but I'm sure I need to return one
	sess := mng.Sessions[id]
	cookie := http.Cookie{
		Name:     sessionCookieName + mng.ID,
		Value:    sess.CookieID,
		MaxAge:   mng.SessionLength,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
	return nil
}

func (mng *sessionManager) DeleteSessionCookie(w http.ResponseWriter, r *http.Request) error { //see setSessionCookie
	c, err := r.Cookie(sessionCookieName + mng.ID)
	if err != nil {
		return err
	}
	c.Expires = time.Now()
	c.MaxAge = -1
	http.SetCookie(w, c)
	return nil
}
