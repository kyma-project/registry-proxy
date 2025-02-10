package controller

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get

// TODO: check if needed
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups="",resources=pods/status,verbs=get

//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete;deletecollection
