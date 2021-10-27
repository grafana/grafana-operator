package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/require"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var labelSelector = &metav1.LabelSelector{
	MatchExpressions: []metav1.LabelSelectorRequirement{
		{
			Key:      "app",
			Operator: "In",
			Values:   []string{"grafana"},
		},
	},
}

var listLabelSelector = []*metav1.LabelSelector{
	{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      "ham",
				Operator: "In",
				Values:   []string{"eggs"},
			},
		},
	},
	{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      "app",
				Operator: "NotIn",
				Values:   []string{"grafana"},
			},
		},
	},
}

func TestDashboardMatchesSelectorFalse(t *testing.T) {
	i := &GrafanaDashboard{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dashboard",
			Namespace: "grafana",
			Labels: map[string]string{
				"foo": "bar",
			},
		},
		Spec: GrafanaDashboardSpec{},
	}
	output, err := i.matchesSelector(labelSelector)
	require.NoError(t, err)
	require.Equal(t, output, false)
}

func TestDashboardMatchesSelectorTrue(t *testing.T) {
	i := &GrafanaDashboard{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dashboard",
			Namespace: "grafana",
			Labels: map[string]string{
				"app": "grafana",
			},
		},
		Spec: GrafanaDashboardSpec{},
	}
	output, err := i.matchesSelector(labelSelector)
	require.NoError(t, err)
	require.Equal(t, output, true)
}

func TestDashboardMatchesSelectorsMultipleNotIn(t *testing.T) {
	i := &GrafanaDashboard{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dashboard",
			Namespace: "grafana",
			Labels: map[string]string{
				"foo": "bar",
				"app": "grafana",
			},
		},
		Spec: GrafanaDashboardSpec{},
	}
	output, err := i.MatchesSelectors(listLabelSelector)
	require.NoError(t, err)
	require.Equal(t, output, false)
}

func TestMatchesDashboardSelectorsMultipleTrue(t *testing.T) {
	i := &GrafanaDashboard{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dashboard",
			Namespace: "grafana",
			Labels: map[string]string{
				"foo": "bar",
			},
		},
		Spec: GrafanaDashboardSpec{},
	}
	output, err := i.MatchesSelectors(listLabelSelector)
	require.NoError(t, err)
	require.Equal(t, output, true)
}
