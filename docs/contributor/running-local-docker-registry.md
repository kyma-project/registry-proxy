# Set up Local Docker Registry for Testing

Set up a  Docker Registry running in the local Docker context for testing purposes.

## Prerequisites
- Docker installed and running (for the Basic authorization option)
- Htpasswd and openssl installed

1. Export the following environment variables:

   ```bash
   export REG_USER_NAME={REGISTRY_USERNAME}
   export REG_USER_PASSWD={REGISTRY_PASSWORD}
   export IMAGE_TAG={TAG}
   export IMAGE_NAME="on-prem-nginx"
   export DOCKER_REGISTRY_HOST="myregistry.acme"
   export DOCKER_REGISTRY_PORT="25002"
   export DOCKER_REGISTRY="${DOCKER_REGISTRY_HOST}:${DOCKER_REGISTRY_PORT}"
   export IMAGE_PATH="${DOCKER_REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}"
   ```


2. Generate a self-signed certificate:

   ```bash
   mkdir -p certs
   openssl req \
      -newkey rsa:4096 -nodes -sha256 -keyout certs/domain.key \
      -addext "subjectAltName = DNS:myregistry.acme" \
      -x509 -days 365 -out certs/domain.crt
   ```

3. Add the certificate to the system keychain:

   ```bash
   security add-trusted-cert -d -r trustRoot -k ~/Library/Keychains/login.keychain certs/domain.crt
   ```

4. Generate an authentication file for the local Docker registry:

   ```bash
   mkdir -p secret
   htpasswd -Bbn ${REG_USER_NAME} ${REG_USER_PASSWD} > ./secret/htpasswd
   ```

5. Add the required configuration file:

   ```bash
   mkdir -p config
   cat << EOF > config/config.yml
   version: 0.1
   log:
     fields:
       service: registry
   storage:
     cache:
       blobdescriptor: inmemory
     filesystem:
       rootdirectory: /var/lib/registry
   auth:
     htpasswd:
       realm: basic-realm
       path: /secret/htpasswd
   http:
     addr: 0.0.0.0:443
     tls:
       certificate: /certs/domain.crt
       key: /certs/domain.key
       minimumtls: tls1.2
       ciphersuites:
         - TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
         - TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
         - TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256
         - TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256
         - TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
         - TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
         - TLS_AES_128_GCM_SHA256
         - TLS_CHACHA20_POLY1305_SHA256
         - TLS_AES_256_GCM_SHA384
         - TLS_AES_128_GCM_SHA256
     headers:
       X-Content-Type-Options: [nosniff]
   EOF
   ```

6. Run the local Docker registry:

   ```bash
   docker run -d \
		   -v docker-reg-vol:/var/lib/registry \
		   -v $(pwd)/certs:/certs \
		   -v $(pwd)/config/config.yml:/etc/distribution/config.yml \
		   -v $(pwd)/secret/htpasswd:/secret/htpasswd \
		   -p 25002:443 \
		   --restart=always \
		   --name on-prem-docker-registry \
		   registry:3.0.0
   ```

7. Edit the `/etc/hosts` file and add the following line: `127.0.0.1 myregistry.acme`


8. Authenticate to the local Docker registry to push the test image:

   ```bash
   docker login ${DOCKER_REGISTRY} -u ${REG_USER_NAME} -p ${REG_USER_PASSWD}
   ```

9. Create and push a test image to the on-premise Docker Registry:

   ```bash
   echo -e "FROM nginx:alpine\nRUN echo \"<h1>Test image created on $(date +%F+%T)</h1>\" > /usr/share/nginx/html/index.html" | docker buildx build --push --platform linux/amd64 -t ${IMAGE_PATH} -
   ```
