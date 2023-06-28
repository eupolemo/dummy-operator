# dummy
This is a Kubernetes controller 

## Description
This dummy controller with a custom resource called dummy with a POD associated.

## Getting Started
You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Running on the cluster
1. Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=docker.io/eupolemo/dummy:latest
```

**NOTE** You need permission to push the container. For purpose of run without 

2. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=docker.io/eupolemo/dummy:latest
```

3. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/dummy_v1alpha1_dummy.yaml
```

### Delete Resources:

```sh
kubectl delete -f config/samples/dummy_v1alpha1_dummy.yaml
```

### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller
UnDeploy the controller from the cluster:

```sh
make undeploy
```

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/),
which provide a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster.

### Test It Out
1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

3. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/dummy_v1alpha1_dummy.yaml
```

**NOTE:** The logs will show the status changing and the Dummy.Spec info

4. Describe the Custom Resource will show the state of the resource:

```sh
kubectl describe dummies.dummy.interview.com dummy-sample
```

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

