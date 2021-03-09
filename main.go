package main

import (
	"fmt"
	"log"
	"net/http"
)

func homePage(w http.ResponseWriter, r *http.Request){
	fmt.Fprint(w, "Homepage Endpoint Hit")
}

func handleRequests(){
	http.HandleFunc("/", homePage)
	log.Fatal(http.ListenAndServe(":8081", nil))
}
func main() {
//fmt.Println("Hello")
	handleRequests()
}