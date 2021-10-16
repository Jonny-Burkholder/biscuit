package arbiter

import (
	"log"
	"net/http"
	"os"
)

/*this will be an eventually robust middleware package that will be in fact so robust that it
will split off from this repo and become its own ultra-popular middleware package that will
eventually gain sentience and take over the world. Probably*/

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

//CheckFunc reads an error and excecutes a function if non-nil
func CheckRedirect(err error, w http.ResponseWriter, r *http.Request, s string, i int) {
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, s, i)
		return
	}
}

/* UNFINISHED AND PROBABLY ALWAYS WILL BE
//RestrictedRedirect wraps a handler function and redirects a user to a new page if they do not have
//permission to access the restricted page
func RestrictedRedirect(handler func(w http.ResponseWriter, r *http.Request), redirect func(w http.ResponseWriter, r *http.Request, s string, i int)) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request){
		cookie, err := r.Cookie(biscuit.SessionCookie())
		if err != nil{
			redirect()
		}
	}
}
*/

//Restricted wraps a handler to check authentication before allowing a user to access the page, and
//returns http.Error() if the user is unauthorized. User may decide on a case-to-case basis which
//roles are able to access each page
func Restricted(h http.Handler, mng sessionManager, cookieID string, roles ...string) http.Handler { //this needs to be updated to make role variadic
	fn := func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(cookieID)
		if err != nil {
			log.Println(err)
			http.Error(w, "Unauthorized", 401)
			return
		}

		if err := mng.VerifySession(cookie.Value); err != nil {
			log.Println(err)
			http.Error(w, "Unauthorized", 401)
			return
		}

		if len(roles) == 0 {
			goto serving
		}

		err = mng.CheckRole(roles, cookie.Value)
		if err != nil {
			log.Println(err)
			http.Error(w, "Unauthorized", 401)
			return
		}

	serving:

		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
