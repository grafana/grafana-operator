package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GrafanaLibraryPanelSpec defines the desired state of GrafanaLibraryPanel
// +kubebuilder:validation:XValidation:rule="(has(self.folderUID) && !(has(self.folderRef))) || (has(self.folderRef) && !(has(self.folderUID))) || !(has(self.folderRef) && (has(self.folderUID)))", message="Only one of folderUID or folderRef can be declared at the same time"
// +kubebuilder:validation:XValidation:rule="((!has(oldSelf.uid) && !has(self.uid)) || (has(oldSelf.uid) && has(self.uid)))", message="spec.uid is immutable"
type GrafanaLibraryPanelSpec struct {
	GrafanaCommonSpec  `json:",inline"`
	GrafanaContentSpec `json:",inline"`

	// UID of the target folder for this dashboard
	// +optional
	FolderUID string `json:"folderUID,omitempty"`

	// Name of a `GrafanaFolder` resource in the same namespace
	// +optional
	FolderRef string `json:"folderRef,omitempty"`

	// plugins
	// +optional
	Plugins PluginList `json:"plugins,omitempty"`
}

// GrafanaLibraryPanelStatus defines the observed state of GrafanaLibraryPanel
type GrafanaLibraryPanelStatus struct {
	GrafanaCommonStatus  `json:",inline"`
	GrafanaContentStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GrafanaLibraryPanel is the Schema for the grafanalibrarypanels API
// +kubebuilder:printcolumn:name="Last resync",type="date",format="date-time",JSONPath=".status.lastResync",description=""
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""
// +kubebuilder:resource:categories={grafana-operator}
type GrafanaLibraryPanel struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaLibraryPanelSpec   `json:"spec"`
	Status GrafanaLibraryPanelStatus `json:"status,omitempty"`
}

var _ CommonResource = (*GrafanaLibraryPanel)(nil)

//+kubebuilder:object:root=true

// GrafanaLibraryPanelList contains a list of GrafanaLibraryPanel
type GrafanaLibraryPanelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrafanaLibraryPanel `json:"items"`
}

// FolderRef implements FolderReferencer.
func (in *GrafanaLibraryPanel) FolderRef() string {
	return in.Spec.FolderRef
}

// FolderUID implements FolderReferencer.
func (in *GrafanaLibraryPanel) FolderUID() string {
	return in.Spec.FolderUID
}

// FolderNamespace implements FolderReferencer.
func (in *GrafanaLibraryPanel) FolderNamespace() string {
	return in.Namespace
}

// Conditions implements FolderReferencer.
func (in *GrafanaLibraryPanel) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

// CurrentGeneration implements FolderReferencer.
func (in *GrafanaLibraryPanel) CurrentGeneration() int64 {
	return in.Generation
}

func (in *GrafanaLibraryPanel) MatchLabels() *metav1.LabelSelector {
	return in.Spec.InstanceSelector
}

func (in *GrafanaLibraryPanel) MatchNamespace() string {
	return in.Namespace
}

func (in *GrafanaLibraryPanel) AllowCrossNamespace() bool {
	return in.Spec.AllowCrossNamespaceImport
}

func (in *GrafanaLibraryPanel) CommonStatus() *GrafanaCommonStatus {
	return &in.Status.GrafanaCommonStatus
}

// GrafanaContentSpec implements GrafanaContentResource
func (in *GrafanaLibraryPanel) GrafanaContentSpec() *GrafanaContentSpec {
	return &in.Spec.GrafanaContentSpec
}

// GrafanaContentSpec implements GrafanaContentResource
func (in *GrafanaLibraryPanel) GrafanaContentStatus() *GrafanaContentStatus {
	return &in.Status.GrafanaContentStatus
}

func (in *GrafanaLibraryPanel) NamespacedResource(uid string) NamespacedResource {
	// Not enough context to call content.CustomUIDOrUID(uid).
	// Hence, use uid from args as the caller has more context
	return NewNamespacedResource(in.Namespace, in.Name, uid)
}

var _ GrafanaContentResource = &GrafanaLibraryPanel{}

func (in *GrafanaLibraryPanelList) Exists(namespace, name string) bool {
	for _, item := range in.Items {
		if item.Namespace == namespace && item.Name == name {
			return true
		}
	}

	return false
}

func init() {
	SchemeBuilder.Register(&GrafanaLibraryPanel{}, &GrafanaLibraryPanelList{})
}
