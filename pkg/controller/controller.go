package controller

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// AddToManagerFuncs is a list of functions to add all Controllers to the Manager
var AddToManagerFuncs []func(manager.Manager, chan schema.GroupVersionKind) error

// AddToManager adds all Controllers to the Manager
func AddToManager(m manager.Manager, autodetect chan schema.GroupVersionKind) error {
	for _, f := range AddToManagerFuncs {
		if err := f(m, autodetect); err != nil {
			return err
		}
	}
	return nil
}
