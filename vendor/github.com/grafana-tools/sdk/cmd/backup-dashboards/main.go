// This is a simple example of usage of Grafana client
// for copying dashboards and saving them to a disk.
// It really useful for Grafana backups!
//
// Usage:
//   backup-dashboards http://grafana.host:3000 api-key-string-here
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
	"fmt"
	"io/ioutil"
	"os"

	"github.com/grafana-tools/sdk"
)

func main() {
	var (
		boardLinks []sdk.FoundBoard
		rawBoard   []byte
		meta       sdk.BoardProperties
		err        error
	)
	if len(os.Args) != 3 {
		fmt.Fprint(os.Stderr, "Usage:  backup-dashboards http://grafana.host:3000 api-key-string-here\n")
		os.Exit(0)
	}
	c := sdk.NewClient(os.Args[1], os.Args[2], sdk.DefaultHTTPClient)
	if boardLinks, err = c.SearchDashboards("", false); err != nil {
		fmt.Fprintf(os.Stderr, fmt.Sprintf("%s\n", err))
		os.Exit(1)
	}
	for _, link := range boardLinks {
		if rawBoard, meta, err = c.GetRawDashboard(link.URI); err != nil {
			fmt.Fprintf(os.Stderr, fmt.Sprintf("%s for %s\n", err, link.URI))
			continue
		}
		if err = ioutil.WriteFile(fmt.Sprintf("%s.json", meta.Slug), rawBoard, os.FileMode(int(0666))); err != nil {
			fmt.Fprintf(os.Stderr, fmt.Sprintf("%s for %s\n", err, meta.Slug))
		}
	}
}
