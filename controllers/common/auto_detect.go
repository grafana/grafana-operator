package common

import (
	"github.com/grafana-operator/grafana-operator/v4/internal/k8sutil"
	routev1 "github.com/openshift/api/route/v1"

	config2 "github.com/grafana-operator/grafana-operator/v4/controllers/config"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"time"

	//routev1 "github.com/openshift/api/route/v1"
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
	SubscriptionChannel schema.GroupVersionKind
}

// New creates a new auto-detect runner
func NewAutoDetect(mgr manager.Manager) (*Background, error) {
	dc, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
	if err != nil {
		return nil, err
	}

	return &Background{dc: dc, client: mgr.GetClient()}, nil
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
}

func (b *Background) autoDetectCapabilities() {
	b.detectRoute()
}

//
func (b *Background) detectRoute() {
	resourceExists, err := k8sutil.ResourceExists(b.dc, routev1.SchemeGroupVersion.String(), RouteKind)
	if resourceExists && err == nil {
		config := config2.GetControllerConfig()
		config.AddConfigItem(config2.ConfigOpenshift, true)

		b.SubscriptionChannel = routev1.SchemeGroupVersion.WithKind(RouteKind)
	}
}
