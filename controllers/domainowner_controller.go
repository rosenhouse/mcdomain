/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/labels"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"

	discoveryv1beta1 "k8s.io/api/discovery/v1beta1"
	clusterv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/imikushin/controllers-af/function"
	"github.com/imikushin/controllers-af/reconciler"

	funv1alpha1 "github.com/rosenhouse/mcdomain/api/v1alpha1"
)

// DomainOwnerReconciler reconciles a DomainOwner object
type DomainOwnerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *DomainOwnerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&funv1alpha1.DomainOwner{}).
		Watches(&source.Kind{Type: &clusterv1alpha3.Cluster{}}, reconciler.EnqueueRequestsForQuery(r.Client, r.Log, clustersInSameNS)).
		Complete(reconciler.New(r.Client, r.Log, &funv1alpha1.DomainOwner{}, ReconcileFun))
}

func clustersInSameNS(obj client.Object) function.Query {
	return function.Query{
		Namespace: obj.GetNamespace(),
		Type:      &funv1alpha1.DomainOwnerList{},
	}
}

// +kubebuilder:rbac:groups=fun.xcc,resources=domainowners,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=fun.xcc,resources=domainowners/status,verbs=get;update;patch

func ReconcileFun(_ context.Context, object client.Object, getDetails function.GetDetails) (*function.Effects, error) {
	domainOwner := object.(*funv1alpha1.DomainOwner)
	log.Printf("reconcile %+v\n", domainOwner)

	clusterSelector, err := labelSelector(domainOwner.Spec.Owners)
	if err != nil {
		return nil, errors.Wrap(err, "getting cluster selector")
	}
	log.Printf("got cluster selector %+v\n", clusterSelector)

	backendClusters := getDetails(function.Query{
		Namespace: domainOwner.Namespace,
		Type:      &clusterv1alpha3.ClusterList{},
		Selector:  clusterSelector,
	}).(*clusterv1alpha3.ClusterList)
	log.Printf("got backend clusters %+v\n", backendClusters)

	endpointSliceRef := getDetails(function.Query{
		Type:      &discoveryv1beta1.EndpointSlice{},
		Namespace: domainOwner.Namespace,
		Name:      domainOwner.Name,
	})
	var endpointSlice *discoveryv1beta1.EndpointSlice = nil
	if endpointSliceRef != nil {
		endpointSlice = endpointSliceRef.(*discoveryv1beta1.EndpointSlice)
	}

	if len(backendClusters.Items) == 0 {
		log.Println("no clusters found")
		if endpointSliceRef == nil {
			log.Println("no existing endpointslice, nothing to do")
			return &function.Effects{}, nil
		}
		return &function.Effects{
			Deletes: []client.Object{endpointSliceRef.(*discoveryv1beta1.EndpointSlice)},
		}, nil
	}

	clusterFQDNs := make([]string, len(backendClusters.Items))
	for i, c := range backendClusters.Items {
		clusterFQDNs[i] = gatewayFQDNForCluster(c, DefaultSuffix)
	}
	sort.Strings(clusterFQDNs)
	log.Printf("got cluster FQDNs %+v\n", clusterFQDNs)

	if endpointSlice == nil {
		endpointSlice = &discoveryv1beta1.EndpointSlice{
			ObjectMeta: metav1.ObjectMeta{
				Name:      domainOwner.Name,
				Namespace: domainOwner.Namespace,
				Labels:    map[string]string{"fun.xcc/scatter": "true"},
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: domainOwner.APIVersion,
						Kind:       domainOwner.Kind,
						Name:       domainOwner.Name,
						UID:        domainOwner.UID,
					},
				},
			},
		}
	}
	endpointSlice.AddressType = discoveryv1beta1.AddressTypeFQDN
	endpointSlice.Endpoints = []discoveryv1beta1.Endpoint{
		{
			Addresses: clusterFQDNs,
		},
	}
	log.Printf("formed endpointslice: %+v\n", endpointSlice)

	return &function.Effects{
		Persists: []client.Object{endpointSlice},
	}, nil
}

const DefaultSuffix = "xcc.test"

func gatewayFQDNForCluster(c clusterv1alpha3.Cluster, suffix string) string {
	return fmt.Sprintf("x.gateway.%s.%s.clusters.%s", c.Name, c.Namespace, suffix)
}

func labelSelector(mv1Selector *metav1.LabelSelector) (labels.Selector, error) {
	if mv1Selector == nil {
		return labels.Everything(), nil
	}
	return metav1.LabelSelectorAsSelector(mv1Selector)
}
