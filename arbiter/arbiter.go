package arbiter

import (
	"log"
	"net/http"
	"os"
)

/*Arbiter is a subpackage for logging, security, and middleware*/

type sessionManager interface {
	CheckRole(roles []string, id string) error
	VerifySession(id string) error
}

var logpath string = "../internal/log"

//SetLogPath changes where the
func SetLogging(path string) error {
	//is there a way to check to make sure the log path is valid? Maybe try that, return error if invalid
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	log.SetOutput(f)
	return nil
}

//DefaultLogging sets logging with the defaul logpath
func DefaultLogging() error {
	return SetLogging(logpath)
}

//ConsoleLogging sets logging to os.Stdout, rather than to a log file. Alternatively, a user can simply
//choose not to set logging
func ConsoleLogging() {
	log.SetOutput(os.Stdout)
}

//CheckLog reads an error and logs if non-nil
func CheckLog(err error) {
	if err != nil {
		log.Println(err)
	}
}

//CheckFatal reads an error and stops the program if non-nil
func CheckFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

//CheckPanic reads an error and panics if non-nil
func CheckPanic(err error) {
	if err != nil {
		log.Panic(err)
	}
}

//Restricted wraps a handler to check authentication before allowing a user to access the page, and
//returns http.Error() if the user is unauthorized. User may decide on a case-to-case basis which
//roles are able to access each page
func Restricted(next http.Handler, mng sessionManager, cookieID string, roles ...string) http.Handler { //this needs to be updated to make role variadic
	fn := func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(cookieID)
		if err != nil {
			log.Println(err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if err := mng.VerifySession(cookie.Value); err != nil {
			log.Println(err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if len(roles) == 0 {
			goto serving
		}

		err = mng.CheckRole(roles, cookie.Value)
		if err != nil {
			log.Println(err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

	serving:

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

//ValidateSession checks for a valid biscuit session cookie. It returns
//401 if no such cookie is found
func ValidateSession(next http.Handler, mng sessionManager, cookieID string) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(cookieID)
		if err != nil {
			log.Println(err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}
		if err := mng.VerifySession(cookie.Value); err != nil {
			log.Println(err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

//ValidateRedirect checks for a valid biscuit session cookie. It redirects to
//a given endpoint if no such cookie is found
func ValidateRedirect(next http.Handler, mng sessionManager, cookieID, redirect string) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(cookieID)
		if err != nil {
			log.Println(err)
			http.Redirect(w, r, redirect, http.StatusSeeOther)
			return
		}
		if err := mng.VerifySession(cookie.Value); err != nil {
			log.Println(err)
			http.Redirect(w, r, redirect, http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
