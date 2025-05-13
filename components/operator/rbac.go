package operator

// +kubebuilder:rbac:groups=operator.kyma-project.io,resources=registryproxies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.kyma-project.io,resources=registryproxies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=operator.kyma-project.io,resources=registryproxies/finalizers,verbs=update

// +kubebuilder:rbac:groups="",resources=secrets,verbs=list;get;watch;create;update;patch;delete

// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=list;get;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=list;get;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apiextensions.k8s.io",resources=customresourcedefinitions,verbs=list;get;watch;create;update;patch;delete

// +kubebuilder:rbac:groups="apps",resources=daemonsets,verbs=list;get;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apps",resources=deployments,verbs=list;get;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apps",resources=replicasets,verbs=list
// +kubebuilder:rbac:groups="apps",resources=deployments/status,verbs=list;get;watch;create;update;patch;delete

// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterrolebindings;clusterroles;rolebindings;roles,verbs=list;get;watch;create;update;patch;delete;bind;escalate

// +kubebuilder:rbac:groups="scheduling.k8s.io",resources=priorityclasses,verbs=list;get;watch;create;update;patch;delete
