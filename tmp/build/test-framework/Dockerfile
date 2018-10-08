ARG BASEIMAGE

FROM ${BASEIMAGE}

ADD tmp/_output/bin/hello-operator-test /usr/local/bin/hello-operator-test
ARG NAMESPACEDMAN
ADD $NAMESPACEDMAN /namespaced.yaml
ADD tmp/build/go-test.sh /go-test.sh
