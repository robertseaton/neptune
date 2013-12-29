package cookies

import(

	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type User struct {
	Email     string
	Password  string
	SessionID string
}

// Create a login cookie.
func LoginCookie(username string) http.Cookie {
	cookieValue := username + ":" + strconv.Itoa(rand.Intn(100000000))
	expire := time.Now().AddDate(0, 0, 1)
	return http.Cookie{Name: "SessionID", Value: cookieValue, Expires: expire, HttpOnly: true}
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
	//	return "", "Failed to find user in database."
	}

	z := strings.Split(result.SessionID, ":")
	if z != nil {
		return "", ""
	}

	return z[1], ""

}

// Check if the user is logged in.
func IsLoggedIn(r *http.Request) bool {
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
