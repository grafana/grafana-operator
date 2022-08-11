package v1beta1

import "encoding/json"

type NamespacedDashboard struct {
	Name string `json:"name"`
	UID  string `json:"uid"`
}

type NamespacedDashboards map[string][]NamespacedDashboard

func (in *NamespacedDashboards) Serialize() ([]byte, error) {
	return json.Marshal(in)
}

func (in *NamespacedDashboards) Deserialize(from string) {
	if from == "" {
		return
	}
	json.Unmarshal([]byte(from), in)
}

func (in *NamespacedDashboards) Find(namespace string, name string) (bool, string) {
	if val, ok := (*in)[namespace]; ok {
		for _, dashboard := range val {
			if dashboard.Name == name {
				return true, dashboard.UID
			}
		}
	}
	return false, ""
}

func (in *NamespacedDashboards) ForNamespace(namespace string) []NamespacedDashboard {
	return (*in)[namespace]
}

func (in *NamespacedDashboards) AddDashboard(namespace string, name string, uid string) NamespacedDashboards {
	copy := NamespacedDashboards{}
	for ns, dashboards := range *in {
		copy[ns] = dashboards
	}
	if _, ok := copy[namespace]; !ok {
		copy[namespace] = []NamespacedDashboard{}
	}
	for _, dashboard := range copy[namespace] {
		if dashboard.UID == uid {
			return copy
		}
	}
	copy[namespace] = append(copy[namespace], NamespacedDashboard{
		Name: name,
		UID:  uid,
	})
	return copy
}

func (in NamespacedDashboards) RemoveDashboard(namespace string, name string) NamespacedDashboards {
	copy := NamespacedDashboards{}
	for ns, dashboards := range in {
		copy[ns] = []NamespacedDashboard{}
		for _, dashboard := range dashboards {
			if ns == namespace && dashboard.Name == name {
				continue
			}
			copy[ns] = append(copy[ns], dashboard)
		}
	}
	return copy
}
