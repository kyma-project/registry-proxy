package resources

import "github.tools.sap/kyma/image-pull-reverse-proxy/components/controller/api/v1alpha1"

func labels(rp *v1alpha1.ImagePullReverseProxy, resource string) map[string]string {
	return map[string]string{
		v1alpha1.LabelApp:        rp.Name,
		v1alpha1.LabelName:       rp.Name,
		v1alpha1.LabelManagedBy:  "image-pull-reverse-proxy",
		v1alpha1.LabelModuleName: "image-pull-reverse-proxy",
		v1alpha1.LabelResource:   resource,
		v1alpha1.LabelPartOf:     "image-pull-reverse-proxy",
	}

}
