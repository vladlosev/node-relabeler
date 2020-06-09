package kube

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vladlosev/node-relabeler/pkg/specs"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	go_testing "k8s.io/client-go/testing"
)

func TestControllerLabelUpdate(t *testing.T) {
	testData := []struct {
		name      string
		specs     []string
		labels    map[string]string
		validator func(t *testing.T, node *core_v1.Node)
	}{
		{
			"AddsNewLabel",
			[]string{"abc=def:uvw=xyz"},
			map[string]string{"abc": "def"},
			func(t *testing.T, node *core_v1.Node) {
				require.Contains(t, node.Labels, "uvw")
				assert.Equal(t, node.Labels["uvw"], "xyz")
				assert.Len(t, node.Labels, 2)
			},
		},
		{
			"ReplaceExistingLabel",
			[]string{"abc=def:abc=xyz"},
			map[string]string{"abc": "def"},
			func(t *testing.T, node *core_v1.Node) {
				require.Contains(t, node.Labels, "abc")
				assert.Equal(t, node.Labels["abc"], "xyz")
				assert.Len(t, node.Labels, 1)
			},
		},
		{
			"AddsNewLabelWithWildcard",
			[]string{"abc=*:def*=x"},
			map[string]string{"abc": "123"},
			func(t *testing.T, node *core_v1.Node) {
				require.Contains(t, node.Labels, "def123")
				assert.Equal(t, node.Labels["def123"], "x")
				assert.Len(t, node.Labels, 2)
			},
		},
		{
			"ModifiesExistingLabelWithWildcard",
			[]string{"abc=*:abc=x*x"},
			map[string]string{"abc": "123"},
			func(t *testing.T, node *core_v1.Node) {
				require.Contains(t, node.Labels, "abc")
				assert.Equal(t, node.Labels["abc"], "x123x")
				assert.Len(t, node.Labels, 1)
			},
		},
		{
			"HandlesMultipleReplacements",
			[]string{
				"abc=*:def=*",
				"uvw=xyz:uvw=ABC",
			},
			map[string]string{
				"abc": "123",
				"uvw": "xyz",
			},
			func(t *testing.T, node *core_v1.Node) {
				require.Contains(t, node.Labels, "def")
				assert.Equal(t, node.Labels["def"], "123")
				require.Contains(t, node.Labels, "uvw")
				assert.Equal(t, node.Labels["uvw"], "ABC")
				assert.Len(t, node.Labels, 3)
			},
		},
	}
	for _, testItem := range testData {
		t.Run(testItem.name, func(t *testing.T) {
			specs, err := specs.Parse(testItem.specs)
			require.NoError(t, err)
			node := &core_v1.Node{ObjectMeta: meta_v1.ObjectMeta{
				Name:   "test-node",
				Labels: testItem.labels,
			}}
			fakeClient := fake.NewSimpleClientset(node)
			updateChan := make(chan struct{})
			fakeClient.PrependReactor(
				"update",
				"nodes",
				func(action go_testing.Action) (bool, runtime.Object, error) {
					// Make sure we don't close updateChan more than once when multiple
					// updates arrive.
					select {
					case <-updateChan:
						break
					default:
						close(updateChan)
					}
					return false, nil, nil
				},
			)
			controller, err := NewController(fakeClient, specs)
			require.NoError(t, err)
			stopChan := make(chan struct{})
			stopCacheSyncChan := make(chan struct{})
			doneChan := make(chan struct{})
			go func(stop <-chan struct{}, done chan<- struct{}) {
				// We use runInternal here to avoid interupting the cache sync.
				err = controller.runInternal(stopChan, stopCacheSyncChan)
				assert.NoError(t, err)
				close(done)
			}(stopChan, doneChan)
			select {
			case <-updateChan:
				close(stopChan)
				<-doneChan
				updated, err := fakeClient.CoreV1().Nodes().Get(
					context.TODO(),
					node.Name,
					meta_v1.GetOptions{},
				)
				require.NoError(t, err)
				testItem.validator(t, updated)
				break
			case <-time.After(1000 * time.Microsecond):
				assert.Fail(t, "No expected node updates received")
				close(stopChan)
				<-doneChan
			}
		})
	}
}
