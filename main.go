package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
	"github.com/gorilla/mux"
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

func returnAllArticles(w http.ResponseWriter, r *http.Request){
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
	myRouter.HandleFunc("/article/{id}", deleteArticle).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":10000", myRouter))
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
}