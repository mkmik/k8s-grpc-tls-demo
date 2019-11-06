# Overview

This is small demo showing got to setup gRPC behind an ingress controller (possibly encrypted by TLS).

# Run

```
./stage.sh build diff apply -- --context k8s-dev -n myns --image-prefix mkmik/tlsdemo
```

# Load balancing

## Services

In Kubernetes, a Service is an abstraction which defines a logical set of Pods and a policy by which to access them.

The most common way to access your pods is to let k8s allocate a single IP address and offer a L3 (TCP) load balancer
that then forwards the packets to one of your pods.

However, since the load balancing is at the TCP level, if you only make one connection, you'll only hit one of the pods.

The gRPC protocol relies heavily on persistent HTTP/2 connections and multiplexes many requests through mutiple channels over a few low-level TCP connections. This means that using a standard "ClusterIP" K8s service won't give you a good load balancing strategy. Indeed, you can see that yourself from the logs a `client-clusterip` pod.

```
$ kubectl logs --tail=5 $(kubectl get pod -lapp=client-clusterip -oname | head -n1)
2019/11/08 14:24:35 Greeting: "demo-5754bf6f97-dgngm" says: Hello world
2019/11/08 14:24:36 Greeting: "demo-5754bf6f97-dgngm" says: Hello world
2019/11/08 14:24:37 Greeting: "demo-5754bf6f97-dgngm" says: Hello world
2019/11/08 14:24:38 Greeting: "demo-5754bf6f97-dgngm" says: Hello world
2019/11/08 14:24:39 Greeting: "demo-5754bf6f97-dgngm" says: Hello world
```

## Headless Services

The simplest way around that is to use a [Headless Service](https://kubernetes.io/docs/concepts/services-networking/service/#headless-services):

> For headless Services that define selectors, the endpoints controller creates Endpoints records in the API, and modifies the DNS configuration to return records (addresses) that point directly to the Pods backing the Service.

This means that your gRPC client will now resolve multiple IP addresses and the gRPC library (at least the ones for popular languages like Go) should be smart enough to open multiple TCP connections to each of those backends.

However, there are downsides (some of them covered by the [official docs](https://kubernetes.io/docs/concepts/services-networking/service/#why-not-use-round-robin-dns)):

1. not all gRPC clients might correctly resolve all the A records.
2. some clients might not correctly handle DNS record caching and low TTLs.

To prove that there are issues, while preparing this README I saw the headless service demo properly load balance once, but now that I'm capturing those docs I tried again and noticed that it's not working:

```
$ kubectl logs --tail=5 $(kubectl get pod -lapp=client-headless -oname | head -n1)
2019/11/08 15:16:46 Greeting: "demo-dbbfffc4c-2fddj" says: Hello world
2019/11/08 15:16:47 Greeting: "demo-dbbfffc4c-2fddj" says: Hello world
2019/11/08 15:16:48 Greeting: "demo-dbbfffc4c-2fddj" says: Hello world
2019/11/08 15:16:49 Greeting: "demo-dbbfffc4c-2fddj" says: Hello world
2019/11/08 15:16:50 Greeting: "demo-dbbfffc4c-2fddj" says: Hello world
```

## Ingress

An alternative is to use a gRPC aware Layer 7 load balancer, such as Nginx or Envoy.
I tried
First, let's prove it works:

```
$ kubectl logs --tail=5 $(kubectl get pod -lapp=client-ingress -oname | head -n1)
2019/11/08 15:24:09 Greeting: "demo-dbbfffc4c-cvfzb" says: Hello world
2019/11/08 15:24:10 Greeting: "demo-dbbfffc4c-4tp58" says: Hello world
2019/11/08 15:24:11 Greeting: "demo-dbbfffc4c-2fddj" says: Hello world
2019/11/08 15:24:12 Greeting: "demo-dbbfffc4c-cvfzb" says: Hello world
2019/11/08 15:24:13 Greeting: "demo-dbbfffc4c-4tp58" says: Hello world
```

One gotcha: the k8s [nginx ingress controller](https://github.com/kubernetes/ingress-nginx) does support gRPC, but you must enable TLS in the ingress, otherwise it cannot serve gRPC traffic.
