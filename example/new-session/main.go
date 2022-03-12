package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/Jonny-Burkholder/biscuit/biscuit"
)

/*this example shows the basics of setting up a new session using the session manager, and taking
requests from http*/

//first we will retrieve a new session manager from biscuit. This is important, as all of our
//session management will be run as methods from the session manager. You may create as many
//session managers as your server can handle, but I recommend starting with one to see how it
//feels. The new session manager will run automatically upon instantiation
var manager = biscuit.NewSessionManager()

//next, we'll retrieve our templates and write a function to render them
var templates = template.Must(template.ParseGlob("./*.html"))

type page struct {
	Title string
}

func renderTemplate(tmpl string, w http.ResponseWriter) {
	buf := new(bytes.Buffer) //get a buffer to write to so we avoid those nasty superfluous writer calls
	p := page{Title: "Login"}
	err := templates.ExecuteTemplate(buf, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	buf.WriteTo(w)
	return
}

//now we will create a basic server to handle user login
func main() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/login", handleLogin) //we'll eschew pages unnecessary to the example, for simplicity
	http.HandleFunc("/vallidate", handleVallidate)
	http.HandleFunc("/home/", handleHome)

	fmt.Println("Now serving on port 8080")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusFound)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	renderTemplate("login", w)
}

func handleVallidate(w http.ResponseWriter, r *http.Request) {
	//normally here we would check the password
	r.ParseForm()
	userID, err := manager.NewSession(r.FormValue("username"), r) //create new session in manager by passing a username string and the http request
	if err != nil {
		//handle error
		log.Println(err)
		return //for simplicity, since this is just an example
	}
	manager.SetSessionCookie(w, userID)
	err = manager.Login(userID)
	if err != nil {
		log.Println(err)
	}
	http.Redirect(w, r, "home/"+r.FormValue("username"), http.StatusFound)
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[len("/home/"):]
	cookies := r.Cookies()
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "Congratulations %v! You have successfully logged in.\n\n", name)
	for _, cookie := range cookies {
		fmt.Fprintf(buf, "Your cookie ID is %v\n", cookie)
	}
	buf.WriteTo(w)
}
