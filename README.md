# Hello Operator
This repo was created from my desire to better understand how to create an operator using the [Operator SDK](https://github.com/operator-framework/operator-sdk). The Hello Operator is able to scale the number of [hello-go image](https://github.com/awgreene/hello-go) containers to the value specified in the Hello Custom Resource (CR). The work shown here is heavily influenced by the [Memcached Operator Sample](https://github.com/operator-framework/operator-sdk-samples/tree/master/memcached-operator).

## Prerequisites
1. Have the [OpenShift CLI installed](https://www.okd.io/download.html)
1. Have an OpenShift cluster up and running.
2. Be logged into the OpenShift cluster as a user with the Cluster Admin role.
3. Have the [Operator SDK installed](https://github.com/operator-framework/operator-sdk#quick-start)

## Steps to recreate this project from scratch
1. Create the hello-operator project and move into it.
```bash
$ operator-sdk new hello-operator --api-version=github.awgreene.com/v1alpha1 --kind=Hello \
    && cd hello-operator
```

2. Modify the spec and status of the Hello CR at `pkg/apis/cache/v1alpha1/types.go`.
```Go
...
type HelloSpec struct {
  // Size is the size of the hello deployment
  Size  int32  `json:"size"`
  
  // World is an Environment variable passed into the containers
  World string `json:"world"`
}

type HelloStatus struct {
  // Nodes are the names of the hello pods
  Nodes []string `json:"nodes"`
}
```

3. Update the generated code for the CR:
```bash
$ operator-sdk generate k8s
```

4. Define where you will store your hello-operator image.
```bash
$ export REGISTRY=<SOME_REGISTRY> \
   && export NAMESPACE=<SOME_NAMESPACE> \
   && export REPOSITORY=<SOME_REPOSITORY> \
   && export TAG=<SOME_TAG>
```

5. Update the controller logic by changing the `pkg/stub/handler.go` file to match the same file in this repo. Make sure not to replace your hello-operator project import.
	
	NOTE: Notice that we watch for new values for each field defined in the `HelloSpec` strcuture found in `pkg/apis/cache/v1alpha1/types.go` as shown on  [line 51](https://github.com/awgreene/hello-operator/blob/master/pkg/stub/handler.go#L51).

    NOTE: Notice that the Deployment yaml is defined within the deploymentForHello function on [line 79](https://github.com/awgreene/hello-operator/blob/master/pkg/stub/handler.go#L79)

6. Build and push the hello-operator image to a public registry such as quay.io.
```bash
$ operator-sdk build $REGISTRY/$NAMESPACE/$REPOSITORY:$TAG \
   && docker push $REGISTRY/$NAMESPACE/$REPOSITORY:$TAG
```

7. Update the `deploy/operator.yaml` file to deploy the image you just pushed.
```bash
$ sed -i "s|REPLACE_IMAGE|${REGISTRY}/${NAMESPACE}/${REPOSITORY}:${TAG}|g" deploy/operator.yaml
```

7. Deploy the hello-operator.
```bash
$ oc apply -f deploy/sa.yaml \
    && oc apply -f deploy/crd.yaml \
    && oc apply -f deploy/rbac.yaml \
    && oc apply -f deploy/operator.yaml
```

8. View your hello-operator.
```bash
$ oc get pods
// Expected output
NAME                             READY     STATUS    RESTARTS   AGE
hello-operator-5bb798cd5-5rrjx   1/1       Running   0          9m
```

9. Modify `deploy/cr.yaml` as shown.
```yaml
apiVersion: "github.awgreene.com/v1alpha1"
kind: "Hello"
metadata:
  name: "hello-go"
spec:
  size: 2
  world: "Go Programmer"
```

10. Create a Hello CR. The hello-operator will deploy two hello-go pods in response.
```bash
$ oc apply -f deploy/cr.yaml
```

11. View your deployment. The number of example pods should match the `Size` field defined in the `deploy/cr.yaml` file.
```bash
$ oc get pods
// Expected output
NAME                             READY     STATUS    RESTARTS   AGE
hello-go-69d74f6f56-jx27j         1/1       Running   0          8m
hello-go-69d74f6f56-qq7d9         1/1       Running   0          8m
hello-operator-5bb798cd5-5rrjx   1/1       Running   0          9m
```

12. Run a cURL command from within a container to view your message. The recipient of the hello should match the `WORLD` field defined in `deploy/cr.yaml:
```bash
$ kubectl exec -it $(kubectl get pods -o go-template -l app=hello -o jsonpath='{.items[0].metadata.name}') curl localhost:8000/env
// Expected output
Hello, Go Programmer!
```

With that, you have successfully create the Hello Operator. Try changing the size of the deployment and the message!

## Testing with the Operator SDK
It's possible to use the Operator SDK for end-to-end testing. In this example, our tests will deploy two clusters in ephemeral namespaces that will scale four hello-go images.
1. Copy the `tests` directory from this repo to your project.

2. Make sure all dependencies are in your project.
```bash
$ dep ensure
```

3. Run the tests using the operator sdk.
```bash
$ operator-sdk test --test-location ./tests/e2e
```