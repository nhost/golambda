//
//	The function can import dependencies that are NOT in main package.
//	But the function to be built, must always be in the main package.
//
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

//	The first exported "Handler" function will be attached to the router.
//	Multiple Handler functions are not allowed.
//	Handler funcs with any other name than "Handler" are also not allowed
//	"Handler" func should be exported, it cannot be "handler".

type Body struct {
	FirstName string
	LastName  string
}

//	After build this function with this golambda utility,
//	and then deploying it to lambda,
//	send a GET request to your lambda function endpoint,
//	with following body:
//
//	{
//		"first_name": "Mrinal",
//		"last_name": "Wahal"
//	}
//
//	Expected Output: `Nhost pays it's respects to Wahal, Mrinal!`

func Handler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	var payload Body
	json.Unmarshal(body, &payload)
	fmt.Fprintf(w, "Nhost pays it's respects to %s, %s!", payload.LastName, payload.FirstName)
}
