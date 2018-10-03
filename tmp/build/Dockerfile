FROM alpine:3.6

RUN adduser -D hello-operator
USER hello-operator

ADD tmp/_output/bin/hello-operator /usr/local/bin/hello-operator
