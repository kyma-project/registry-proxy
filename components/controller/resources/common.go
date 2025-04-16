package resources

import "github.tools.sap/kyma/registry-proxy/components/controller/api/v1alpha1"

func labels(rp *v1alpha1.ImagePullReverseProxy, resource string) map[string]string {
	return map[string]string{
		v1alpha1.LabelApp:        rp.Name,
		v1alpha1.LabelName:       rp.Name,
		v1alpha1.LabelManagedBy:  "registry-proxy",
		v1alpha1.LabelModuleName: "registry-proxy",
		v1alpha1.LabelResource:   resource,
		v1alpha1.LabelPartOf:     "registry-proxy",
	}

}
