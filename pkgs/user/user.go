package user

import (
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	"os"
	"strings"

	"neptune/pkgs/cookies"
)

type User struct {
	Email     string
	Password  string
	SessionID string
}

// Creates an account and adds it to the Database
func CreateAccount(usr *User) bool {

	session, err := mgo.Dial("127.0.0.1:27017/")
	if err != nil {
		fmt.Println(err)
		return false
	}

	c := session.DB("test").C("users")
	err = c.Insert(*usr)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

// Checks if an account exists in the userbase.
func DoesAccountExist(email string) bool {

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
func CheckCredentials(email string, password string) bool {

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

// Updates the user's data in the database (e.g. cookie data).
func UpdateUser(usr *User) bool {
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

// Loads users info - or supplys links to login or register
func LoadUserInfo(title string, r *http.Request) (filename string, option []byte, usr []byte) {

	if cookies.IsLoggedIn(r) {
		cookie, _ := r.Cookie("SessionID")
		z := strings.Split(cookie.Value, ":")
		filename = "accounts/" + z[0]
		usr = []byte("<a href='" + filename + "'>" + z[0] + "</a>: ")
		option = []byte("<a href='/logout'>logout</a>")
	} else {
		option = []byte("<a href='/login'>login</a> or <a href='/register'>register</a>")
	}

	filename = "web/" + title + ".txt"

	// Adds link to profile if clicked.
	file := strings.Split(title, "/")
	if len(file) > 1 {
		filename = "accounts/" + file[1] + ".profile"
	}

	return filename, option, usr

}

func CreateUserFile(usrName string) {

	file, err := os.Create("accounts/" + usrName + ".profile") // creates a file with that usrName
	if err != nil {
		fmt.Println("Error creating user profile")
	}

	s := usrName + " welcome!<br>"
	// TODO add a plugin to a database of books
	file.WriteString(s)

}

func ReadUserFile(usrName string) (file *os.File) {

	file, err := os.Open(usrName)
	if err != nil {
		fmt.Printf("error readUserFile FIX")
	}

	return file

}

func AppendUserFile(usrName string) {

	file, err := os.OpenFile(usrName, os.O_WRONLY, 0666)
	if err != nil {
		fmt.Printf("error appendUserFile FIX")
	}

	s := "nothing"

	file.WriteString(s)

}
