package grafananotificationchannel

import (
	"context"
	"encoding/json"
	defaultErrors "errors"
	"fmt"
	"time"

	grafanav1alpha1 "github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/common"
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/config"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	ControllerName = "controller_grafananotificationchannel"
)

var log = logf.Log.WithName(ControllerName)

type GrafanaChannel struct {
	ID                    *uint   `json:"id"`
	UID                   *string `json:"uid"`
	Name                  *string `json:"name"`
	Type                  *string `json:"type"`
	IsDefault             *bool   `json:"isDefault"`
	SendReminder          *bool   `json:"sendReminder"`
	DisableResolveMessage *bool   `json:"disableResolveMessage"`
}

// Add creates a new GrafanaNotificationChannel Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, namespace string) error {
	return add(mgr, newReconciler(mgr), namespace)
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	return &ReconcileGrafanaNotificationChannel{
		client:   mgr.GetClient(),
		config:   config.GetNotificationControllerConfig(),
		context:  ctx,
		cancel:   cancel,
		recorder: mgr.GetEventRecorderFor(ControllerName),
		state:    common.ControllerState{},
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler, namespace string) error {
	// Create a new controller
	c, err := controller.New("grafananotificationchannel-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource GrafanaNotificationChannel
	err = c.Watch(&source.Kind{Type: &grafanav1alpha1.GrafanaNotificationChannel{}}, &handler.EnqueueRequestForObject{})
	if err == nil {
		log.Info("Starting notificationchannel controller")
	}

	ref := r.(*ReconcileGrafanaNotificationChannel)
	ticker := time.NewTicker(config.RequeueDelay)
	sendEmptyRequest := func() {
		request := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: namespace,
				Name:      "",
			},
		}
		r.Reconcile(request)
	}

	go func() {
		for range ticker.C {
			log.Info("running periodic notificationchannel resync")
			sendEmptyRequest()
		}
	}()

	go func() {
		for stateChange := range common.ControllerEvents {
			// Controller state updated
			ref.state = stateChange
		}
	}()

	return err
}

var _ reconcile.Reconciler = &ReconcileGrafanaNotificationChannel{}

// ReconcileGrafanaNotificationChannel reconciles a GrafanaNotificationChannel object
type ReconcileGrafanaNotificationChannel struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client   client.Client
	config   *config.NotificationControllerConfig
	context  context.Context
	cancel   context.CancelFunc
	recorder record.EventRecorder
	state    common.ControllerState
}

// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileGrafanaNotificationChannel) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// If Grafana is not running there is no need to continue
	if r.state.GrafanaReady == false {
		log.Info("no grafana instance available")
		return reconcile.Result{Requeue: false}, nil
	}

	client, err := r.getClient()
	if err != nil {
		return reconcile.Result{RequeueAfter: config.RequeueDelay}, nil
	}

	// Initial request?
	if request.Name == "" {
		return r.reconcileNotificationChannels(request, client)
	}

	// Check if the label selectors are available yet. If not then the grafana controller
	// has not finished initializing and we can't continue. Reschedule for later.
	if r.state.DashboardSelectors == nil {
		return reconcile.Result{RequeueAfter: config.RequeueDelay}, nil
	}

	// Fetch the GrafanaNotificationChannel instance
	instance := &grafanav1alpha1.GrafanaNotificationChannel{}
	err = r.client.Get(r.context, request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// If some notificationchannel has been deleted, then always re sync the world
			log.Info(fmt.Sprintf("deleting notificationchannel %v/%v", request.Namespace, request.Name))
			return r.reconcileNotificationChannels(request, client)
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// If the notificationchannel does not match the label selectors then we ignore it
	cr := instance.DeepCopy()
	if !r.isMatch(cr) {
		log.Info(fmt.Sprintf("notificationchannel %v/%v found but selectors do not match",
			cr.Namespace, cr.Name))
		return reconcile.Result{}, nil
	}

	// Otherwise always re sync all notificationchannels in the namespace
	return r.reconcileNotificationChannels(request, client)
}

func (r *ReconcileGrafanaNotificationChannel) reconcileNotificationChannels(request reconcile.Request, grafanaClient GrafanaClient) (reconcile.Result, error) {
	// Collect known and namespace notificationchannels
	knownNotificationChannels := r.config.GetNotificationChannels(request.Namespace)
	namespaceNotificationChannels := &grafanav1alpha1.GrafanaNotificationChannelList{}
	err := r.client.List(r.context, namespaceNotificationChannels)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Prepare lists
	var notificationchannelsToDelete []*grafanav1alpha1.GrafanaNotificationChannelRef

	// Check if a given notificationchannel (by name) is present in the list of
	// notificationchannels in the namespace
	inNamespace := func(item string) bool {
		for _, notificationchannel := range namespaceNotificationChannels.Items {
			if notificationchannel.Name == item {
				return true
			}
		}
		return false
	}

	// Returns the hash of a notificationchannel if it is known
	findHash := func(item *grafanav1alpha1.GrafanaNotificationChannel) string {
		for _, d := range knownNotificationChannels {
			if item.Name == d.Name {
				return d.Hash
			}
		}
		return ""
	}

	// NotificationChannels to delete: notificationchannels that are known but not found
	// any longer in the namespace
	for _, notificationchannel := range knownNotificationChannels {
		if !inNamespace(notificationchannel.Name) {
			notificationchannelsToDelete = append(notificationchannelsToDelete, notificationchannel)
		}
	}

	// Process new/updated notificationchannels
	for _, notificationchannel := range namespaceNotificationChannels.Items {
		// Is this a notificationchannel we care about (matches the label selectors)?
		if !r.isMatch(&notificationchannel) {
			log.Info(fmt.Sprintf("notificationchannel %v/%v found but selectors do not match",
				notificationchannel.Namespace, notificationchannel.Name))
			continue
		}

		// Process the notificationchannel. Use the known hash of an existing notificationchannel
		// to determine if an update is required
		knownHash := findHash(&notificationchannel)
		pipeline := NewNotificationChannelPipeline(r.client, &notificationchannel)
		processed, err := pipeline.ProcessNotificationChannel(knownHash)
		if err != nil {
			log.Error(err, fmt.Sprintf("cannot process notificationchannel %v/%v", notificationchannel.Namespace, notificationchannel.Name))
			r.manageError(&notificationchannel, err)
			continue
		}

		if processed == nil {
			continue
		}

		var status GrafanaResponse
		// if knownHash is empty channel wasn't added before
		// if channel is not in knownNotificationChannels, but already presents in Grafana (in case if operator was restarted) corner case
		// API creating call will return response code 500.
		// Try to check if channel already present then create/updated existing channel
		if knownHash == "" {
			rawJson := GrafanaChannel{}
			if err := json.Unmarshal(processed, &rawJson); err != nil {
				return reconcile.Result{}, err
			}
			if _, err := grafanaClient.GetNotificationChannel(*rawJson.UID); err != nil {
				status, err = grafanaClient.CreateNotificationChannel(processed)
			} else {
				status, err = grafanaClient.UpdateNotificationChannel(processed, notificationchannel.Status.UID)
			}
		} else {
			status, err = grafanaClient.UpdateNotificationChannel(processed, notificationchannel.Status.UID)
		}

		if err != nil {
			log.Info(fmt.Sprintf("cannot submit notificationchannel %v/%v", notificationchannel.Namespace, notificationchannel.Name))
			r.manageError(&notificationchannel, err)
			continue
		}

		err = r.manageSuccess(&notificationchannel, status, pipeline.NewHash())
		if err != nil {
			r.manageError(&notificationchannel, err)
		}
	}

	for _, notificationchannel := range notificationchannelsToDelete {
		status, err := grafanaClient.DeleteNotificationChannelByUID(notificationchannel.UID)
		if err != nil {
			log.Error(err, fmt.Sprintf("error deleting notificationchannel %v, status was %v/%v",
				notificationchannel.UID,
				*status.Message,
				*status.UID))
		}
		log.Info(fmt.Sprintf("delete result was %v", *status.UID))
		r.config.RemoveNotificationChannel(notificationchannel.Namespace, notificationchannel.Name)
	}

	// Mark the notificationchannels as synced so that the current state can be written
	// to the Grafana CR by the grafana controller
	r.config.AddConfigItem(config.ConfigGrafanaNotificationChannelsSynced, true)
	return reconcile.Result{Requeue: false}, nil
}

// Handle success case: update notificationchannel metadata (id, uid) and update the list
// of plugins
func (r *ReconcileGrafanaNotificationChannel) manageSuccess(notificationchannel *grafanav1alpha1.GrafanaNotificationChannel, status GrafanaResponse, hash string) error {
	msg := fmt.Sprintf("notificationchannel %v/%v successfully submitted",
		notificationchannel.Namespace,
		notificationchannel.Name)

	r.recorder.Event(notificationchannel, "Normal", "Success", msg)
	log.Info(msg)

	notificationchannel.Status.UID = *status.UID
	notificationchannel.Status.ID = *status.ID
	notificationchannel.Status.Phase = grafanav1alpha1.PhaseReconciling
	notificationchannel.Status.Hash = hash
	notificationchannel.Status.Message = "success"

	r.config.AddNotificationChannel(notificationchannel)

	return r.client.Status().Update(r.context, notificationchannel)
}

// Handle error case: update notificationchannel with error message and status
func (r *ReconcileGrafanaNotificationChannel) manageError(notificationchannel *grafanav1alpha1.GrafanaNotificationChannel, issue error) {
	r.recorder.Event(notificationchannel, "Warning", "ProcessingError", issue.Error())
	notificationchannel.Status.Phase = grafanav1alpha1.PhaseFailing
	notificationchannel.Status.Message = issue.Error()

	err := r.client.Status().Update(r.context, notificationchannel)
	if err != nil {
		// Ignore conclicts. Resource might just be outdated.
		if errors.IsConflict(err) {
			return
		}
		log.Error(err, "error updating notificationchannel status")
	}
}

// Get an authenticated grafana API client
func (r *ReconcileGrafanaNotificationChannel) getClient() (GrafanaClient, error) {
	url := r.state.AdminUrl
	if url == "" {
		return nil, defaultErrors.New("cannot get grafana admin url")
	}

	username := r.state.AdminUsername
	if username == "" {
		return nil, defaultErrors.New("invalid credentials (username)")
	}

	password := r.state.AdminPassword
	if password == "" {
		return nil, defaultErrors.New("invalid credentials (password)")
	}

	duration := time.Duration(r.state.ClientTimeout)
	return NewGrafanaClient(url, username, password, duration), nil
}

// Test if a given notificationchannel matches an array of label selectors
func (r *ReconcileGrafanaNotificationChannel) isMatch(item *grafanav1alpha1.GrafanaNotificationChannel) bool {
	if r.state.DashboardSelectors == nil {
		return false
	}

	match, err := item.MatchesSelectors(r.state.DashboardSelectors)
	if err != nil {
		log.Error(err, fmt.Sprintf("error matching selectors against %v/%v",
			item.Namespace,
			item.Name))
		return false
	}
	return match
}
