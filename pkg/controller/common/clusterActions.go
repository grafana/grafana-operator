package common

import (
	"context"
	stdErr "errors"
	"fmt"
	"github.com/go-logr/logr"
	v13 "github.com/openshift/api/route/v1"
	v12 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type ActionRunner interface {
	RunAll(desiredState DesiredClusterState) error
	create(obj runtime.Object) error
	update(obj runtime.Object) error
	delete(obj runtime.Object) error
	updateCr(obj runtime.Object) error
	routeReady(obj runtime.Object) error
	deploymentReady(obj runtime.Object) error
}

type ClusterAction interface {
	Run(runner ActionRunner) (string, error)
}

// The desired cluster state is defined by a list of actions that have to be run to
// get from the current state to the desired state
type DesiredClusterState []ClusterAction

func (d *DesiredClusterState) AddAction(action ClusterAction) DesiredClusterState {
	if action != nil {
		*d = append(*d, action)
	}
	return *d
}

func (d *DesiredClusterState) AddActions(actions []ClusterAction) DesiredClusterState {
	for _, action := range actions {
		d.AddAction(action)
	}
	return *d
}

type ClusterActionRunner struct {
	scheme *runtime.Scheme
	client client.Client
	ctx    context.Context
	log    logr.Logger
	cr     runtime.Object
}

func NewClusterActionRunner(ctx context.Context, client client.Client, scheme *runtime.Scheme, cr runtime.Object) ActionRunner {
	return &ClusterActionRunner{
		scheme: scheme,
		client: client,
		log:    logf.Log.WithName("action-runner"),
		ctx:    ctx,
		cr:     cr,
	}
}

func (i *ClusterActionRunner) RunAll(desiredState DesiredClusterState) error {
	for index, action := range desiredState {
		msg, err := action.Run(i)
		if err != nil {
			i.log.Info(fmt.Sprintf("(%5d) %10s %s", index, "FAILED", msg))
			return err
		}
		i.log.Info(fmt.Sprintf("(%5d) %10s %s", index, "SUCCESS", msg))
	}

	return nil
}

func (i *ClusterActionRunner) create(obj runtime.Object) error {
	err := controllerutil.SetControllerReference(i.cr.(v1.Object), obj.(v1.Object), i.scheme)
	if err != nil {
		return err
	}

	return i.client.Create(i.ctx, obj)
}

func (i *ClusterActionRunner) update(obj runtime.Object) error {
	err := controllerutil.SetControllerReference(i.cr.(v1.Object), obj.(v1.Object), i.scheme)
	if err != nil {
		return err
	}

	err = i.client.Update(i.ctx, obj)
	if err != nil {
		// Update conflicts can happen frequently when kubernetes updates the resource
		// in the background
		if errors.IsConflict(err) {
			return nil
		}
		return err
	}
	return nil
}

func (i *ClusterActionRunner) updateCr(obj runtime.Object) error {
	err := i.client.Update(i.ctx, obj)
	if err != nil {
		// Update conflicts can happen frequently when kubernetes updates the resource
		// in the background
		if errors.IsConflict(err) {
			return nil
		}
		return err
	}
	return nil
}

func (i *ClusterActionRunner) delete(obj runtime.Object) error {
	return i.client.Delete(i.ctx, obj)
}

func (i *ClusterActionRunner) routeReady(obj runtime.Object) error {
	ready := IsRouteReady(obj.(*v13.Route))
	if !ready {
		return stdErr.New("route not ready")
	}
	return nil
}

func (i *ClusterActionRunner) deploymentReady(obj runtime.Object) error {
	ready, err := IsDeploymentReady(obj.(*v12.Deployment))
	if err != nil {
		return err
	}

	if !ready {
		return stdErr.New("deployment not ready")
	}
	return nil
}

// An action to create generic kubernetes resources
// (resources that don't require special treatment)
type GenericCreateAction struct {
	Ref runtime.Object
	Msg string
}

// An action to update generic kubernetes resources
// (resources that don't require special treatment)
type GenericUpdateAction struct {
	Ref runtime.Object
	Msg string
}

type WaitForRouteAction struct {
	Ref runtime.Object
	Msg string
}

type LogAction struct {
	Msg string
}

type RouteReadyAction struct {
	Ref runtime.Object
	Msg string
}

type DeploymentReadyAction struct {
	Ref runtime.Object
	Msg string
}

type UpdateCrAction struct {
	Ref runtime.Object
	Msg string
}

// An action to delete generic kubernetes resources
// (resources that don't require special treatment)
type GenericDeleteAction struct {
	Ref runtime.Object
	Msg string
}

func (i GenericCreateAction) Run(runner ActionRunner) (string, error) {
	return i.Msg, runner.create(i.Ref)
}

func (i GenericUpdateAction) Run(runner ActionRunner) (string, error) {
	return i.Msg, runner.update(i.Ref)
}

func (i GenericDeleteAction) Run(runner ActionRunner) (string, error) {
	return i.Msg, runner.delete(i.Ref)
}

func (i LogAction) Run(runner ActionRunner) (string, error) {
	return i.Msg, nil
}

func (i RouteReadyAction) Run(runner ActionRunner) (string, error) {
	return i.Msg, runner.routeReady(i.Ref)
}

func (i DeploymentReadyAction) Run(runner ActionRunner) (string, error) {
	return i.Msg, runner.deploymentReady(i.Ref)
}

func (i UpdateCrAction) Run(runner ActionRunner) (string, error) {
	return i.Msg, runner.updateCr(i.Ref)
}
