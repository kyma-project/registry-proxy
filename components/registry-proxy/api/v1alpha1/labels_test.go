package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/labels"
)

// TestManagedBySelector locks in the informer-cache scoping that fixes the OOM in
// issue #139: the selector must admit the resources this controller stamps and
// reject unrelated cluster objects, so the cache never holds arbitrary Pods.
func TestManagedBySelector(t *testing.T) {
	selector := ManagedBySelector()

	testCases := []struct {
		name   string
		labels labels.Set
		want   bool
	}{
		{
			name:   "matches a resource managed by this controller",
			labels: labels.Set{LabelManagedBy: ManagedByValue},
			want:   true,
		},
		{
			name: "matches when extra labels are present",
			labels: labels.Set{
				LabelManagedBy: ManagedByValue,
				LabelApp:       "some-connection",
				LabelResource:  "deployment",
			},
			want: true,
		},
		{
			name:   "rejects a resource managed by something else",
			labels: labels.Set{LabelManagedBy: "someone-else"},
			want:   false,
		},
		{
			name:   "rejects an unrelated object with no managed-by label",
			labels: labels.Set{"app": "unrelated"},
			want:   false,
		},
		{
			name:   "rejects an object with no labels at all",
			labels: labels.Set{},
			want:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, selector.Matches(tc.labels))
		})
	}
}

// TestManagedByValueMatchesStampedLabel guards against the selector value drifting
// away from the value resources.labels() stamps on managed objects. The cache is
// scoped by LabelManagedBy=ManagedByValue while handle_pod_status lists Pods by
// LabelApp; the two selectors AND together, so if ManagedByValue stopped matching
// the value the Pods carry, the controller would silently see no Pods at all.
func TestManagedByValueMatchesStampedLabel(t *testing.T) {
	require.Equal(t, "registry-proxy", ManagedByValue,
		"resources.labels() stamps this literal on managed objects; keep them in sync")
}
