package main

import (
	"fmt"
	"net/http"
	"strconv"
)

// add is a math function that takes two integers and returns their sum
func add(a, b int) int {
	return a + b
}

// handler is a function that handles HTTP requests and writes the result of the add function
func handler(w http.ResponseWriter, r *http.Request) {
	// get the query parameters from the request URL
	params := r.URL.Query()
	// get the values of "a" and "b" parameters as strings
	a := params.Get("a")
	b := params.Get("b")
	// convert the strings to integers
	aInt, err1 := strconv.Atoi(a)
	bInt, err2 := strconv.Atoi(b)
	// check for errors in conversion
	if err1 != nil || err2 != nil {
		// write an error message to the response
		fmt.Fprintf(w, "Invalid parameters: %s and %s\n", a, b)
		return
	}
	// call the add function with the integers
	result := add(aInt, bInt)
	// write the result to the response
	fmt.Fprintf(w, "%d + %d = %d\n", aInt, bInt, result)
}

// main is the entry point of the program
func main() {
	// register the handler function for the root path
	http.HandleFunc("/", handler)
	// start the webserver on port 8080
	http.ListenAndServe(":8080", nil)
}
