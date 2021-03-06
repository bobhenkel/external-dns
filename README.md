# ExternalDNS
[![Build Status](https://travis-ci.org/kubernetes-incubator/external-dns.svg?branch=master)](https://travis-ci.org/kubernetes-incubator/external-dns)
[![Coverage Status](https://coveralls.io/repos/github/kubernetes-incubator/external-dns/badge.svg?branch=master)](https://coveralls.io/github/kubernetes-incubator/external-dns?branch=master)
[![GitHub release](https://img.shields.io/github/release/kubernetes-incubator/external-dns.svg)](https://github.com/kubernetes-incubator/external-dns/releases)

ExternalDNS synchronizes exposed Kubernetes Services and Ingresses with DNS providers.

## What It Does

Inspired by [Kubernetes DNS](https://github.com/kubernetes/dns), Kubernetes' cluster-internal DNS server, ExternalDNS makes Kubernetes resources discoverable via public DNS servers. Like KubeDNS, it retrieves a list of resources (Services, Ingresses, etc.) from the [Kubernetes API](https://kubernetes.io/docs/api/) to determine a desired list of DNS records. *Unlike* KubeDNS, however, it's not a DNS server itself, but merely configures other DNS providers accordingly—e.g. [AWS Route 53](https://aws.amazon.com/route53/) or [Google CloudDNS](https://cloud.google.com/dns/docs/).

In a broader sense, ExternalDNS allows you to control DNS records dynamically via Kubernetes resources in a DNS provider-agnostic way.

The [FAQ](docs/faq.md) contains additional information and addresses several questions about key concepts of ExternalDNS.

## Getting started

ExternalDNS' current release is `v0.2`. This version allows you to keep a managed zone in Google's [CloudDNS](https://cloud.google.com/dns/docs/) or [AWS' Route 53](https://aws.amazon.com/route53/) synchronized with Ingresses and Services of `type=LoadBalancer` in your cluster.

In this release, ExternalDNS is limited to—and takes full ownership of—a single managed zone. In other words, if you have any existing records in that zone, they will be removed. We encourage you to try out ExternalDNS in its own zone first to see if that model works for you. However, ExternalDNS runs in dryRun mode by default, and won't make any changes to your infrastructure. So as long as you don't change that flag, you're safe.

### Technical Requirements

Make sure you have the following prerequisites:
* A local Go 1.7+ development environment.
* Access to a Google project with the DNS API enabled.
* Access to a Kubernetes cluster that supports exposing Services, e.g. GKE.
* A properly set up, **unused**, and **empty** hosted zone in Google CloudDNS.

### Setup Steps

First, get ExternalDNS:

```console
$ go get -u github.com/kubernetes-incubator/external-dns
```

Next, run an application and expose it via a Kubernetes Service:

```console
$ kubectl run nginx --image=nginx --replicas=1 --port=80
$ kubectl expose deployment nginx --port=80 --target-port=80 --type=LoadBalancer
```

Annotate the Service with your desired external DNS name. Make sure to change `example.org` to your domain.

```console
$ kubectl annotate service nginx "external-dns.alpha.kubernetes.io/hostname=nginx.example.org."
```

Locally run a single sync loop of ExternalDNS. Make sure to change the Google project to one you control, and the zone identifier to an **unused** and **empty** hosted zone in that project's Google CloudDNS:

```console
$ external-dns --zone example-org --provider google --google-project example-project --source service --once
```

This should output the DNS records it will modify to match the managed zone with the DNS records you desire.

Once you're satisfied with the result, you can run ExternalDNS like you would run it in your cluster: as a control loop, and not in dryRun mode:

```console
$ external-dns --zone example-org --provider google --google-project example-project --source service --dry-run=false
```

Check that ExternalDNS has created the desired DNS record for your Service and that it points to its load balancer's IP. Then try to resolve it:

```console
$ dig +short nginx.example.org.
104.155.60.49
```

Now you can experiment and watch how ExternalDNS makes sure that your DNS records are configured as desired. Here are a couple of things you can try out:
* Change the desired hostname by modifying the Service's annotation.
* Recreate the Service and see that the DNS record will be updated to point to the new load balancer IP.
* Add another Service to create more DNS records.
* Remove Services to clean up your managed zone.

The [tutorials](docs/tutorials) section contains examples, including Ingress resources, and shows you how to set up ExternalDNS in different environments such as other cloud providers and alternative Ingress controllers.

# Roadmap

ExternalDNS was built with extensibility in mind. Adding and experimenting with new DNS providers and sources of desired DNS records should be as easy as possible. It should also be possible to modify how ExternalDNS behaves—e.g. whether it should add records but never delete them.

We're working on an ownership system that allows ExternalDNS to never modify records over which it lacks control.

Here's a rough outline on what is to come:

### v0.1

* Support for Google CloudDNS
* Support for Kubernetes Services

### v0.2

* Support for AWS Route 53
* Support for Kubernetes Ingresses

### v0.3

* Support for AWS Route 53 via ALIAS
* Support for multiple zones
* Ownership System

### v1.0

* Ability to replace Kops' [DNS Controller](https://github.com/kubernetes/kops/tree/master/dns-controller)
* Ability to replace Zalando's [Mate](https://github.com/zalando-incubator/mate)
* Ability to replace Molecule Software's [route53-kubernetes](https://github.com/wearemolecule/route53-kubernetes)

### Yet to be defined

* Support for CoreDNS and Azure DNS
* Support for record weights
* Support for different behavioral policies
* Support for Services with `type=NodePort`
* Support for TPRs
* Support for more advanced DNS record configurations

Have a look at [the milestones](https://github.com/kubernetes-incubator/external-dns/milestones) to get an idea of where we currently stand.

## Contributing

We encourage you to get involved with ExternalDNS, as users as well as contributors. Read the [contributing guidelines](CONTRIBUTING.md) and have a look at [the contributing docs](docs/contributing/getting-started.md) to learn about building the project, the project structure, and the purpose of each package.

Feel free to reach out to us on the [Kubernetes slack](http://slack.k8s.io) in the #sig-network channel.

## Heritage

ExternalDNS is an effort to unify the following similar projects in order to bring the Kubernetes community an easy and predictable way of managing DNS records across cloud providers based on their Kubernetes resources:

* Kops' [DNS Controller](https://github.com/kubernetes/kops/tree/master/dns-controller)
* Zalando's [Mate](https://github.com/zalando-incubator/mate)
* Molecule Software's [route53-kubernetes](https://github.com/wearemolecule/route53-kubernetes)

## Kubernetes Incubator

This is a [Kubernetes Incubator project](https://github.com/kubernetes/community/blob/master/incubator.md).
The project was established 2017-Feb-9 (initial announcement [here](https://groups.google.com/forum/#!searchin/kubernetes-dev/external$20dns%7Csort:relevance/kubernetes-dev/2wGQUB0fUuE/9OXz01i2BgAJ)).
The incubator team for the project is:

* Sponsor: sig-network
* Champion: Tim Hockin (@thockin)
* SIG: sig-network

For more information about sig-network, such as meeting times and agenda, check out the [community site](https://github.com/kubernetes/community/tree/master/sig-network).

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).
