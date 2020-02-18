#!/bin/bash

# https://kubernetes.io/docs/concepts/cluster-administration/certificates/#openssl

MASTER_IP=immutable-checker.default.svc

openssl genrsa -out ca.key 4096
openssl req -x509 -new -nodes -key ca.key -subj "/CN=${MASTER_IP}" -days 10000 -out ca.crt
openssl genrsa -out server.key 4096

cat << _EOF_ > ./csr.conf
[ req ]
default_bits = 2048
prompt = no
default_md = sha256
req_extensions = req_ext
distinguished_name = dn

[ dn ]
C = JP
ST = dummy
L = dummy
O = dummy
OU = dummy
CN = $MASTER_IP

[ req_ext ]
subjectAltName = @alt_names

[ alt_names ]
DNS.1 = $MASTER_IP

[ v3_ext ]
authorityKeyIdentifier=keyid,issuer:always
basicConstraints=CA:FALSE
keyUsage=keyEncipherment,dataEncipherment
extendedKeyUsage=serverAuth,clientAuth
subjectAltName=@alt_names
_EOF_

openssl req -new -key server.key -out server.csr -config csr.conf
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 10000 -extensions v3_ext -extfile csr.conf

openssl x509  -noout -text -in ./server.crt


