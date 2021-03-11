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
	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/articles", returnAllArticles)
	//myRouter.HandleFunc("/article/{id}", returnSingleArticle)
	myRouter.HandleFunc("/article", createNewArticle).Methods("POST")
	//myRouter.HandleFunc("/article/{id}", deleteArticle).Methods("DELETE")
	myRouter.HandleFunc("/article/{id}", updateArticles).Methods("PUT")
	log.Fatal(http.ListenAndServe(":10000", myRouter))
}

var mySigningKey = []byte("secretphrase")

func GenerateJWT()(string, error){
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["user"] = "Elliot Forbes"
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

	tokenString, err := token.SignedString(mySigningKey)

	if err != nil{
		fmt.Errorf("Something went wrong: %s", err.Error())
		return "", err
	}

	return tokenString, err
}

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