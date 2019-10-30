package common

import "fmt"

type GrafanaNotExists struct{}

func (g *GrafanaNotExists) Error() string {
	return fmt.Sprint("grafana does not exist")
}
