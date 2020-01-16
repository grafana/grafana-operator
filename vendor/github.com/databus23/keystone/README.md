[![Build Status](https://travis-ci.org/databus23/keystone.png?branch=master)](https://travis-ci.org/databus23/keystone)

Go Keystone Middleware
======================
A go http middleware for authenticating incoming http request against Openstack Keystone. It it modelled after the original [python middleware for keystone](http://docs.openstack.org/developer/keystonemiddleware/middlewarearchitecture.html).

The middleware authenticates incoming requests by validating the `X-Auth-Token` header and adding additional headers to the incoming request containing the validation result. The final authentication/authorisation decision is delegated to subsequent http handlers.

You can view the API docs here:
http://godoc.org/github.com/databus23/keystone

Usage
-----
```
// main.go
package main

import (
	"fmt"
	"net/http"

	"github.com/databus23/keystone"
)

var myApp = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Identity-Status") == "Confirmed" {
		fmt.Fprintf(w, "This is an authenticated request")
		fmt.Fprintf(w, "Username: %s", r.Header.Get("X-User-Name"))

	} else {
		w.WriteHeader(401)
		fmt.Fprintf(w, "Invalid or no token provided")
	}
})

func main() {
	auth := keystone.New("http://keystone.endpoint:5000/v3")
	handler := auth.Handler(myApp)
	http.ListenAndServe("0.0.0.0:3000", handler)
}
```

Headers 
-------
The middleware sets the following HTTP header for subsequent handlers.

 * `X-Identity-Status`: Token validation result. Either `Confirmed` or `Invalid`

If the validation was successful the following headers are also set

 * `X-User-Id`
 * `X-User-Name`
 * `X-User-Domain-Id`
 * `X-User-Domain-Name`
 * `X-Project-Name` *project scoped tokens only*
 * `X-Project-Id` *project scoped tokens only*
 * `X-Project-Domain-Name` *project scoped tokens only*
 * `X-Project-Domain-Id` *project scoped tokens only*
 * `X-Domain-Id` *domain scoped tokens only*
 * `X-Domain-Name` *domain scoped tokens only*
 * `X-Roles` A comma separated list of role names associated with the user for the current scope
