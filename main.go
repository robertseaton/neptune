package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"neptune/pkgs/user"
	"neptune/pkgs/codify"
)

type Page struct {
	Title string
	Body  template.HTML
	User   template.HTML
}

type User struct {
	Email     string
	Password  string
	SessionID string
}

func loadPage(title string) (*Page, error) {
	filename := "web/" + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: template.HTML(body)}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	/*if isLoggedIn(r) {
		fmt.Println("The user is logged in.")
	} else {
		fmt.Println("The user is not logged in.")
	}
	*/
	title := r.URL.Path[len("/"):]
	// check user status loged-in/not
	p, err := loadPage(title)

	if err != nil {
		http.Redirect(w, r, "/home", http.StatusFound)
		return
	}

	t, _ := template.ParseFiles("view.html")
	t.Execute(w, p)
}

func createAccount(usr *User) bool {
	session, err := mgo.Dial("127.0.0.1:27017/")
	if err != nil {
		fmt.Println(err)
		return false
	}

	c := session.DB("test").C("users")
	err = c.Insert(usr)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

// Checks if an account exists in the userbase.
func doesAccountExist(email string) bool {
	session, err := mgo.Dial("127.0.0.1:27017/")
	if err != nil {
		panic(err)
	}

	result := User{}
	c := session.DB("test").C("users")

	err = c.Find(bson.M{"email": email}).One(&result)
	if err != nil {
		return false
	}

	return true
}

func checkCredentials(email string, password string) bool {
	password = codify.SHA(password)
	session, err := mgo.Dial("127.0.0.1:27017/")
	if err != nil {
		panic(err)
	}

	result := User{}
	c := session.DB("test").C("users")

	err = c.Find(bson.M{"email": email}).One(&result)
	if err != nil {
		return false
	}

	if result.Password == password && result.Email == email {
		return true
	}

	return false
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	usr := new(User)
	usr.Email = r.FormValue("email")
	usr.Password = r.FormValue("pwd")

	if len(usr.Password) > 0 {
		ok := checkCredentials(usr.Email, usr.Password)
		if ok {
			user.CreateUserFile(usr.Email)
			cookie := loginCookie(usr.Email)
			http.SetCookie(w, &cookie)
			usr.SessionID = cookie.Value
			_ = updateUser(usr)
			http.Redirect(w, r, "/login-succeeded", http.StatusFound)
		} else {
			http.Redirect(w, r, "/login-failed", http.StatusFound)
		}
	} else {
		viewHandler(w, r)
	}
}

// Create a login cookie.
func loginCookie(username string) http.Cookie {
	cookieValue := username + ":" + strconv.Itoa(rand.Intn(100000000))
	expire := time.Now().AddDate(0, 0, 1)
	return http.Cookie{Name: "SessionID", Value: cookieValue, Expires: expire, HttpOnly: true}
}

// Updates the user's data in the database (e.g. cookie data).
func updateUser(usr *User) bool {
	session, err := mgo.Dial("127.0.0.1:27017/")
	if err != nil {
		fmt.Println(err)
		return false
	}

	c := session.DB("test").C("users")
	err = c.Update(bson.M{"email": usr.Email}, usr)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

// Check the database for the user's session ID.
func lookupSessionID(email string) (string, string) {
	session, err := mgo.Dial("127.0.0.1:27017/")
	if err != nil {
		return "", "Failed to connect to database."
	}

	result := User{}
	c := session.DB("test").C("users")

	err = c.Find(bson.M{"email": email}).One(&result)
	if err != nil {
		return "", "Failed to find user in database."
	}

	z := strings.Split(result.SessionID, ":")

	return z[1], ""

}

// Check if the user is logged in.
func isLoggedIn(r *http.Request) bool {
	cookie, err := r.Cookie("SessionID")
	if err != nil {
		fmt.Println(err)
		return false
	}

	sessionID := cookie.Value

	fmt.Println("SessionID: ", sessionID)

	z := strings.Split(sessionID, ":")
	email := z[0]
	sessionID = z[1]

	expectedSessionID, errz := lookupSessionID(email)

	if errz != "" {
		fmt.Println(errz)
		return false
	}

	if sessionID == expectedSessionID {
		return true
	}

	//fmt.Println("Got %s, expected %s.", sessionID, expectedSessionID)
	return false
}

func getUserID() {

}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	usr := new(User)
	usr.Email = r.FormValue("email")
	pass := r.FormValue("pwd")
	
	if len(pass) > 0 {
		usr.Password = codify.SHA(pass)
		if doesAccountExist(usr.Email) {
			http.Redirect(w, r, "/account-exists", http.StatusFound)
		} else {
			ok := createAccount(usr)
			if ok {
				http.Redirect(w, r, "/success", http.StatusFound)
			} else {
				http.Redirect(w, r, "/failed", http.StatusFound)
			}
		}
	} else {
		viewHandler(w, r)
	}
}

func main() {
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/", viewHandler)
	http.ListenAndServe(":8080", nil)
}
