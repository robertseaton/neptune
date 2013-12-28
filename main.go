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
	UserData   template.HTML
}

type User struct {
	Email     string
	Password  string
	SessionID string
}

// Loads a page for use
func loadPage(title string, r *http.Request) (*Page, error) {
	var filename string
	var usr []byte
	if isLoggedIn(r) {
		cookie, _ := r.Cookie("SessionID")
		z := strings.Split(cookie.Value, ":")
		filename = "accounts/" + z[0] + ".txt"
		usr, _ = ioutil.ReadFile(filename)
	} 
	filename = "web/" + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return &Page{Title: title, Body: template.HTML(body), UserData: template.HTML(usr)}, nil
}

// Shows a particular page
func viewHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/"):]
	p, err := loadPage(title, r)

	if err != nil && !isLoggedIn(r) {
		http.Redirect(w, r, "/home", http.StatusFound)
		return
	} else if err != nil {
		http.Redirect(w, r, "/login-succeeded", http.StatusFound)
		return
	}

	t, _ := template.ParseFiles("view.html")
	t.Execute(w, p)
}

// Creates an account and adds it to the Database
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

// Checks to assure credientials
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

// Handles the users loggin and gives them a cookie for doing so
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

	return false
}

// Registers the new user
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

// Logs out the user, removes their cookie from the database
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	
	cookie, err := r.Cookie("SessionID")
	if err != nil {
		fmt.Println(err)
		return
	}

	sessionID := cookie.Value

	z := strings.Split(sessionID, ":")
	usr := new(User)
	usr.Email = z[0]
	sessionID = z[1]

	session, err := mgo.Dial("127.0.0.1:27017/")
	if err != nil {
		return
	}
	c := session.DB("test").C("users")
	c.Remove(bson.M{"email": usr.Email})
	
	http.Redirect(w, r, "/home", http.StatusFound)

}

func main() {
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/", viewHandler)
	http.ListenAndServe(":8080", nil)
}
