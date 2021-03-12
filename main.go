package main

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var counter int
var mutex = &sync.Mutex{}

type Article struct{
	Id string `json: "Id"`
	Title string `json:"Title"`
	Desc string `json:"desc"`
	Content string `json:"content"`
}

var Articles []Article

//func allArticles(w http.ResponseWriter, r *http.Request){
//	fmt.Println("Endpoint Hit: All articles endpoint")
//	json.NewEncoder(w).Encode(Articles)
//}



func returnSingleArticle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["id"]

	// Loop over all of our Articles
	// if the article.Id equals the key we pass in
	// return the article encoded as JSON
	for _, article := range Articles {
		if article.Id == key {
			json.NewEncoder(w).Encode(article)
		}
	}
}


var users = map[string]string{
	"test": "secret",
	"user1": "password1",
	"user2": "password2",
}

type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}
// Create the JWT key used to create the signature
var jwtKey = []byte("my_secret_key")

func Signin(w http.ResponseWriter, r *http.Request){
	var creds Credentials
	// Get the JSON body and decode into credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get the expected password from our in memory map
	expectedPassword, ok := users[creds.Username]

	// If a password exists for the given user
	// AND, if it is the same as the password we received, the we can move ahead
	// if NOT, then we return an "Unauthorized" status
	if !ok || expectedPassword != creds.Password {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Declare the expiration time of the token
	// here, we have kept it as 5 minutes
	expirationTime := time.Now().Add(5 * time.Minute)
	// Create the JWT claims, which includes the username and expiry time

	claims := &Claims{
		Username: creds.Username,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Create the JWT string
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		// If there is an error in creating the JWT return an internal server error
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Finally, we set the client cookie for "token" as the JWT we just generated
	// We also set an expiry time which is the same as the token itself
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})

}

func Welcome(w http.ResponseWriter, r *http.Request){
	// We can obtain the session token from the requests cookies, which come with every request

	c,err := r.Cookie("token")

	if err != nil {
		if err == http.ErrNoCookie {
			//If the cookie is not set, return an unauthorized status
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// For any other type of error, return a bad request status
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get the JWT string from the cookie
	tknStr := c.Value

	// Initialize a new instance of 'Claims'
	claims := &Claims{}

	// Parse the JWT string and store the result in `claims`.
	// Note that we are passing the key in this method as well. This method will return an error
	// if the token is invalid (if it has expired according to the expiry time we set on sign in),
	// or if the signature does not match
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// Finally, return the welcome message to the user, along with their
	// username given in the token
	w.Write([]byte(fmt.Sprintf("Welcome %s!", claims.Username)))
}

func Refresh(w http.ResponseWriter, r *http.Request){
	c, err := r.Cookie("token")

	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	tknstr := c.Value
	claims := &Claims{}

	tkn, err := jwt.ParseWithClaims(tknstr, claims, func(token *jwt.Token) (interface{}, error){
		return jwtKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid{
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !tkn.Valid{
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// (END) The code up-till this point is the same as the first part of the `Welcome` route

	// We ensure that a new token is not issued until enough time has elapsed
	// In this case, a new token will only be issued if the old token is within
	// 30 seconds of expiry. Otherwise, return a bad request status

	if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) > 30*time.Second {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Now, create a new token for the current use, with a renewed expiration time
	expirationTime := time.Now().Add(5 * time.Minute)
	claims.ExpiresAt = expirationTime.Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//Set the new token as the users 'token' cookie
	http.SetCookie(w, &http.Cookie{
		Name: "token",
		Value: tokenString,
		Expires: expirationTime,
	})
}

func isAuthorised(username, password string) bool{
	pass, ok := users[username]
	if !ok {
		return false
	}
	return password == pass
}

func returnAllArticles(w http.ResponseWriter, r *http.Request){
	w.Header().Add("Content-Type", "application/json")
	username, password, ok := r.BasicAuth()

	if !ok{
		w.Header().Add("WWW-Authenticate", `Basic realm="Give username and password"`)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message": "No basic auth present"}`))
		return
	}

	if !isAuthorised(username, password) {
		w.Header().Add("WWW-Authenticate", `Basic realm="Give username and password"`)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message": "Invalid username or password"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Println("Endpoint Hit: returnAllArticles")
	json.NewEncoder(w).Encode(Articles)
}

func homePage(w http.ResponseWriter, r *http.Request){
	fmt.Fprintf(w, "Welcome to the Homepage!")
	fmt.Println(w, "Homepage Endpoint Hit")
	//http.ServeFile(w, r, r.URL.Path[1:])
}

func incrementCounter(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	counter++
	fmt.Fprintf(w, strconv.Itoa(counter))
	mutex.Unlock()
}

func createNewArticle(w http.ResponseWriter, r *http.Request) {
	// get the body of our POST request
	// return the string response containing the request body
	reqBody, _ := ioutil.ReadAll(r.Body)
	var article Article
	json.Unmarshal(reqBody, &article)
	// update our global Articles array to include
	// our new Article
	Articles = append(Articles, article)

	json.NewEncoder(w).Encode(article)
	//fmt.Fprintf(w, "%+v", string(reqBody))
}

func deleteArticle(w http.ResponseWriter, r *http.Request){
	// once again, we will need to parse the path parameters
	vars := mux.Vars(r)
	// we will need to extract the `id` of the article we
	// wish to delete
	id := vars["id"]

	// we then need to loop through all our articles
	for index, article := range Articles {
		// if our id path parameter matches one of our
		// articles
		if article.Id == id {
			// updates our Articles array to remove the
			// article
			Articles = append(Articles[:index], Articles[index+1:]...)
		}
	}

	fmt.Println(w, "DELETE Endpoint Hit")
}

func updateArticles(w http.ResponseWriter, r *http.Request){

	vars := mux.Vars(r)
	id := vars["id"]

	reqbody,_ := ioutil.ReadAll(r.Body)
	var article Article
	json.Unmarshal(reqbody, &article)

	for index, art := range Articles {
		// if our id path parameter matches one of our
		// articles
		if art.Id == id {
			// updates our Articles array to remove the
			// article
			Articles[index] = article
			//Articles = append(Articles[:index], article)
		}
	}

	json.NewEncoder(w).Encode(article)
}

func handleRequests(){
	//http.HandleFunc("/", homePage)
	//http.HandleFunc("/articles", allArticles)
	//http.HandleFunc("/increment", incrementCounter)
	//log.Fatal(http.ListenAndServe(":10000", nil))

	//http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	//	http.ServeFile(w, r, r.URL.Path[1:])
	//})
	//http.Handle("/", http.FileServer(http.Dir("./Static")))
	//log.Fatal(http.ListenAndServe(":8081", nil))

	myRouter := mux.NewRouter().StrictSlash(true)
	// replace http.HandleFunc with myRouter.HandleFunc
	myRouter.HandleFunc("/signin", Signin)
	myRouter.HandleFunc("/welcome", Welcome)
	myRouter.HandleFunc("/refresh", Refresh)

	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/articles", returnAllArticles)
	//myRouter.HandleFunc("/article/{id}", returnSingleArticle)
	myRouter.HandleFunc("/article", createNewArticle).Methods("POST")
	//myRouter.HandleFunc("/article/{id}", deleteArticle).Methods("DELETE")
	myRouter.HandleFunc("/article/{id}", updateArticles).Methods("PUT")
	log.Fatal(http.ListenAndServe(":10000", myRouter))
}

//var mySigningKey = []byte("secretphrase")
//
//func GenerateJWT()(string, error){
//	token := jwt.New(jwt.SigningMethodHS256)
//	claims := token.Claims.(jwt.MapClaims)
//
//	claims["authorized"] = true
//	claims["user"] = "Elliot Forbes"
//	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()
//
//	tokenString, err := token.SignedString(mySigningKey)
//
//	if err != nil{
//		fmt.Errorf("Something went wrong: %s", err.Error())
//		return "", err
//	}
//
//	return tokenString, err
//}

func main() {
//fmt.Println("Hello")
	Articles = []Article{
		Article{Id: "1", Title: "Test title", Desc: "Test Description", Content: "Hello World"},
		Article{Id: "2", Title: "Hello", Desc: "Article Description", Content: "Article Content"},
		Article{Id: "3", Title: "Hello 2", Desc: "Article Description", Content: "Article Content"},
	}

	fmt.Println("Rest API v2.0 - Mux Routers")
	handleRequests()

	//tokenString, err := GenerateJWT()
	//if err != nil{
	//	fmt.Println("Error generating token string")
	//}
	//
	//fmt.Println(tokenString)
}