/*
Copyright 2017 The Kubernetes Authors.

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

package source

import (
	"testing"

	"github.com/kubernetes-incubator/external-dns/endpoint"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

// Validates that ingressSource is a Source
var _ Source = &ingressSource{}

func TestIngress(t *testing.T) {
	t.Run("endpointsFromIngress", testEndpointsFromIngress)
	t.Run("Endpoints", testIngressEndpoints)
}

func testEndpointsFromIngress(t *testing.T) {
	for _, ti := range []struct {
		title    string
		ingress  fakeIngress
		expected []*endpoint.Endpoint
	}{
		{
			title: "one rule.host one lb.hostname",
			ingress: fakeIngress{
				dnsnames:  []string{"foo.bar"}, // Kubernetes requires removal of trailing dot
				hostnames: []string{"lb.com"},  // Kubernetes omits the trailing dot
			},
			expected: []*endpoint.Endpoint{
				{
					DNSName: "foo.bar",
					Target:  "lb.com",
				},
			},
		},
		{
			title: "one rule.host one lb.IP",
			ingress: fakeIngress{
				dnsnames: []string{"foo.bar"},
				ips:      []string{"8.8.8.8"},
			},
			expected: []*endpoint.Endpoint{
				{
					DNSName: "foo.bar",
					Target:  "8.8.8.8",
				},
			},
		},
		{
			title: "one rule.host two lb.IP and two lb.Hostname",
			ingress: fakeIngress{
				dnsnames:  []string{"foo.bar"},
				ips:       []string{"8.8.8.8", "127.0.0.1"},
				hostnames: []string{"elb.com", "alb.com"},
			},
			expected: []*endpoint.Endpoint{
				{
					DNSName: "foo.bar",
					Target:  "8.8.8.8",
				},
				{
					DNSName: "foo.bar",
					Target:  "127.0.0.1",
				},
				{
					DNSName: "foo.bar",
					Target:  "elb.com",
				},
				{
					DNSName: "foo.bar",
					Target:  "alb.com",
				},
			},
		},
		{
			title: "no rule.host",
			ingress: fakeIngress{
				ips:       []string{"8.8.8.8", "127.0.0.1"},
				hostnames: []string{"elb.com", "alb.com"},
			},
			expected: []*endpoint.Endpoint{},
		},
		{
			title: "one empty rule.host",
			ingress: fakeIngress{
				dnsnames:  []string{""},
				ips:       []string{"8.8.8.8", "127.0.0.1"},
				hostnames: []string{"elb.com", "alb.com"},
			},
			expected: []*endpoint.Endpoint{},
		},
		{
			title: "no targets",
			ingress: fakeIngress{
				dnsnames: []string{""},
			},
			expected: []*endpoint.Endpoint{},
		},
	} {
		t.Run(ti.title, func(t *testing.T) {
			realIngress := ti.ingress.Ingress()
			validateEndpoints(t, endpointsFromIngress(realIngress), ti.expected)
		})
	}
}

func testIngressEndpoints(t *testing.T) {
	namespace := "testing"
	for _, ti := range []struct {
		title           string
		targetNamespace string
		ingressItems    []fakeIngress
		expected        []*endpoint.Endpoint
	}{
		{
			title:           "no ingress",
			targetNamespace: "",
		},
		{
			title:           "two simple ingresses",
			targetNamespace: "",
			ingressItems: []fakeIngress{
				{
					name:      "fake1",
					namespace: namespace,
					dnsnames:  []string{"example.org"},
					ips:       []string{"8.8.8.8"},
				},
				{
					name:      "fake2",
					namespace: namespace,
					dnsnames:  []string{"new.org"},
					hostnames: []string{"lb.com"},
				},
			},
			expected: []*endpoint.Endpoint{
				{
					DNSName: "example.org",
					Target:  "8.8.8.8",
				},
				{
					DNSName: "new.org",
					Target:  "lb.com",
				},
			},
		},
		{
			title:           "two simple ingresses on different namespaces",
			targetNamespace: "",
			ingressItems: []fakeIngress{
				{
					name:      "fake1",
					namespace: "testing1",
					dnsnames:  []string{"example.org"},
					ips:       []string{"8.8.8.8"},
				},
				{
					name:      "fake2",
					namespace: "testing2",
					dnsnames:  []string{"new.org"},
					hostnames: []string{"lb.com"},
				},
			},
			expected: []*endpoint.Endpoint{
				{
					DNSName: "example.org",
					Target:  "8.8.8.8",
				},
				{
					DNSName: "new.org",
					Target:  "lb.com",
				},
			},
		},
		{
			title:           "two simple ingresses on different namespaces with target namespace",
			targetNamespace: "testing1",
			ingressItems: []fakeIngress{
				{
					name:      "fake1",
					namespace: "testing1",
					dnsnames:  []string{"example.org"},
					ips:       []string{"8.8.8.8"},
				},
				{
					name:      "fake2",
					namespace: "testing2",
					dnsnames:  []string{"new.org"},
					hostnames: []string{"lb.com"},
				},
			},
			expected: []*endpoint.Endpoint{
				{
					DNSName: "example.org",
					Target:  "8.8.8.8",
				},
			},
		},
		{
			title:           "our controller type is dns-controller",
			targetNamespace: "",
			ingressItems: []fakeIngress{
				{
					name:      "fake1",
					namespace: namespace,
					annotations: map[string]string{
						controllerAnnotationKey: controllerAnnotationValue,
					},
					dnsnames: []string{"example.org"},
					ips:      []string{"8.8.8.8"},
				},
			},
			expected: []*endpoint.Endpoint{
				{
					DNSName: "example.org",
					Target:  "8.8.8.8",
				},
			},
		},
		{
			title:           "different controller types are ignored",
			targetNamespace: "",
			ingressItems: []fakeIngress{
				{
					name:      "fake1",
					namespace: namespace,
					annotations: map[string]string{
						controllerAnnotationKey: "some-other-tool",
					},
					dnsnames: []string{"example.org"},
					ips:      []string{"8.8.8.8"},
				},
			},
			expected: []*endpoint.Endpoint{},
		},
	} {
		t.Run(ti.title, func(t *testing.T) {
			ingresses := make([]*v1beta1.Ingress, 0)
			for _, item := range ti.ingressItems {
				ingresses = append(ingresses, item.Ingress())
			}

			fakeClient := fake.NewSimpleClientset()
			ingressSource := NewIngressSource(fakeClient, ti.targetNamespace)
			for _, ingress := range ingresses {
				_, err := fakeClient.Extensions().Ingresses(ingress.Namespace).Create(ingress)
				if err != nil {
					t.Errorf("fake kubernetes ingress creation should not fail. Ingress %v. Error: %v", *ingress, err)
				}
			}

			res, err := ingressSource.Endpoints()
			if err != nil {
				t.Errorf("ingress endpoints should not fail on valid fake client call")
			}
			validateEndpoints(t, res, ti.expected)

		})
	}
}

// ingress specific helper functions
type fakeIngress struct {
	dnsnames    []string
	ips         []string
	hostnames   []string
	namespace   string
	name        string
	annotations map[string]string
}

func (ing fakeIngress) Ingress() *v1beta1.Ingress {
	ingress := &v1beta1.Ingress{
		ObjectMeta: v1.ObjectMeta{
			Namespace:   ing.namespace,
			Name:        ing.name,
			Annotations: ing.annotations,
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{},
		},
		Status: v1beta1.IngressStatus{
			LoadBalancer: v1.LoadBalancerStatus{
				Ingress: []v1.LoadBalancerIngress{},
			},
		},
	}
	for _, dnsname := range ing.dnsnames {
		ingress.Spec.Rules = append(ingress.Spec.Rules, v1beta1.IngressRule{
			Host: dnsname,
		})
	}
	for _, ip := range ing.ips {
		ingress.Status.LoadBalancer.Ingress = append(ingress.Status.LoadBalancer.Ingress, v1.LoadBalancerIngress{
			IP: ip,
		})
	}
	for _, hostname := range ing.hostnames {
		ingress.Status.LoadBalancer.Ingress = append(ingress.Status.LoadBalancer.Ingress, v1.LoadBalancerIngress{
			Hostname: hostname,
		})
	}
	return ingress
}
