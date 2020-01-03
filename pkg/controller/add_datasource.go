package controller

import (
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/grafanadatasource"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, grafanadatasource.Add)
}
