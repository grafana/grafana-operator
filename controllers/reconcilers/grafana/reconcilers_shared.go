package grafana

import (
	"fmt"
	"time"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func setInvalidMergeCondition(cr *v1beta1.Grafana, object string, err error) {
	meta.SetStatusCondition(&cr.Status.Conditions, metav1.Condition{
		Type:               fmt.Sprintf("Invalid%sOverride", object),
		Reason:             fmt.Sprintf("MergingCustom%s", object),
		Message:            err.Error(),
		Status:             metav1.ConditionTrue,
		ObservedGeneration: cr.Generation,
		LastTransitionTime: metav1.Time{Time: time.Now()},
	})
}

func removeInvalidMergeCondition(cr *v1beta1.Grafana, object string) {
	meta.RemoveStatusCondition(&cr.Status.Conditions, fmt.Sprintf("Invalid%sOverride", object))
}
