# Hello Operator
This repo creates a `hello world` example operator.

## Steps to recreate this project
1. Create the hello-operator project and move into the repo.
```bash
$ operator-sdk new hello-operator --api-version=github.awgreene.com/v1alpha1 --kind=Hello \
    && cd hello-operator
```

2. Build and push the app-operator image to a public registry such as quay.io.
```bash
$ export REPO=<YOUR REPO> \
    && export VERSION=0.0.1 \
    && operator-sdk build $REPO/hello-operator:$VERSION \
    && docker push $REPO/hello-operator:$VERSION
```

3. Update the operator manifest to use the built image name.
```bash
$ sed -i "s|REPLACE_IMAGE|${REPO}/hello-operator:${VERSION}|g" deploy/operator.yaml
```

4. Deploy the hello-operator
```bash
$ oc create -f deploy/sa.yaml \
    && oc create -f deploy/crd.yaml \
    && oc create -f deploy/rbac.yaml \
    && oc create -f deploy/operator.yaml
```

5. Creating a custom resource (App) triggers the hello-operator to deploy a busybox pod.
```bash
$ oc create -f deploy/cr.yaml
```

