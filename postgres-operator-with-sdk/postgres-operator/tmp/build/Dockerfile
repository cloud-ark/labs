FROM alpine:3.6

ADD tmp/_output/bin/postgres-operator /usr/local/bin/postgres-operator

RUN adduser -D postgres-operator
USER postgres-operator
