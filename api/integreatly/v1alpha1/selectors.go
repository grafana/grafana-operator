/*
Copyright 2021.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (d *GrafanaDashboard) matchesSelector(s *metav1.LabelSelector) (bool, error) {
	selector, err := metav1.LabelSelectorAsSelector(s)
	if err != nil {
		return false, err
	}

	return selector.Empty() || selector.Matches(labels.Set(d.Labels)), nil
}

// Check if the dashboard matches at least one of the selectors
func (d *GrafanaDashboard) MatchesSelectors(s []*metav1.LabelSelector) (bool, error) {
	result := false

	for _, selector := range s {
		match, err := d.matchesSelector(selector)
		if err != nil {
			return false, err
		}

		result = result || match
	}

	return result, nil
}

func (d *GrafanaFolder) matchesSelector(s *metav1.LabelSelector) (bool, error) {
	selector, err := metav1.LabelSelectorAsSelector(s)
	if err != nil {
		return false, err
	}

	return selector.Empty() || selector.Matches(labels.Set(d.Labels)), nil
}

// Check if the dashboard-folder matches at least one of the selectors
func (d *GrafanaFolder) MatchesSelectors(s []*metav1.LabelSelector) (bool, error) {
	result := false

	for _, selector := range s {
		match, err := d.matchesSelector(selector)
		if err != nil {
			return false, err
		}

		result = result || match
	}

	return result, nil
}

func (d *GrafanaNotificationChannel) matchesSelector(s *metav1.LabelSelector) (bool, error) {
	selector, err := metav1.LabelSelectorAsSelector(s)
	if err != nil {
		return false, err
	}

	return selector.Empty() || selector.Matches(labels.Set(d.Labels)), nil
}

// Check if the notification channel matches at least one of the selectors
func (d *GrafanaNotificationChannel) MatchesSelectors(s []*metav1.LabelSelector) (bool, error) {
	result := false

	for _, selector := range s {
		match, err := d.matchesSelector(selector)
		if err != nil {
			return false, err
		}

		result = result || match
	}

	return result, nil
}
