# HostPort Allocator
A Kubernetes Operator to allocate host ports

## Overview

Kubernetes host ports are useful for exposing pods directly on the node they are running on. However the overhead
of managing allocation and limiting usage on a multi-tenant cluster can be difficult.

HostPort Allocator aims to solve this problem by treating host ports similar to persistent volumes. To use a host port
a `HostPortClaim` and `HostPort` must be created, the `HostPort` will then be allocated a port from a `HostPortClass`. 
Once allocated a `Pod` can then be created referencing the `HostPortClaim` via an annotation. The HostPort Allocator
will then allow the pod to use the allocated host port and will automatically modify the pod to have a host port defined
in its ports list.

## TODO

- [ ] Quota to restrict the number of `HostPortClaims` in a namespace
- [ ] Qutoa to restrict the number of `HostPortClaims` using a certain `HostPortClass` in a namespace
- [ ] Allow StatefulSets to use a `HostPortClaimTemplate` if unique host ports per pod are required

## Prerequisites

* Kubernetes `>=1.16.0`
* [Cert Manager](https://github.com/jetstack/cert-manager) with CAInjector `>=v0.15.2` (Optional)
  * If automatic certificate generation is desired for admission webhooks.

## Custom Resource Definitions

* **`HostPortClass`**, which defines a desired class and its pools of ports.

* **`HostPortClaim`**, which defines a desired claim for a host port.

* **`HostPort`**, which defines a desired allocation for a host port.

## Dynamic Admission Control

### Custom Resources

To prevent invalid resources from being created or modified an admission webhook is provided.

### Pods

To prevent pods from being created with invalid host ports an admission webhook is provided. This webhook only acts on
pods created in namespaces with certain labels, the default label selector is `hostport.rmb938.com: "true"`.

## Quickstart

### Install Cert Manager

```shell script
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v0.15.2/cert-manager.yaml
kubectl wait -n cert-manager --for=condition=Available --timeout=300s deployment cert-manager-webhook
```

### Install HostPort Allocator

```shell script
kubectl apply -f https://github.com/rmb938/hostport-allocator/releases/download/v0.1.3/hostport-allocator.yaml
kubectl wait -n hostport-allocator --for=condition=Available --timeout=300s deployment hostport-allocator
```

### Usage

1. Label the default namespace
    ```shell script
    kubectl label namespace default "hostport.rmb938.com='true'"
    ```
1. Create a `HostPortClass`
    ```yaml
    hostportclass.yaml
   
    ---
    apiVersion: hostport.rmb938.com/v1alpha1
    kind: HostPortClass
    metadata:
      name: sample
    spec:
      pools:
        - start: 9000
          end: 9500
    ```
    ```shell script
    kubectl apply -f hostportclass.yaml
    ```
1. Create a `HostPortClaim`
    ```yaml
    hostportclaim.yaml
   
    ---
    apiVersion: hostport.rmb938.com/v1alpha1
    kind: HostPortClaim
    metadata:
      name: echo-web
      namespace: default
    spec:
      hostPortClassName: sample
    ```
    ```shell script
    kubectl apply -f hostportclaim.yaml
    ```
1. Create a `Pod` using the `HostPortClaim`
    ```yaml
    pod.yaml
   
    ---
    apiVersion: v1
    kind: Pod
    metadata:
      name: echo
      namespace: default
      annotations:
        claim.hostport.rmb938.com/web: echo-web
    spec:
      containers:
        - name: echo
          image: k8s.gcr.io/echoserver:1.4
          ports:
            - name: web
              containerPort: 8080
          env:
            - name: MY_HOST_PORT
              valueFrom:
                fieldRef:
                  fieldPath: metadata.annotations['port.hostport.rmb938.com/web']
    ```
    ```shell script
    kubectl apply -f pod.yaml
    ```
1. The `Pod` will now be allocated the `HostPort` created by the `HostPortClaim` and will have an
environment variable of `MY_HOST_PORT` set to the port that was allocated.
   
## Development

### Prerequisites

* Golang `>=1.13`
* Docker
* Kind
* Tilt

### Running Locally

### Testing

#### Unit Tests

#### End-to-End Tests
