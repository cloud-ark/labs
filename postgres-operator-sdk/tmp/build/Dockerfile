FROM alpine:3.6

RUN adduser -D postgres-operator-sdk
USER postgres-operator-sdk

ADD tmp/_output/bin/postgres-operator-sdk /usr/local/bin/postgres-operator-sdk
