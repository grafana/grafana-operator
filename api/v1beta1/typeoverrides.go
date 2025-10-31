package v1beta1

import (
	"encoding/json"
	"fmt"
	"maps"
	"reflect"

	"k8s.io/apimachinery/pkg/util/intstr"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// +kubebuilder:object:generate=true

// ObjectMeta contains only a [subset of the fields included in k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta).
type ObjectMeta struct {
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// Merge merges it's receivers values into the incoming ObjectMeta by overwriting values for existing keys and adding new ones.
func (override *ObjectMeta) Merge(meta metav1.ObjectMeta) metav1.ObjectMeta {
	if override == nil {
		return meta
	}

	if len(override.Annotations) > 0 {
		if meta.Annotations == nil {
			meta.Annotations = make(map[string]string)
		}

		maps.Copy(meta.Annotations, override.Annotations)
	}

	if len(override.Labels) > 0 {
		if meta.Labels == nil {
			meta.Labels = make(map[string]string)
		}

		maps.Copy(meta.Labels, override.Labels)
	}

	return meta
}

// +kubebuilder:object:generate=true

type DeploymentV1 struct {
	ObjectMeta ObjectMeta       `json:"metadata,omitempty"`
	Spec       DeploymentV1Spec `json:"spec,omitempty"`
}

type DeploymentV1Spec struct {
	// +optional
	Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,1,opt,name=replicas"`

	Selector *metav1.LabelSelector `json:"selector,omitempty" protobuf:"bytes,2,opt,name=selector"`

	Template *DeploymentV1PodTemplateSpec `json:"template,omitempty" protobuf:"bytes,3,opt,name=template"`

	// +optional
	// +patchStrategy=retainKeys
	Strategy *appsv1.DeploymentStrategy `json:"strategy,omitempty" patchStrategy:"retainKeys" protobuf:"bytes,4,opt,name=strategy"`

	// +optional
	MinReadySeconds int32 `json:"minReadySeconds,omitempty" protobuf:"varint,5,opt,name=minReadySeconds"`

	// +optional
	RevisionHistoryLimit *int32 `json:"revisionHistoryLimit,omitempty" protobuf:"varint,6,opt,name=revisionHistoryLimit"`

	// +optional
	Paused bool `json:"paused,omitempty" protobuf:"varint,7,opt,name=paused"`

	ProgressDeadlineSeconds *int32 `json:"progressDeadlineSeconds,omitempty" protobuf:"varint,9,opt,name=progressDeadlineSeconds"`
}

type DeploymentV1PodTemplateSpec struct {
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Specification of the desired behavior of the pod.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Spec *DeploymentV1PodSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

type DeploymentV1PodSpec struct {
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	Volumes []corev1.Volume `json:"volumes,omitempty" patchStrategy:"merge,retainKeys" patchMergeKey:"name" protobuf:"bytes,1,rep,name=volumes"`
	// +patchMergeKey=name
	// +patchStrategy=merge
	// +optional
	InitContainers []corev1.Container `json:"initContainers,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,20,rep,name=initContainers"`
	// +patchMergeKey=name
	// +patchStrategy=merge
	// +optional
	Containers []corev1.Container `json:"containers,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,2,rep,name=containers"`
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	EphemeralContainers []corev1.EphemeralContainer `json:"ephemeralContainers,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,34,rep,name=ephemeralContainers"`
	// +optional
	RestartPolicy corev1.RestartPolicy `json:"restartPolicy,omitempty" protobuf:"bytes,3,opt,name=restartPolicy,casttype=RestartPolicy"`
	// +optional
	TerminationGracePeriodSeconds *int64 `json:"terminationGracePeriodSeconds,omitempty" protobuf:"varint,4,opt,name=terminationGracePeriodSeconds"`
	// +optional
	ActiveDeadlineSeconds *int64 `json:"activeDeadlineSeconds,omitempty" protobuf:"varint,5,opt,name=activeDeadlineSeconds"`
	// +optional
	DNSPolicy corev1.DNSPolicy `json:"dnsPolicy,omitempty" protobuf:"bytes,6,opt,name=dnsPolicy,casttype=DNSPolicy"`
	// NodeSelector is a selector which must be true for the pod to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	// More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
	// +optional
	// +mapType=atomic
	NodeSelector map[string]string `json:"nodeSelector,omitempty" protobuf:"bytes,7,rep,name=nodeSelector"`

	// ServiceAccountName is the name of the ServiceAccount to use to run this pod.
	// More info: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty" protobuf:"bytes,8,opt,name=serviceAccountName"`
	// DeprecatedServiceAccount is a depreciated alias for ServiceAccountName.
	// Deprecated: Use serviceAccountName instead.
	// +k8s:conversion-gen=false
	// +optional
	DeprecatedServiceAccount string `json:"serviceAccount,omitempty" protobuf:"bytes,9,opt,name=serviceAccount"`
	// AutomountServiceAccountToken indicates whether a service account token should be automatically mounted.
	// +optional
	AutomountServiceAccountToken *bool `json:"automountServiceAccountToken,omitempty" protobuf:"varint,21,opt,name=automountServiceAccountToken"`

	// NodeName is a request to schedule this pod onto a specific node. If it is non-empty,
	// the scheduler simply schedules this pod onto that node, assuming that it fits resource
	// requirements.
	// +optional
	NodeName string `json:"nodeName,omitempty" protobuf:"bytes,10,opt,name=nodeName"`
	// Host networking requested for this pod. Use the host's network namespace.
	// If this option is set, the ports that will be used must be specified.
	// Default to false.
	// +k8s:conversion-gen=false
	// +optional
	HostNetwork bool `json:"hostNetwork,omitempty" protobuf:"varint,11,opt,name=hostNetwork"`
	// Use the host's pid namespace.
	// Optional: Default to false.
	// +k8s:conversion-gen=false
	// +optional
	HostPID bool `json:"hostPID,omitempty" protobuf:"varint,12,opt,name=hostPID"`
	// Use the host's ipc namespace.
	// Optional: Default to false.
	// +k8s:conversion-gen=false
	// +optional
	HostIPC bool `json:"hostIPC,omitempty" protobuf:"varint,13,opt,name=hostIPC"`
	// Share a single process namespace between all of the containers in a pod.
	// When this is set containers will be able to view and signal processes from other containers
	// in the same pod, and the first process in each container will not be assigned PID 1.
	// HostPID and ShareProcessNamespace cannot both be set.
	// Optional: Default to false.
	// +k8s:conversion-gen=false
	// +optional
	ShareProcessNamespace *bool `json:"shareProcessNamespace,omitempty" protobuf:"varint,27,opt,name=shareProcessNamespace"`
	// SecurityContext holds pod-level security attributes and common container settings.
	// Optional: Defaults to empty.  See type description for default values of each field.
	// +optional
	SecurityContext *corev1.PodSecurityContext `json:"securityContext,omitempty" protobuf:"bytes,14,opt,name=securityContext"`
	// ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.
	// If specified, these secrets will be passed to individual puller implementations for them to use.
	// More info: https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,15,rep,name=imagePullSecrets"`
	// Specifies the hostname of the Pod
	// If not specified, the pod's hostname will be set to a system-defined value.
	// +optional
	Hostname string `json:"hostname,omitempty" protobuf:"bytes,16,opt,name=hostname"`
	// If specified, the fully qualified Pod hostname will be "<hostname>.<subdomain>.<pod namespace>.svc.<cluster domain>".
	// If not specified, the pod will not have a domainname at all.
	// +optional
	Subdomain string `json:"subdomain,omitempty" protobuf:"bytes,17,opt,name=subdomain"`
	// If specified, the pod's scheduling constraints
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty" protobuf:"bytes,18,opt,name=affinity"`
	// If specified, the pod will be dispatched by specified scheduler.
	// If not specified, the pod will be dispatched by default scheduler.
	// +optional
	SchedulerName string `json:"schedulerName,omitempty" protobuf:"bytes,19,opt,name=schedulerName"`
	// If specified, the pod's tolerations.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty" protobuf:"bytes,22,opt,name=tolerations"`
	// HostAliases is an optional list of hosts and IPs that will be injected into the pod's hosts
	// file if specified. This is only valid for non-hostNetwork pods.
	// +optional
	// +patchMergeKey=ip
	// +patchStrategy=merge
	HostAliases []corev1.HostAlias `json:"hostAliases,omitempty" patchStrategy:"merge" patchMergeKey:"ip" protobuf:"bytes,23,rep,name=hostAliases"`
	// If specified, indicates the pod's priority. "system-node-critical" and
	// "system-cluster-critical" are two special keywords which indicate the
	// highest priorities with the former being the highest priority. Any other
	// name must be defined by creating a PriorityClass object with that name.
	// If not specified, the pod priority will be default or zero if there is no
	// default.
	// +optional
	PriorityClassName string `json:"priorityClassName,omitempty" protobuf:"bytes,24,opt,name=priorityClassName"`
	// The priority value. Various system components use this field to find the
	// priority of the pod. When Priority Admission Controller is enabled, it
	// prevents users from setting this field. The admission controller populates
	// this field from PriorityClassName.
	// The higher the value, the higher the priority.
	// +optional
	Priority *int32 `json:"priority,omitempty" protobuf:"bytes,25,opt,name=priority"`
	// Specifies the DNS parameters of a pod.
	// Parameters specified here will be merged to the generated DNS
	// configuration based on DNSPolicy.
	// +optional
	DNSConfig *corev1.PodDNSConfig `json:"dnsConfig,omitempty" protobuf:"bytes,26,opt,name=dnsConfig"`
	// If specified, all readiness gates will be evaluated for pod readiness.
	// A pod is ready when all its containers are ready AND
	// all conditions specified in the readiness gates have status equal to "True"
	// More info: https://git.k8s.io/enhancements/keps/sig-network/580-pod-readiness-gates
	// +optional
	ReadinessGates []corev1.PodReadinessGate `json:"readinessGates,omitempty" protobuf:"bytes,28,opt,name=readinessGates"`
	// RuntimeClassName refers to a RuntimeClass object in the node.k8s.io group, which should be used
	// to run this pod.  If no RuntimeClass resource matches the named class, the pod will not be run.
	// If unset or empty, the "legacy" RuntimeClass will be used, which is an implicit class with an
	// empty definition that uses the default runtime handler.
	// More info: https://git.k8s.io/enhancements/keps/sig-node/585-runtime-class
	// +optional
	RuntimeClassName *string `json:"runtimeClassName,omitempty" protobuf:"bytes,29,opt,name=runtimeClassName"`
	// EnableServiceLinks indicates whether information about services should be injected into pod's
	// environment variables, matching the syntax of Docker links.
	// Optional: Defaults to true.
	// +optional
	EnableServiceLinks *bool `json:"enableServiceLinks,omitempty" protobuf:"varint,30,opt,name=enableServiceLinks"`
	// PreemptionPolicy is the Policy for preempting pods with lower priority.
	// One of Never, PreemptLowerPriority.
	// Defaults to PreemptLowerPriority if unset.
	// +optional
	PreemptionPolicy *corev1.PreemptionPolicy `json:"preemptionPolicy,omitempty" protobuf:"bytes,31,opt,name=preemptionPolicy"`
	// Overhead represents the resource overhead associated with running a pod for a given RuntimeClass.
	// This field will be autopopulated at admission time by the RuntimeClass admission controller. If
	// the RuntimeClass admission controller is enabled, overhead must not be set in Pod create requests.
	// The RuntimeClass admission controller will reject Pod create requests which have the overhead already
	// set. If RuntimeClass is configured and selected in the PodSpec, Overhead will be set to the value
	// defined in the corresponding RuntimeClass, otherwise it will remain unset and treated as zero.
	// More info: https://git.k8s.io/enhancements/keps/sig-node/688-pod-overhead/README.md
	// +optional
	Overhead corev1.ResourceList `json:"overhead,omitempty" protobuf:"bytes,32,opt,name=overhead"`
	// TopologySpreadConstraints describes how a group of pods ought to spread across topology
	// domains. Scheduler will schedule pods in a way which abides by the constraints.
	// All topologySpreadConstraints are ANDed.
	// +optional
	// +patchMergeKey=topologyKey
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=topologyKey
	// +listMapKey=whenUnsatisfiable
	TopologySpreadConstraints []corev1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty" patchStrategy:"merge" patchMergeKey:"topologyKey" protobuf:"bytes,33,opt,name=topologySpreadConstraints"`
	// If true the pod's hostname will be configured as the pod's FQDN, rather than the leaf name (the default).
	// In Linux containers, this means setting the FQDN in the hostname field of the kernel (the nodename field of struct utsname).
	// In Windows containers, this means setting the registry value of hostname for the registry key HKEY_LOCAL_MACHINE\\SYSTEM\\CurrentControlSet\\Services\\Tcpip\\Parameters to FQDN.
	// If a pod does not have FQDN, this has no effect.
	// Default to false.
	// +optional
	SetHostnameAsFQDN *bool `json:"setHostnameAsFQDN,omitempty" protobuf:"varint,35,opt,name=setHostnameAsFQDN"`
	// Specifies the OS of the containers in the pod.
	// Some pod and container fields are restricted if this is set.
	//
	// If the OS field is set to linux, the following fields must be unset:
	// -securityContext.windowsOptions
	//
	// If the OS field is set to windows, following fields must be unset:
	// - spec.hostPID
	// - spec.hostIPC
	// - spec.hostUsers
	// - spec.securityContext.seLinuxOptions
	// - spec.securityContext.seccompProfile
	// - spec.securityContext.fsGroup
	// - spec.securityContext.fsGroupChangePolicy
	// - spec.securityContext.sysctls
	// - spec.shareProcessNamespace
	// - spec.securityContext.runAsUser
	// - spec.securityContext.runAsGroup
	// - spec.securityContext.supplementalGroups
	// - spec.containers[*].securityContext.seLinuxOptions
	// - spec.containers[*].securityContext.seccompProfile
	// - spec.containers[*].securityContext.capabilities
	// - spec.containers[*].securityContext.readOnlyRootFilesystem
	// - spec.containers[*].securityContext.privileged
	// - spec.containers[*].securityContext.allowPrivilegeEscalation
	// - spec.containers[*].securityContext.procMount
	// - spec.containers[*].securityContext.runAsUser
	// - spec.containers[*].securityContext.runAsGroup
	// +optional
	OS *corev1.PodOS `json:"os,omitempty" protobuf:"bytes,36,opt,name=os"`
	// Use the host's user namespace.
	// Optional: Default to true.
	// If set to true or not present, the pod will be run in the host user namespace, useful
	// for when the pod needs a feature only available to the host user namespace, such as
	// loading a kernel module with CAP_SYS_MODULE.
	// When set to false, a new userns is created for the pod. Setting false is useful for
	// mitigating container breakout vulnerabilities even allowing users to run their
	// containers as root without actually having root privileges on the host.
	// This field is alpha-level and is only honored by servers that enable the UserNamespacesSupport feature.
	// +k8s:conversion-gen=false
	// +optional
	HostUsers *bool `json:"hostUsers,omitempty" protobuf:"bytes,37,opt,name=hostUsers"`
}

// +kubebuilder:object:generate=true

type IngressNetworkingV1 struct {
	ObjectMeta ObjectMeta                `json:"metadata,omitempty"`
	Spec       *networkingv1.IngressSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:generate=true

type RouteOpenshiftV1 struct {
	ObjectMeta ObjectMeta            `json:"metadata,omitempty"`
	Spec       *RouteOpenShiftV1Spec `json:"spec,omitempty"`
}

// +kubebuilder:object:generate=true

type HTTPRouteV1 struct {
	ObjectMeta ObjectMeta            `json:"metadata,omitempty"`
	Spec       gwapiv1.HTTPRouteSpec `json:"spec,omitempty"`
}

// RouteTargetReference specifies the target that resolve into endpoints. Only the 'Service'
// kind is allowed. Use 'weight' field to emphasize one over others.
type RouteTargetReference struct {
	// The kind of target that the route is referring to. Currently, only 'Service' is allowed
	//
	// +kubebuilder:validation:Enum=Service;""
	// +kubebuilder:default=Service
	Kind string `json:"kind" protobuf:"bytes,1,opt,name=kind"`

	// name of the service/target that is being referred to. e.g. name of the service
	//
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name" protobuf:"bytes,2,opt,name=name"`

	// weight as an integer between 0 and 256, default 100, that specifies the target's relative weight
	// against other target reference objects. 0 suppresses requests to this backend.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=256
	// +kubebuilder:default=100
	Weight *int32 `json:"weight" protobuf:"varint,3,opt,name=weight"`
}

// RoutePort defines a port mapping from a router to an endpoint in the service endpoints.
type RoutePort struct {
	// The target port on pods selected by the service this route points to.
	// If this is a string, it will be looked up as a named port in the target
	// endpoints port list. Required
	TargetPort intstr.IntOrString `json:"targetPort" protobuf:"bytes,1,opt,name=targetPort"`
}

// TLSTerminationType dictates where the secure communication will stop
// TODO: Reconsider this type in v2
type TLSTerminationType string

// InsecureEdgeTerminationPolicyType dictates the behavior of insecure
// connections to an edge-terminated route.
type InsecureEdgeTerminationPolicyType string

const (
	// TLSTerminationEdge terminate encryption at the edge router.
	TLSTerminationEdge TLSTerminationType = "edge"
	// TLSTerminationPassthrough terminate encryption at the destination, the destination is responsible for decrypting traffic
	TLSTerminationPassthrough TLSTerminationType = "passthrough"
	// TLSTerminationReencrypt terminate encryption at the edge router and re-encrypt it with a new certificate supplied by the destination
	TLSTerminationReencrypt TLSTerminationType = "reencrypt"

	// InsecureEdgeTerminationPolicyNone disables insecure connections for an edge-terminated route.
	InsecureEdgeTerminationPolicyNone InsecureEdgeTerminationPolicyType = "None"
	// InsecureEdgeTerminationPolicyAllow allows insecure connections for an edge-terminated route.
	InsecureEdgeTerminationPolicyAllow InsecureEdgeTerminationPolicyType = "Allow"
	// InsecureEdgeTerminationPolicyRedirect redirects insecure connections for an edge-terminated route.
	// As an example, for routers that support HTTP and HTTPS, the
	// insecure HTTP connections will be redirected to use HTTPS.
	InsecureEdgeTerminationPolicyRedirect InsecureEdgeTerminationPolicyType = "Redirect"
)

// LocalObjectReference contains enough information to let you locate the
// referenced object inside the same namespace.
// +structType=atomic
type LocalObjectReference struct {
	// name of the referent.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	// +optional
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
}

// TLSConfig defines config used to secure a route and provide termination
//
// +kubebuilder:validation:XValidation:rule="has(self.termination) && has(self.insecureEdgeTerminationPolicy) ? !((self.termination=='passthrough') && (self.insecureEdgeTerminationPolicy=='Allow')) : true", message="cannot have both spec.tls.termination: passthrough and spec.tls.insecureEdgeTerminationPolicy: Allow"
// +openshift:validation:FeatureGateAwareXValidation:featureGate=RouteExternalCertificate,rule="!(has(self.certificate) && has(self.externalCertificate))", message="cannot have both spec.tls.certificate and spec.tls.externalCertificate"
type OpenshiftTLSConfig struct {
	// termination indicates termination type.
	//
	// * edge - TLS termination is done by the router and http is used to communicate with the backend (default)
	// * passthrough - Traffic is sent straight to the destination without the router providing TLS termination
	// * reencrypt - TLS termination is done by the router and https is used to communicate with the backend
	//
	// Note: passthrough termination is incompatible with httpHeader actions
	// +kubebuilder:validation:Enum=edge;reencrypt;passthrough
	Termination TLSTerminationType `json:"termination" protobuf:"bytes,1,opt,name=termination,casttype=TLSTerminationType"`

	// certificate provides certificate contents. This should be a single serving certificate, not a certificate
	// chain. Do not include a CA certificate.
	Certificate string `json:"certificate,omitempty" protobuf:"bytes,2,opt,name=certificate"`

	// key provides key file contents
	Key string `json:"key,omitempty" protobuf:"bytes,3,opt,name=key"`

	// caCertificate provides the cert authority certificate contents
	CACertificate string `json:"caCertificate,omitempty" protobuf:"bytes,4,opt,name=caCertificate"`

	// destinationCACertificate provides the contents of the ca certificate of the final destination.  When using reencrypt
	// termination this file should be provided in order to have routers use it for health checks on the secure connection.
	// If this field is not specified, the router may provide its own destination CA and perform hostname validation using
	// the short service name (service.namespace.svc), which allows infrastructure generated certificates to automatically
	// verify.
	DestinationCACertificate string `json:"destinationCACertificate,omitempty" protobuf:"bytes,5,opt,name=destinationCACertificate"`

	// insecureEdgeTerminationPolicy indicates the desired behavior for insecure connections to a route. While
	// each router may make its own decisions on which ports to expose, this is normally port 80.
	//
	// If a route does not specify insecureEdgeTerminationPolicy, then the default behavior is "None".
	//
	// * Allow - traffic is sent to the server on the insecure port (edge/reencrypt terminations only).
	//
	// * None - no traffic is allowed on the insecure port (default).
	//
	// * Redirect - clients are redirected to the secure port.
	//
	// +kubebuilder:validation:Enum=Allow;None;Redirect;""
	InsecureEdgeTerminationPolicy InsecureEdgeTerminationPolicyType `json:"insecureEdgeTerminationPolicy,omitempty" protobuf:"bytes,6,opt,name=insecureEdgeTerminationPolicy,casttype=InsecureEdgeTerminationPolicyType"`

	// externalCertificate provides certificate contents as a secret reference.
	// This should be a single serving certificate, not a certificate
	// chain. Do not include a CA certificate. The secret referenced should
	// be present in the same namespace as that of the Route.
	// Forbidden when `certificate` is set.
	// The router service account needs to be granted with read-only access to this secret,
	// please refer to openshift docs for additional details.
	//
	// +openshift:enable:FeatureGate=RouteExternalCertificate
	// +optional
	ExternalCertificate *LocalObjectReference `json:"externalCertificate,omitempty" protobuf:"bytes,7,opt,name=externalCertificate"`
}

// WildcardPolicyType indicates the type of wildcard support needed by routes.
type WildcardPolicyType string

const (
	// WildcardPolicyNone indicates no wildcard support is needed.
	WildcardPolicyNone WildcardPolicyType = "None"

	// WildcardPolicySubdomain indicates the host needs wildcard support for the subdomain.
	// Example: For host = "www.acme.test", indicates that the router
	//          should support requests for *.acme.test
	//          Note that this will not match acme.test only *.acme.test
	WildcardPolicySubdomain WildcardPolicyType = "Subdomain"
)

type RouteOpenShiftV1Spec struct {
	Host string `json:"host,omitempty" protobuf:"bytes,1,opt,name=host"`
	Path string `json:"path,omitempty" protobuf:"bytes,2,opt,name=path"`

	To *RouteTargetReference `json:"to,omitempty" protobuf:"bytes,3,opt,name=to"`

	AlternateBackends []RouteTargetReference `json:"alternateBackends,omitempty" protobuf:"bytes,4,rep,name=alternateBackends"`

	Port *RoutePort `json:"port,omitempty" protobuf:"bytes,5,opt,name=port"`

	TLS *OpenshiftTLSConfig `json:"tls,omitempty" protobuf:"bytes,6,opt,name=tls"`

	// +optional
	Subdomain string `json:"subdomain,omitempty" protobuf:"bytes,8,opt,name=subdomain"`

	WildcardPolicy WildcardPolicyType `json:"wildcardPolicy,omitempty" protobuf:"bytes,7,opt,name=wildcardPolicy"`
}

type ServiceV1 struct {
	ObjectMeta ObjectMeta          `json:"metadata,omitempty"`
	Spec       *corev1.ServiceSpec `json:"spec,omitempty"`
}

type PersistentVolumeClaimV1 struct {
	ObjectMeta ObjectMeta                   `json:"metadata,omitempty"`
	Spec       *PersistentVolumeClaimV1Spec `json:"spec,omitempty"`
}

type PersistentVolumeClaimV1Spec struct {
	// +optional
	AccessModes []corev1.PersistentVolumeAccessMode `json:"accessModes,omitempty" protobuf:"bytes,1,rep,name=accessModes,casttype=PersistentVolumeAccessMode"`
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty" protobuf:"bytes,4,opt,name=selector"`
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,2,opt,name=resources"`
	// VolumeName is the binding reference to the PersistentVolume backing this claim.
	// +optional
	VolumeName string `json:"volumeName,omitempty" protobuf:"bytes,3,opt,name=volumeName"`
	// +optional
	StorageClassName *string `json:"storageClassName,omitempty" protobuf:"bytes,5,opt,name=storageClassName"`
	// +optional
	VolumeMode *corev1.PersistentVolumeMode `json:"volumeMode,omitempty" protobuf:"bytes,6,opt,name=volumeMode,casttype=PersistentVolumeMode"`
	// +optional
	DataSource *corev1.TypedLocalObjectReference `json:"dataSource,omitempty" protobuf:"bytes,7,opt,name=dataSource"`
	// +optional
	DataSourceRef *corev1.TypedLocalObjectReference `json:"dataSourceRef,omitempty" protobuf:"bytes,8,opt,name=dataSourceRef"`
}

type ServiceAccountV1 struct {
	ObjectMeta ObjectMeta               `json:"metadata,omitempty"`
	Secrets    []corev1.ObjectReference `json:"secrets,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,2,rep,name=secrets"`
	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty" protobuf:"bytes,3,rep,name=imagePullSecrets"`
	// +optional
	AutomountServiceAccountToken *bool `json:"automountServiceAccountToken,omitempty" protobuf:"varint,4,opt,name=automountServiceAccountToken"`
}

// Merge merges `overrides` into `base` using the SMP (structural merge patch) approach.
// - It intentionally does not remove fields present in base but missing from overrides
// - It merges slices only if the `patchStrategy:"merge"` tag is present and the `patchMergeKey` identifies the unique field
func Merge(base, overrides any) error {
	if overrides == nil {
		return nil
	}

	baseBytes, err := json.Marshal(base)
	if err != nil {
		return fmt.Errorf("failed to convert current object to byte sequence: %w", err)
	}

	overrideBytes, err := json.Marshal(overrides)
	if err != nil {
		return fmt.Errorf("failed to convert current object to byte sequence: %w", err)
	}

	patchMeta, err := strategicpatch.NewPatchMetaFromStruct(base)
	if err != nil {
		return fmt.Errorf("failed to produce patch meta from struct: %w", err)
	}

	patch, err := strategicpatch.CreateThreeWayMergePatch(overrideBytes, overrideBytes, baseBytes, patchMeta, true)
	if err != nil {
		return fmt.Errorf("failed to create three way merge patch: %w", err)
	}

	merged, err := strategicpatch.StrategicMergePatchUsingLookupPatchMeta(baseBytes, patch, patchMeta)
	if err != nil {
		return fmt.Errorf("failed to apply patch: %w", err)
	}

	valueOfBase := reflect.Indirect(reflect.ValueOf(base))

	into := reflect.New(valueOfBase.Type())
	if err := json.Unmarshal(merged, into.Interface()); err != nil {
		return err
	}

	if !valueOfBase.CanSet() {
		return fmt.Errorf("unable to set unmarshalled value into base object")
	}

	valueOfBase.Set(reflect.Indirect(into))

	return nil
}
