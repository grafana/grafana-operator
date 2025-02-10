package v1beta1

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GrafanaLibraryPanelSpec defines the desired state of GrafanaLibraryPanel
// +kubebuilder:validation:XValidation:rule="(has(self.folderUID) && !(has(self.folderRef))) || (has(self.folderRef) && !(has(self.folderUID))) || !(has(self.folderRef) && (has(self.folderUID)))", message="Only one of folderUID or folderRef can be declared at the same time"
// +kubebuilder:validation:XValidation:rule="(has(self.folder) && !(has(self.folderRef) || has(self.folderUID))) || !(has(self.folder))", message="folder field cannot be set when folderUID or folderRef is already declared"
// +kubebuilder:validation:XValidation:rule="((!has(oldSelf.uid) && !has(self.uid)) || (has(oldSelf.uid) && has(self.uid)))", message="spec.uid is immutable"
type GrafanaLibraryPanelSpec struct {
	GrafanaCommonSpec  `json:",inline"`
	GrafanaContentSpec `json:",inline"`

	// folder assignment for dashboard
	// +optional
	FolderTitle string `json:"folder,omitempty"`

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

	// The instanceSelector can't find matching grafana instances
	NoMatchingInstances bool `json:"NoMatchingInstances,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GrafanaLibraryPanel is the Schema for the grafanalibrarypanels API
// +kubebuilder:printcolumn:name="No matching instances",type="boolean",JSONPath=".status.NoMatchingInstances",description=""
// +kubebuilder:printcolumn:name="Last resync",type="date",format="date-time",JSONPath=".status.lastResync",description=""
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""
type GrafanaLibraryPanel struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaLibraryPanelSpec   `json:"spec,omitempty"`
	Status GrafanaLibraryPanelStatus `json:"status,omitempty"`
}

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

func (in *GrafanaLibraryPanel) IsAllowCrossNamespaceImport() bool {
	return in.Spec.AllowCrossNamespaceImport
}

func (in *GrafanaLibraryPanel) ResyncPeriodHasElapsed() bool {
	deadline := in.Status.LastResync.Add(in.Spec.ResyncPeriod.Duration)
	return time.Now().After(deadline)
}

// GrafanaContentSpec implements GrafanaContentResource
func (in *GrafanaLibraryPanel) GrafanaContentSpec() *GrafanaContentSpec {
	return &in.Spec.GrafanaContentSpec
}

// GrafanaContentSpec implements GrafanaContentResource
func (in *GrafanaLibraryPanel) GrafanaContentStatus() *GrafanaContentStatus {
	return &in.Status.GrafanaContentStatus
}

func (in *GrafanaLibraryPanelList) Find(namespace string, name string) *GrafanaLibraryPanel {
	for _, e := range in.Items {
		if e.Namespace == namespace && e.Name == name {
			return &e
		}
	}
	return nil
}

func init() {
	SchemeBuilder.Register(&GrafanaLibraryPanel{}, &GrafanaLibraryPanelList{})
}
