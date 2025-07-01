package controller

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get

//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=pods/status,verbs=get

//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups="security.istio.io",resources=peerauthentications,verbs=get;list;watch;create;update;patch;delete;deletecollection

//+kubebuilder:rbac:groups="connectivityproxy.sap.com",resources=connectivityproxies,verbs=get;list;watch

// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch

// +kubebuilder:rbac:groups=registry-proxy.kyma-project.io,resources=connections,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=registry-proxy.kyma-project.io,resources=connections/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=registry-proxy.kyma-project.io,resources=connections/finalizers,verbs=update
