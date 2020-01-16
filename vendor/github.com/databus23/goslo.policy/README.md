[![Build Status](https://travis-ci.org/databus23/goslo.policy.png?branch=master)](https://travis-ci.org/databus23/goslo.policy)

A go implementation of OpenStack's oslo.policy
==============================================

This repository provides a reimplementation of the original [oslo.policy](https://github.com/openstack/oslo.policy) library written in python. It is meant to provide the same RBAC semantics for OpenStack enabled applications written in go.

You can view the API docs here:
http://godoc.org/github.com/databus23/goslo.policy

Usage
-----
```
package main

import (
	"log"

	policy "github.com/databus23/goslo.policy"
)

func main() {
	rules := map[string]string{
		"admin_required": "role:admin",
		"cloud_admin":    "rule:admin_required and domain_id:default",
		"owner":          "user_id:%(user_id)s",
	}
	//Load and parse policy
	enforcer, err := policy.NewEnforcer(rules)
	if err != nil {
		log.Fatal("Failed to parse policy ", err)
	}
	//Context provides the current token & request information needed for enforcement
	ctx := policy.Context{
		Auth: map[string]string{
			"user_id":   "u-1",
			"domain_id": "default",
		},
		Roles: []string{"admin"},
		Request: map[string]string{
			"user_id": "u-1",
		},
	}

	if enforcer.Enforce("cloud_admin", ctx) {
		log.Println("user is a cloud admin")
	}
	if enforcer.Enforce("owner", ctx) {
		log.Println("user is owner")
	}
}
```

The package includes optional debug logging that can be enabled per context:

```
if os.Getenv("DEBUG") == "1" {
    ctx.Logger = log.Printf //or any other function with the same signature
}
```
