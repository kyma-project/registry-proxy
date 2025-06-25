package connectivityproxy

import (
	"github.tools.sap/kyma/registry-proxy/tests/common/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateMockStatefulSet(utils *utils.TestUtils) error {
	statefulSet := fixStatefulSet(utils)
	return utils.Client.Create(utils.Ctx, statefulSet)
}

func fixStatefulSet(testUtils *utils.TestUtils) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "connectivity-proxy",
			Namespace: testUtils.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "connectivity-proxy",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "connectivity-proxy",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "connectivity-proxy",
							Image: "alpine:latest",
							Command: []string{
								"sh",
								"-c",
								"echo 'Connectivity Proxy is running'; sleep infinity",
							},
						},
					},
				},
			},
		},
	}
}
