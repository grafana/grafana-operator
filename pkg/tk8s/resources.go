package tk8s

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ctrl "sigs.k8s.io/controller-runtime"
)

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
