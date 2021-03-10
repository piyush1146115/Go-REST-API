package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"github.com/gorilla/mux"
)

var counter int
var mutex = &sync.Mutex{}

type Article struct{
	Title string `json:"Title"`
	Desc string `json:"desc"`
	Content string `json:"content"`
}

type Articles []Article

func allArticles(w http.ResponseWriter, r *http.Request){
	articles := Articles{
		Article{Title: "Test title", Desc: "Test Description", Content: "Hello World"},
		Article{Title: "Hello", Desc: "Article Description", Content: "Article Content"},
		Article{Title: "Hello 2", Desc: "Article Description", Content: "Article Content"},
	}

	fmt.Println("Endpoint Hit: All articles endpoint")
	json.NewEncoder(w).Encode(articles)
}

//func returnAllArticles(w http.ResponseWriter, r *http.Request){
//	fmt.Println("Endpoint Hit: returnAllArticles")
//	json.NewEncoder(w).Encode(articles)
//}

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
	myRouter.HandleFunc("/articles", allArticles)
	log.Fatal(http.ListenAndServe(":10000", myRouter))
}
func main() {
//fmt.Println("Hello")
	fmt.Println("Rest API v2.0 - Mux Routers")
	handleRequests()
}