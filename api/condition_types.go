/*
Copyright 2022.

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

package api

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WithConditions interface {
	GetGeneration() int64
	GetConditions() []metav1.Condition
	SetConditions(conditions []metav1.Condition)
}

func SetCondition(obj WithConditions, newCond metav1.Condition) bool {
	replaced := false
	newConds := make([]metav1.Condition, 0, len(obj.GetConditions()))
	newCond.LastTransitionTime = metav1.Now()
	newCond.ObservedGeneration = obj.GetGeneration()
	for _, cond := range obj.GetConditions() {
		if cond.Type == newCond.Type {
			if cond.Status == newCond.Status && cond.Reason == newCond.Reason && cond.Message == newCond.Message {
				return false
			}
			newConds = append(newConds, newCond)
			replaced = true
		} else {
			newConds = append(newConds, cond)
		}
	}

	if !replaced {
		newConds = append(newConds, newCond)
	}

	obj.SetConditions(newConds)
	return true
}

func SetReadyCondition(obj WithConditions, status metav1.ConditionStatus, reason string, message string) bool {
	return SetCondition(obj, metav1.Condition{
		Type:               "Ready",
		Status:             status,
		Reason:             reason,
		LastTransitionTime: metav1.Now(),
		ObservedGeneration: obj.GetGeneration(),
		Message:            message,
	})
}

func GetReadyCondition(obj WithConditions) *metav1.Condition {
	for _, cond := range obj.GetConditions() {
		if cond.Type == "Ready" {
			return &cond
		}
	}
	return nil
}
