package config

import (
	"testing"
	"time"

	"github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestControllerConfig_ConcurrentlyReadAndWritePlugins(t *testing.T) {
	go func() {
		// Continuously read all Plugins
		for {
			_ = GetControllerConfig().GetAllPlugins()
		}
	}()

	go func() {
		// Continuously overwrite existing Plugins
		for {
			c := GetControllerConfig()

			d := &v1alpha1.GrafanaDashboard{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "crash",
				},
				Spec: v1alpha1.GrafanaDashboardSpec{
					Plugins: []v1alpha1.GrafanaPlugin{
						{Name: "one", Version: "0"},
						{Name: "two", Version: "0"},
						{Name: "tttt", Version: "0"},
						{Name: "four", Version: "0"},
					},
				},
			}
			c.SetPluginsFor(d)
		}
	}()

	// it will take less then a second to panic if read and write happen at the same time
	time.Sleep(time.Second)
}
