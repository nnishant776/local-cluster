osType() {
    echo "$(echo $(uname) | tr '[[:upper:]]' '[[:lower:]]')"
}

cpuArch() {
    arch=$(uname -m)
    case $arch in
        "x86_64" )
            echo ${arch/x86_/amd}
            ;;
        * )
            echo $arch
            ;;
    esac
}

genCA() {
    openssl genrsa -out ca.key 2048
    openssl req -x509 -new -nodes -key ca.key -subj "/CN=${CLUSTER_HOSTNAME}" -days 365 -out ca.crt
}

genCSR() {
cat << EOF | envsubst | tee csr.conf
[ req ]
default_bits = 2048
prompt = no
default_md = sha256
req_extensions = req_ext
distinguished_name = dn
[ dn ]
C = UC
ST = Andromeda
L = Mirach
O = Constellation
OU = Development
CN = ${CLUSTER_HOSTNAME}
[ req_ext ]
subjectAltName = @alt_names
[ alt_names ]
DNS.1 = *.${CLUSTER_HOSTNAME}
[ v3_ext ]
authorityKeyIdentifier=keyid,issuer:always
basicConstraints=CA:FALSE
keyUsage=keyEncipherment,dataEncipherment
extendedKeyUsage=serverAuth,clientAuth
subjectAltName=@alt_names
EOF
}

genTLS() {
    openssl genrsa -out server.key 2048
    openssl req -new -key server.key -out server.csr -config csr.conf
    openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key \
        -CAcreateserial -out server.crt -days 10000 \
        -extensions v3_ext -extfile csr.conf
}

printTLSCert() {
    openssl x509 -noout -text -in ./server.crt
}
