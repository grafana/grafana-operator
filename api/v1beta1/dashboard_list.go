package v1beta1

import "encoding/json"

type NamespacedResource struct {
	Name string `json:"name"`
	UID  string `json:"uid"`
}

type NamespacedResources map[string][]NamespacedResource

func (in *NamespacedResources) Serialize() ([]byte, error) {
	return json.Marshal(in)
}

func (in *NamespacedResources) Deserialize(from string) {
	if from == "" {
		return
	}
	json.Unmarshal([]byte(from), in)
}

func (in *NamespacedResources) Find(namespace string, name string) (bool, string) {
	if val, ok := (*in)[namespace]; ok {
		for _, dashboard := range val {
			if dashboard.Name == name {
				return true, dashboard.UID
			}
		}
	}
	return false, ""
}

func (in *NamespacedResources) ForNamespace(namespace string) []NamespacedResource {
	return (*in)[namespace]
}

func (in *NamespacedResources) AddResource(namespace string, name string, uid string) NamespacedResources {
	copy := NamespacedResources{}
	for ns, dashboards := range *in {
		copy[ns] = dashboards
	}
	if _, ok := copy[namespace]; !ok {
		copy[namespace] = []NamespacedResource{}
	}
	for _, dashboard := range copy[namespace] {
		if dashboard.UID == uid {
			return copy
		}
	}
	copy[namespace] = append(copy[namespace], NamespacedResource{
		Name: name,
		UID:  uid,
	})
	return copy
}

func (in NamespacedResources) RemoveResource(namespace string, name string) NamespacedResources {
	copy := NamespacedResources{}
	for ns, dashboards := range in {
		copy[ns] = []NamespacedResource{}
		for _, dashboard := range dashboards {
			if ns == namespace && dashboard.Name == name {
				continue
			}
			copy[ns] = append(copy[ns], dashboard)
		}
	}
	return copy
}
