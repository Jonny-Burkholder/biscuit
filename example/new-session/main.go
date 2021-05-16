package main

import (
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

//next, we'll retrieve our templates
var templates = template.Must(template.ParseGlob("./*.html"))

//now we will create a basic server to handle user login
func main() {
	http.HandleFunc("/login", handleLogin) //we'll eschew pages unnecessary to the example, for simplicity
	http.HandleFunc("/vallidate", handleVallidate)
	http.HandleFunc("/home/", handleHome)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleLogin(w http.ResponseWriter, r *http.Request) {

}

func handleVallidate(w http.ResponseWriter, r *http.Request) {

}

func handleHome(w http.ResponseWriter, r *http.Request)
