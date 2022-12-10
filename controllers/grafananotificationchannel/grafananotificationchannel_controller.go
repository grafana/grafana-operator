/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package grafananotificationchannel

import (
	"context"
	"crypto/tls"
	"encoding/json"
	defaultErrors "errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-logr/logr"
	grafanav1alpha1 "github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	"github.com/grafana-operator/grafana-operator/v4/controllers/constants"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/grafana-operator/grafana-operator/v4/controllers/common"
	"github.com/grafana-operator/grafana-operator/v4/controllers/config"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	ControllerName = "controller_grafananotificationchannel"
)

type GrafanaChannel struct {
	ID                    *uint   `json:"id"`
	UID                   *string `json:"uid"`
	Name                  *string `json:"name"`
	Type                  *string `json:"type"`
	IsDefault             *bool   `json:"isDefault"`
	SendReminder          *bool   `json:"sendReminder"`
	DisableResolveMessage *bool   `json:"disableResolveMessage"`
}

// NewReconciler returns a new reconcile.Reconciler
func NewReconciler(mgr manager.Manager) reconcile.Reconciler {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	return &GrafanaNotificationChannelReconciler{
		client: mgr.GetClient(),
		transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec
			},
		},
		config:   config.GetNotificationControllerConfig(),
		context:  ctx,
		cancel:   cancel,
		recorder: mgr.GetEventRecorderFor(ControllerName),
		state:    common.ControllerState{},
		Log:      mgr.GetLogger(),
	}
}

// Add creates a new GrafanaNotificationChannel Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, namespace string) error {
	return SetupWithManager(mgr, NewReconciler(mgr), namespace)
}

// +kubebuilder:rbac:groups=integreatly.org,resources=grafananotificationchannels,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=integreatly.org,resources=grafananotificationchannels/status,verbs=get;update;patch

// SetupWithManager sets up the controller with the Manager.
func SetupWithManager(mgr ctrl.Manager, r reconcile.Reconciler, namespace string) error {
	c, err := controller.New("grafananotificationchannel-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource GrafanaNotificationChannel
	err = c.Watch(&source.Kind{Type: &grafanav1alpha1.GrafanaNotificationChannel{}}, &handler.EnqueueRequestForObject{})
	if err == nil {
		log.Log.Info("Starting notificationchannel controller")
	}

	ref := r.(*GrafanaNotificationChannelReconciler) //nolint
	ticker := time.NewTicker(config.GetControllerConfig().RequeueDelay)
	sendEmptyRequest := func() {
		request := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: namespace,
				Name:      "",
			},
		}
		_, err := r.Reconcile(ref.context, request)
		if err != nil {
			return
		}
	}

	go func() {
		for range ticker.C {
			log.Log.Info("running periodic notificationchannel resync")
			sendEmptyRequest()
		}
	}()

	go func() {
		for stateChange := range common.NotificationChannelControllerEvents {
			// Controller state updated
			ref.state = stateChange
		}
	}()

	return err
}

var _ reconcile.Reconciler = &GrafanaNotificationChannelReconciler{}

// GrafanaNotificationChannelReconciler reconciles a GrafanaNotificationChannel object
type GrafanaNotificationChannelReconciler struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client    client.Client
	transport *http.Transport
	config    *config.NotificationControllerConfig
	context   context.Context
	cancel    context.CancelFunc
	recorder  record.EventRecorder
	state     common.ControllerState
	Log       logr.Logger
}

// Reconcile , The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *GrafanaNotificationChannelReconciler) Reconcile(context context.Context, request reconcile.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues(ControllerName, request.NamespacedName)

	// If Grafana is not running there is no need to continue
	if !r.state.GrafanaReady {
		logger.Info("no grafana instance available")
		return reconcile.Result{Requeue: false}, nil
	}

	grafanaClient, err := r.getClient()
	if err != nil {
		// we handle the error by requeing, safe to ignore nilerr return
		return reconcile.Result{RequeueAfter: config.GetControllerConfig().RequeueDelay}, nil //nolint:nilerr
	}

	// Initial request?
	if request.Name == "" {
		return r.reconcileNotificationChannels(request, grafanaClient)
	}

	// Check if the label selectors are available yet. If not then the grafana controller
	// has not finished initializing and we can't continue. Reschedule for later.
	if r.state.DashboardSelectors == nil {
		return reconcile.Result{RequeueAfter: config.GetControllerConfig().RequeueDelay}, nil
	}

	// Fetch the GrafanaNotificationChannel instance
	instance := &grafanav1alpha1.GrafanaNotificationChannel{}
	err = r.client.Get(r.context, request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// If some notificationchannel has been deleted, then always re sync the world
			logger.Info(fmt.Sprintf("deleting notificationchannel %v/%v", request.Namespace, request.Name))
			return r.reconcileNotificationChannels(request, grafanaClient)
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// If the notificationchannel does not match the label selectors then we ignore it
	cr := instance.DeepCopy()
	if !r.isMatch(cr) {
		logger.Info(fmt.Sprintf("notificationchannel %v/%v found but selectors do not match",
			cr.Namespace, cr.Name))
		return reconcile.Result{}, nil
	}

	// Otherwise always re sync all notificationchannels in the namespace
	return r.reconcileNotificationChannels(request, grafanaClient)
}

// Check if a given notificationchannel (by name) is present in the list of
// notificationchannels in the namespace
func inNamespace(namespaceNotificationChannels *grafanav1alpha1.GrafanaNotificationChannelList, item *grafanav1alpha1.GrafanaNotificationChannelRef) bool {
	for _, d := range namespaceNotificationChannels.Items {
		if item.Name == d.Name && item.Namespace == d.Namespace {
			return true
		}
	}
	return false
}

// Returns the hash of a notificationchannel if it is known
func findHash(knownNotificationChannels []*grafanav1alpha1.GrafanaNotificationChannelRef, item *grafanav1alpha1.GrafanaNotificationChannel) string {
	for _, d := range knownNotificationChannels {
		if item.Name == d.Name && item.Namespace == d.Namespace {
			return d.Hash
		}
	}
	return ""
}

//nolint:funlen
func (r *GrafanaNotificationChannelReconciler) reconcileNotificationChannels(request reconcile.Request, grafanaClient GrafanaClient) (reconcile.Result, error) { //nolint:cyclop
	// Collect known and namespace notificationchannels
	knownNotificationChannels := r.config.GetNotificationChannels(request.Namespace)
	namespaceNotificationChannels := &grafanav1alpha1.GrafanaNotificationChannelList{}
	err := r.client.List(r.context, namespaceNotificationChannels)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Prepare lists
	var notificationchannelsToDelete []*grafanav1alpha1.GrafanaNotificationChannelRef

	// NotificationChannels to delete: notificationchannels that are known but not found
	// any longer in the namespace
	for _, notificationchannel := range knownNotificationChannels {
		if !inNamespace(namespaceNotificationChannels, notificationchannel) {
			notificationchannelsToDelete = append(notificationchannelsToDelete, notificationchannel)
		}
	}

	// Process new/updated notificationchannels
	for index, notificationchannel := range namespaceNotificationChannels.Items {
		// Is this a notificationchannel we care about (matches the label selectors)?
		if !r.isMatch(&namespaceNotificationChannels.Items[index]) {
			r.Log.Info(fmt.Sprintf("notificationchannel %v/%v found but selectors do not match",
				notificationchannel.Namespace, notificationchannel.Name))
			continue
		}

		// Process the notificationchannel. Use the known hash of an existing notificationchannel
		// to determine if an update is required
		knownHash := findHash(knownNotificationChannels, &namespaceNotificationChannels.Items[index])
		pipeline := NewNotificationChannelPipeline(r.client, &namespaceNotificationChannels.Items[index])

		// We need to always get the spec in order to extract the uid
		// TODO: in v5, better to rely on uid of CR and ignore the one passed in json
		processed, err := pipeline.ProcessNotificationChannel(knownHash, true)
		if err != nil {
			r.Log.Error(err, fmt.Sprintf("cannot process notificationchannel %v/%v", notificationchannel.Namespace, notificationchannel.Name))
			r.manageError(&namespaceNotificationChannels.Items[index], err)
			continue
		}

		// Parse spec and make sure uid is defined
		rawJson := GrafanaChannel{}
		if err := json.Unmarshal(processed, &rawJson); err != nil {
			return reconcile.Result{}, err
		}
		if rawJson.UID == nil {
			r.Log.Info(fmt.Sprintf("cannot process notificationchannel %v/%v, UID is nil", notificationchannel.Namespace, notificationchannel.Name))
			return reconcile.Result{}, nil
		}

		// TODO: in v5 with an updated client, we should have a detailed check for the status code
		exists := false
		if _, err = grafanaClient.GetNotificationChannel(*rawJson.UID); err == nil {
			exists = true
		}

		// This time, we should get a non-nil spec only if the notification channel has a different hash or it doesn't exist in grafana (e.g. was deleted via grafana console)
		processed, err = pipeline.ProcessNotificationChannel(knownHash, !exists)
		if err != nil {
			r.Log.Error(err, fmt.Sprintf("cannot process notificationchannel %v/%v", notificationchannel.Namespace, notificationchannel.Name))
			r.manageError(&namespaceNotificationChannels.Items[index], err)
			continue
		}

		if processed == nil {
			continue
		}

		// Create or Update notification channel
		var status GrafanaResponse
		if exists {
			status, err = grafanaClient.UpdateNotificationChannel(processed, *rawJson.UID)
		} else {
			status, err = grafanaClient.CreateNotificationChannel(processed)
		}
		if err != nil {
			r.Log.Info(fmt.Sprintf("cannot submit notificationchannel %v/%v", notificationchannel.Namespace, notificationchannel.Name))
			r.manageError(&namespaceNotificationChannels.Items[index], err)
			continue
		}

		err = r.manageSuccess(&namespaceNotificationChannels.Items[index], status, pipeline.NewHash())
		if err != nil {
			r.manageError(&namespaceNotificationChannels.Items[index], err)
		}
	}

	for _, notificationchannel := range notificationchannelsToDelete {
		status, err := grafanaClient.DeleteNotificationChannelByUID(notificationchannel.UID)
		if err != nil {
			r.Log.Error(err, fmt.Sprintf("error deleting notificationchannel %v, status was %v/%v",
				notificationchannel.UID,
				*status.Message,
				*status.UID))
		}
		r.Log.Info(fmt.Sprintf("delete result was %v", *status.Message))
		r.config.RemoveNotificationChannel(notificationchannel.Namespace, notificationchannel.Name)
	}

	// Mark the notificationchannels as synced so that the current state can be written
	// to the Grafana CR by the grafana controller
	r.config.AddConfigItem(config.ConfigGrafanaNotificationChannelsSynced, true)
	return reconcile.Result{Requeue: false}, nil
}

// Handle success case: update notificationchannel metadata (id, uid) and update the list
// of plugins
func (r *GrafanaNotificationChannelReconciler) manageSuccess(notificationchannel *grafanav1alpha1.GrafanaNotificationChannel, status GrafanaResponse, hash string) error {
	msg := fmt.Sprintf("notificationchannel %v/%v successfully submitted",
		notificationchannel.Namespace,
		notificationchannel.Name)

	r.recorder.Event(notificationchannel, "Normal", "Success", msg)
	r.Log.Info(msg)

	notificationchannel.Status.UID = *status.UID
	notificationchannel.Status.ID = *status.ID
	notificationchannel.Status.Phase = grafanav1alpha1.PhaseReconciling
	notificationchannel.Status.Hash = hash
	notificationchannel.Status.Message = constants.GrafanaSuccessMsg

	r.config.AddNotificationChannel(notificationchannel)

	return r.client.Status().Update(r.context, notificationchannel)
}

// Handle error case: update notificationchannel with error message and status
func (r *GrafanaNotificationChannelReconciler) manageError(notificationchannel *grafanav1alpha1.GrafanaNotificationChannel, issue error) {
	r.recorder.Event(notificationchannel, "Warning", "ProcessingError", issue.Error())
	notificationchannel.Status.Phase = grafanav1alpha1.PhaseFailing
	notificationchannel.Status.Message = issue.Error()

	err := r.client.Status().Update(r.context, notificationchannel)
	if err != nil {
		// Ignore conclicts. Resource might just be outdated.
		if errors.IsConflict(err) {
			return
		}
		r.Log.Error(err, "error updating notificationchannel status")
	}
}

// Get an authenticated grafana API client
func (r *GrafanaNotificationChannelReconciler) getClient() (GrafanaClient, error) {
	url := r.state.AdminUrl
	if url == "" {
		return nil, defaultErrors.New("cannot get grafana admin url")
	}

	username := os.Getenv(constants.GrafanaAdminUserEnvVar)
	if username == "" {
		return nil, defaultErrors.New("invalid credentials (username)")
	}

	password := os.Getenv(constants.GrafanaAdminPasswordEnvVar)
	if password == "" {
		return nil, defaultErrors.New("invalid credentials (password)")
	}

	duration := time.Duration(r.state.ClientTimeout)

	return NewGrafanaClient(url, username, password, r.transport, duration), nil
}

// Test if a given notificationchannel matches an array of label selectors
func (r *GrafanaNotificationChannelReconciler) isMatch(item *grafanav1alpha1.GrafanaNotificationChannel) bool {
	if r.state.DashboardSelectors == nil {
		return false
	}

	match, err := item.MatchesSelectors(r.state.DashboardSelectors)
	if err != nil {
		r.Log.Error(err, fmt.Sprintf("error matching selectors against %v/%v",
			item.Namespace,
			item.Name))
		return false
	}
	return match
}
