// This is a simple example of usage of Grafana client
// for importing dashboards from a bunch of JSON files (current dir used).
// You are can export dashboards with backup-dashboards utitity.
// NOTE: old dashboards with same names will be silently overrided!
//
// Usage:
//   import-dashboards http://sdk host:3000 api-key-string-here
//
// You need get API key with Admin rights from your Grafana!
package main

/*
   Copyright 2016 Alexander I.Grafov <grafov@gmail.com>

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.

   ॐ तारे तुत्तारे तुरे स्व
*/

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/grafana-tools/sdk"
)

func main() {
	var (
		filesInDir []os.FileInfo
		rawBoard   []byte
		err        error
	)
	if len(os.Args) != 3 {
		fmt.Fprint(os.Stderr, "Usage: import-dashboards http://grafana-host:3000 api-key-string-here\n")
		os.Exit(0)
	}
	c := sdk.NewClient(os.Args[1], os.Args[2], sdk.DefaultHTTPClient)
	filesInDir, err = ioutil.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range filesInDir {
		if strings.HasSuffix(file.Name(), ".json") {
			if rawBoard, err = ioutil.ReadFile(file.Name()); err != nil {
				log.Println(err)
				continue
			}
			var board sdk.Board
			if err = json.Unmarshal(rawBoard, &board); err != nil {
				log.Println(err)
				continue
			}
			c.DeleteDashboard(board.UpdateSlug())
			_, err := c.SetDashboard(board, false)
			if err != nil {
				log.Printf("error on importing dashboard %s", board.Title)
				continue
			}
		}
	}
}
