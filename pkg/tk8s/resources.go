package tk8s

import (
	"slices"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type resourceWithConditions interface {
	Conditions() *[]metav1.Condition
}

func GetRequest(t tHelper, obj client.Object) ctrl.Request {
	t.Helper()

	v := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      obj.GetName(),
			Namespace: obj.GetNamespace(),
		},
	}

	return v
}

func GetRequestKey(t tHelper, obj client.Object) types.NamespacedName {
	t.Helper()

	v := types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}

	return v
}

func HasCondition(t tHelper, cr resourceWithConditions, target metav1.Condition) bool {
	t.Helper()

	isFound := slices.ContainsFunc(*cr.Conditions(), func(c metav1.Condition) bool {
		return c.Type == target.Type && c.Reason == target.Reason
	})

	return isFound
}
