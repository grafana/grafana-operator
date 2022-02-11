package grafanadashboard

import (
	"testing"

	"github.com/stretchr/testify/require"

	grafanav1alpha1 "github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var dashboard = &grafanav1alpha1.GrafanaDashboard{
	TypeMeta: metav1.TypeMeta{},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "dashboard2",
		Namespace: "grafana",
	},
	Spec: grafanav1alpha1.GrafanaDashboardSpec{
		Url: "url1",
	},
}

var knownDashboardsEmpty = []*grafanav1alpha1.GrafanaDashboardRef{
	{
		Name: "dashboard1",
	},
	{
		Name: "dashboard2",
	},
}

var knownDashboards = []*grafanav1alpha1.GrafanaDashboardRef{
	{
		Name:      "dashboard1",
		Namespace: "grafana",
		Hash:      "1234",
		UID:       "uid1234",
	},
	{
		Name:      "dashboard2",
		Namespace: "grafana",
		Hash:      "5678",
		UID:       "uid5678",
	},
}

var dashboarList = &grafanav1alpha1.GrafanaDashboardList{
	TypeMeta: metav1.TypeMeta{},
	ListMeta: metav1.ListMeta{},
	Items: []grafanav1alpha1.GrafanaDashboard{
		{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dashboard1",
				Namespace: "grafana",
			},
			Spec: grafanav1alpha1.GrafanaDashboardSpec{
				Url: "url1",
			},
		},
		{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dashboard2",
				Namespace: "foo",
			},
			Spec: grafanav1alpha1.GrafanaDashboardSpec{
				Url: "url2",
			},
		},
	},
}

func TestFindHashEmpty(t *testing.T) {
	output := findHash(knownDashboardsEmpty, dashboard)
	require.Equal(t, output, "")
}

func TestFindHash(t *testing.T) {
	output := findHash(knownDashboards, dashboard)
	require.Equal(t, output, "5678")
}

func TestTrueInNamespace(t *testing.T) {
	status := inNamespace(dashboarList, knownDashboards[0])
	require.Equal(t, status, true)
}

func TestFalseInNamespace(t *testing.T) {
	status := inNamespace(dashboarList, knownDashboards[1])
	require.Equal(t, status, false)
}

func TestFindUidEmpty(t *testing.T) {
	output := findUid(knownDashboardsEmpty, dashboard)
	require.Equal(t, output, "")
}

func TestFindUid(t *testing.T) {
	output := findUid(knownDashboards, dashboard)
	require.Equal(t, output, "uid5678")
}
