#!/bin/sh

hello-operator-test -test.parallel=1 -root=/ -kubeconfig=incluster -namespacedMan=namespaced.yaml -test.v
