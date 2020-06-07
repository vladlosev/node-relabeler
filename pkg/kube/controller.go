package kube

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	informers_core_v1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/vladlosev/node-relabeler/pkg/spec"
)

// Controller is the class with the relabeling logic.
type Controller struct {
	client          kubernetes.Interface
	informerFactory informers.SharedInformerFactory
	nodeInformer    informers_core_v1.NodeInformer
	specs           []spec.Spec
}

// NewController constructs new instance of Controller.
func NewController(client kubernetes.Interface, specs []spec.Spec) (*Controller, error) {
	informerFactory := informers.NewSharedInformerFactory(client, time.Hour*24)
	controller := &Controller{
		client:          client,
		informerFactory: informerFactory,
		nodeInformer:    informerFactory.Core().V1().Nodes(),
		specs:           specs,
	}
	controller.nodeInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    controller.addNode,
			UpdateFunc: controller.updateNode,
		},
	)
	return controller, nil
}

// Run runs the controller until the stop channel is signalled.
func (c *Controller) Run(stopCh <-chan struct{}) error {
	c.informerFactory.Start(stopCh)
	if !cache.WaitForCacheSync(stopCh, c.nodeInformer.Informer().HasSynced) {
		return fmt.Errorf("Failed to sync node informer cache")
	}
	<-stopCh
	return nil
}

func (c *Controller) addNode(obj interface{}) {
	c.updateNode(nil, obj)
}

func (c *Controller) updateNode(oldObj interface{}, newObj interface{}) {
	node, ok := newObj.(*core_v1.Node)
	if !ok {
		logrus.WithField("obj", newObj).Error("Unexpected object received (not a Node)")
	}
	logrus.WithField("name", node.Name).Info("Received node update")

	replacements := map[string]string{}

	for key, value := range node.Labels {
		for _, spec := range c.specs {
			if spec.OldKey.MatchString(key) && spec.OldValue.MatchString(value) {
				var newKey, newValue string
				if spec.OldKey.NumSubexp() > 0 {
					newKey = spec.OldKey.ReplaceAllString(key, spec.NewKey)
					newValue = spec.OldKey.ReplaceAllString(value, spec.NewValue)
				} else if spec.OldValue.NumSubexp() > 0 {
					newKey = spec.OldValue.ReplaceAllString(key, spec.NewKey)
					newValue = spec.OldValue.ReplaceAllString(value, spec.NewValue)
				} else {
					newKey = spec.NewKey
					newValue = spec.NewValue
				}
				replacements[newKey] = newValue
			}
		}
	}
	updated := false
	for key, value := range replacements {
		if oldValue, ok := node.Labels[key]; !ok || value != oldValue {
			fields := logrus.Fields{"node": node.Name, "key": key, "newValue": value}
			if ok {
				fields["oldValue"] = oldValue
			}
			logrus.WithFields(fields).Debug("Updated node label")
			node.Labels[key] = value
			updated = true
		}
	}
	if updated {
		logrus.WithField("node", node.Name).Info("Updating node")
		_, err := c.client.CoreV1().Nodes().Update(
			context.TODO(),
			node,
			meta_v1.UpdateOptions{})
		if err != nil {
			logrus.WithField("node", node.Name).WithError(err).Error(
				"Failed to update node")
		}
	}
}
