#!/bin/bash

ETCD_IMAGE="quay.io/coreos/etcd:v3.5.17"
GOBGP_IMAGE="jauderho/gobgp:v3.32.0"
ETCD_CONTAINER_NAME="etcd"
GOBGP_CONTAINER_NAME="gobgp"
CERTS_PATH="./certs"

# Generate the new CA certificate
openssl genrsa -out ./certs/ca.key 2048
openssl req -x509 -new -nodes -key ./certs/ca.key -sha256 -days 1024 -out ./certs/ca.crt -subj "/CN=ca"

# Generate new ETCD certificates
openssl genrsa -out ./certs/server.key 2048
openssl req -new -key ./certs/server.key -out ./certs/server.csr -config ./certs/openssl.cnf
openssl x509 -req -in ./certs/server.csr -CA ./certs/ca.crt -CAkey ./certs/ca.key -CAcreateserial -out ./certs/server.crt -days 365 -sha256 -extensions v3_ca -extfile ./certs/openssl.cnf
openssl genrsa -out ./certs/client.key 2048
openssl req -new -key ./certs/client.key -out ./certs/client.csr -subj "/CN=client"
openssl x509 -req -in ./certs/client.csr -CA ./certs/ca.crt -CAkey ./certs/ca.key -CAcreateserial -out ./certs/client.crt -days 365 -sha256

chmod -R 644 ${CERTS_PATH}/*

if docker ps -a | grep -q ${ETCD_CONTAINER_NAME}; then
    docker stop ${ETCD_CONTAINER_NAME} && docker rm ${ETCD_CONTAINER_NAME}
fi
if docker ps -a | grep -q ${GOBGP_CONTAINER_NAME}; then
    docker stop ${GOBGP_CONTAINER_NAME} && docker rm ${GOBGP_CONTAINER_NAME}
fi

docker run -d \
    --name ${ETCD_CONTAINER_NAME} \
    --network host \
    --rm \
    -v ${CERTS_PATH}:/etc/ssl/certs \
    ${ETCD_IMAGE} \
        /usr/local/bin/etcd \
        --data-dir=/etcd-data \
        --advertise-client-urls=https://localhost:2379 \
        --listen-client-urls=https://0.0.0.0:2379 \
        --cert-file=/etc/ssl/certs/server.crt \
        --key-file=/etc/ssl/certs/server.key \
        --client-cert-auth \
        --trusted-ca-file=/etc/ssl/certs/ca.crt

docker run -d \
    --name ${GOBGP_CONTAINER_NAME} \
    --network host \
    --rm \
    -v ${CERTS_PATH}:/etc/gobgp \
    ${GOBGP_IMAGE} \
        gobgpd -f /etc/gobgp/gobgp.toml \
        --tls true \
        --tls-cert-file /etc/gobgp/server.crt \
        --tls-key-file /etc/gobgp/server.key \
        --tls-client-ca-file /etc/ca.crt
