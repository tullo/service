# Deployment

## [Create a Kubernetes cluster](https://console.cloud.google.com/kubernetes/add?project=stackwise-starter-kit&authuser=2)

* Cluster basics
  * Name: `stackwise-starter-cluster`
  * Location type: Zonal
  * Zone: `europe-west3-c`
  * Master version
    * Release channel: `Regular-channel 1.16.xxx`
* Node Pools
  * default-pool
    * nodes
      * Machine type: `e2-small`
      * Boot disk type: `HDD`
* Security
  * Enable Shielded GKE Nodes: `check`
  * Enable Workload Identity: `check`
* Features
  * Enable Istio: `check`
  * Enable Application Manager: `check`

---

## [Cloud SQL](https://console.cloud.google.com)

* Open SQL [instances](https://console.cloud.google.com/sql/instances?authuser=2&project=stackwise-starter-kit) for project (stackwise-starter-kit)
    1. Create an instance
    1. Choose PostgreSQL
    1. Create a PostgreSQL instance
        * Instance ID: `stackwise-starter-db`
        * Default user password: `Generate` & `<COPY>`
        * Region: `europe-west3`
        * Zone: `europe-west3-c`
        * Database version: `PostgreSQL 12`
        * Show configuration options:
            * Private IP: network (default)
            * Public IP: unchecked
            * Availability: Single zone (not for production)
            * [Storage type](https://cloud.google.com/sql/docs/postgres/choosing-ssd-hdd?hl=tr): HDD (cheaper than SDD)
            * [Machine type](https://cloud.google.com/compute/docs/instances/creating-instance-with-custom-machine-type): keep the defaults (Cores:1 vCPU, Memory: 3,75GB)
        * **Create**
    1. Connect to this instance
        * **Private IP address**: `<COPY>` (referenced in deployment manifest)
* Resulting instance: [stackwise-starter-db](https://console.cloud.google.com/sql/instances/stackwise-starter-db/overview?authuser=2&folder=&organizationId=&project=stackwise-starter-kit&supportedpurview=project)

---

## [Connecting from Google Kubernetes Engine](https://cloud.google.com/sql/docs/postgres/connect-kubernetes-engine)

### Secrets

Generate [secret](makefile#L134) with database related data `make kctl-db-secret-create`

```sh
kubectl create secret generic stackwise \
  --from-literal=user=postgres \
  --from-literal=pass=<PASTE> \
  --from-literal=db=stackwise \
  --from-literal=db_host=<PASTE>
```

Configuration > Secret: [stackwise-starter-db](https://console.cloud.google.com/kubernetes/config?authuser=2&project=stackwise-starter-kit)

---

## [Connecting using the Cloud SQL Proxy](https://cloud.google.com/sql/docs/postgres/connect-kubernetes-engine#proxy)

The Cloud SQL Proxy is added to your [pod](https://kubernetes.io/docs/concepts/workloads/pods/) using the [sidecar container](https://cloud.google.com/sql/docs/postgres/connect-kubernetes-engine#running_the_proxy_as_a_sidecar) pattern. The proxy container is in the `same pod as your application`, which enables the application to connect to the proxy using `localhost`, increasing security and performance.

## [Connecting without the Cloud SQL proxy](https://cloud.google.com/sql/docs/postgres/connect-kubernetes-engine#private-ip)

```yaml
        env:
        - name: DB_USER
          valueFrom:
            secretKeyRef:
              name: sales-api
              key: user
        - name: DB_PASS
          valueFrom:
            secretKeyRef:
              name: sales-api
              key: pass
        - name: DB_NAME
          valueFrom:
            secretKeyRef:
              name: sales-api
              key: db
        - name: DB_HOST
          valueFrom:
            secretKeyRef:
              name: sales-api
              key: db_host
```

## [Connect to the cluster](https://console.cloud.google.com/kubernetes/clusters/details/europe-west3-c/stackwise-starter-cluster?project=stackwise-starter-kit&authuser=2)

```sh
gcloud container clusters get-credentials stackwise-starter-cluster --zone europe-west3-c --project stackwise-starter-kit
```

---

## [Port Forwarding](https://console.cloud.google.com/kubernetes/service/europe-west3-c/stackwise-starter-cluster/default/sales-api/overview?authuser=2&project=stackwise-starter-kit)

Port forwarding the `sales-api` port.

Listen on port 3000 locally, forwarding to 3000 in the pod.

```sh
gcloud container clusters get-credentials stackwise-starter-cluster --zone europe-west3-c --project stackwise-starter-kit \
 && kubectl port-forward $(kubectl get pod --selector="app=sales-api" --output jsonpath='{.items[0].metadata.name}') 3000:3000
```

---

## [Debug Running Pods](https://kubernetes.io/docs/tasks/debug-application-cluster/debug-running-pod/)

### Examining pod logs

```sh
kubectl logs ${POD_NAME} ${CONTAINER_NAME}

kubectl logs --previous ${POD_NAME} ${CONTAINER_NAME}
```

### Debugging with container exec

```sh
kubectl exec ${POD_NAME} -c ${CONTAINER_NAME} -- ${CMD} ${ARG1} ${ARG2} ... ${ARGN}

# run a shell
kubectl exec -it cassandra -- sh

# take a look at the logs
kubectl exec cassandra -- cat /var/log/cassandra/system.log
```

### Debugging with an `ephemeral debug container`

FEATURE STATE: Kubernetes `v1.18 (alpha)`

Ephemeral containers are useful for interactive troubleshooting when kubectl exec is insufficient because a container has crashed or a container image doesn't include debugging utilities, such as with [distroless images](https://github.com/GoogleContainerTools/distroless).

Requires the **EphemeralContainers** `feature gate` enabled in your **cluster and kubectl** version v1.18 or later.

```sh
# 1. Create a pod, which simulates a problem ...
kubectl run ephemeral-demo --image=k8s.gcr.io/pause:3.1 --restart=Never

# 2. Add a debugging container; kubectl attaches to the console of 'ephemeral-demo'
kubectl alpha debug -it ephemeral-demo --image=busybox --target=ephemeral-demo

# The --target parameter targets the process namespace of another container. 
```

["Distroless" Docker Images](https://github.com/GoogleContainerTools/distroless) contain **only your application and its runtime dependencies**. They do not contain package managers, shells or any other programs you would expect to find in a standard Linux distribution.

---

## [Kubernetes Tasks](https://kubernetes.io/docs/tasks/)

This section of the Kubernetes documentation contains pages that show how to do individual tasks. A task page shows how to do a single thing, typically by giving a short sequence of steps.
