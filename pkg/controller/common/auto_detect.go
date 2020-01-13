package common

import (
	config2 "github.com/integr8ly/grafana-operator/v3/pkg/controller/config"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"time"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// Route kind is not provided by the openshift api
const (
	RouteKind = "Route"
)

// Background represents a procedure that runs in the background, periodically auto-detecting features
type Background struct {
	client              client.Client
	dc                  discovery.DiscoveryInterface
	ticker              *time.Ticker
	SubscriptionChannel chan schema.GroupVersionKind
}

// New creates a new auto-detect runner
func NewAutoDetect(mgr manager.Manager) (*Background, error) {
	dc, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
	if err != nil {
		return nil, err
	}

	// Create a new channel that GVK type will be sent down
	// The subscription channel can be used in the future to
	// implement actions that are dependant on certain resources
	// being installed on the cluster
	subChan := make(chan schema.GroupVersionKind, 1)

	return &Background{dc: dc, client: mgr.GetClient(), SubscriptionChannel: subChan}, nil
}

// Start initializes the auto-detection process that runs in the background
func (b *Background) Start() {
	// periodically attempts to auto detect all the capabilities for this operator
	b.ticker = time.NewTicker(5 * time.Second)

	go func() {
		b.autoDetectCapabilities()

		for range b.ticker.C {
			b.autoDetectCapabilities()
		}
	}()
}

// Stop causes the background process to stop auto detecting capabilities
func (b *Background) Stop() {
	b.ticker.Stop()
	close(b.SubscriptionChannel)
}

func (b *Background) autoDetectCapabilities() {
	b.detectRoute()
}

func (b *Background) detectRoute() {
	resourceExists, _ := k8sutil.ResourceExists(b.dc, routev1.SchemeGroupVersion.String(), RouteKind)
	if resourceExists {
		config := config2.GetControllerConfig()
		config.AddConfigItem(config2.ConfigOpenshift, true)

		b.SubscriptionChannel <- routev1.SchemeGroupVersion.WithKind(RouteKind)
	}
}
