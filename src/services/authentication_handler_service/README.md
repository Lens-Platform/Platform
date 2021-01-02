# authentication_handler_service

[![e2e](https://github.com/blackspaceInc/BlackspacePlatform/src/services/authentication_handler_service/workflows/e2e/badge.svg)](https://github.com/blackspaceInc/BlackspacePlatform/src/services/authentication_handler_service/blob/master/.github/workflows/e2e.yml)
[![test](https://github.com/blackspaceInc/BlackspacePlatform/src/services/authentication_handler_service/workflows/test/badge.svg)](https://github.com/blackspaceInc/BlackspacePlatform/src/services/authentication_handler_service/blob/master/.github/workflows/test.yml)
[![cve-scan](https://github.com/blackspaceInc/BlackspacePlatform/src/services/authentication_handler_service/workflows/cve-scan/badge.svg)](https://github.com/blackspaceInc/BlackspacePlatform/src/services/authentication_handler_service/blob/master/.github/workflows/cve-scan.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/blackspaceInc/BlackspacePlatform/src/services/authentication_handler_service)](https://goreportcard.com/report/github.com/blackspaceInc/BlackspacePlatform/src/services/authentication_handler_service)
[![Docker Pulls](https://img.shields.io/docker/pulls/github.com/blackspaceInc/BlackspacePlatform/authentication_handler_service)](https://hub.docker.com/r/github.com/blackspaceInc/BlackspacePlatform/authentication_handler_service)

authentication_handler_service is a tiny web application made with Go that showcases best practices of running microservices in Kubernetes.

Specifications:

* Health checks (readiness and liveness)
* Graceful shutdown on interrupt signals
* File watcher for secrets and configmaps
* Instrumented with Prometheus
* Tracing with Istio and Jaeger
* Linkerd service profile
* Structured logging with zap
* 12-factor app with viper
* Fault injection (random errors and latency)
* Swagger docs
* Helm and Kustomize installers
* End-to-End testing with Kubernetes Kind and Helm
* Kustomize testing with GitHub Actions and Open Policy Agent
* Multi-arch container image with Docker buildx and Github Actions
* CVE scanning with trivy

Web API:

* `GET /` prints runtime information
* `GET /version` prints authentication_handler_service version and git commit hash
* `GET /metrics` return HTTP requests duration and Go runtime metrics
* `GET /healthz` used by Kubernetes liveness probe
* `GET /readyz` used by Kubernetes readiness probe
* `POST /readyz/enable` signals the Kubernetes LB that this instance is ready to receive traffic
* `POST /readyz/disable` signals the Kubernetes LB to stop sending requests to this instance
* `GET /status/{code}` returns the status code
* `GET /panic` crashes the process with exit code 255
* `POST /echo` forwards the call to the backend service and echos the posted content
* `GET /env` returns the environment variables as a JSON array
* `GET /headers` returns a JSON with the request HTTP headers
* `GET /delay/{seconds}` waits for the specified period
* `POST /token` issues a JWT token valid for one minute `JWT=$(curl -sd 'anon' authentication_handler_service:9898/token | jq -r .token)`
* `GET /token/validate` validates the JWT token `curl -H "Authorization: Bearer $JWT" authentication_handler_service:9898/token/validate`
* `GET /configs` returns a JSON with configmaps and/or secrets mounted in the `config` volume
* `POST/PUT /cache/{key}` saves the posted content to Redis
* `GET /cache/{key}` returns the content from Redis if the key exists
* `DELETE /cache/{key}` deletes the key from Redis if exists
* `POST /store` writes the posted content to disk at /data/hash and returns the SHA1 hash of the content
* `GET /store/{hash}` returns the content of the file /data/hash if exists
* `GET /ws/echo` echos content via websockets `podcli ws ws://localhost:9898/ws/echo`
* `GET /chunked/{seconds}` uses `transfer-encoding` type `chunked` to give a partial response and then waits for the specified period
* `GET /swagger.json` returns the API Swagger docs, used for Linkerd service profiling and Gloo routes discovery

gRPC API:

* `/grpc.health.v1.Health/Check` health checking

Web UI:

![authentication_handler_service-ui](https://raw.githubusercontent.com/github.com/blackspaceInc/BlackspacePlatform/authentication_handler_service/gh-pages/screens/authentication_handler_service-ui-v3.png)

To access the Swagger UI open `<authentication_handler_service-host>/swagger/index.html` in a browser.

### Guides

* [GitOps Progressive Deliver with Flagger, Helm v3 and Linkerd](https://helm.workshop.flagger.dev/intro/)
* [GitOps Progressive Deliver on EKS with Flagger and AppMesh](https://eks.handson.flagger.dev/prerequisites/)
* [Automated canary deployments with Flagger and Istio](https://medium.com/google-cloud/automated-canary-deployments-with-flagger-and-istio-ac747827f9d1)
* [Kubernetes autoscaling with Istio metrics](https://medium.com/google-cloud/kubernetes-autoscaling-with-istio-metrics-76442253a45a)
* [Autoscaling EKS on Fargate with custom metrics](https://aws.amazon.com/blogs/containers/autoscaling-eks-on-fargate-with-custom-metrics/)
* [Managing Helm releases the GitOps way](https://medium.com/google-cloud/managing-helm-releases-the-gitops-way-207a6ac6ff0e)
* [Securing EKS Ingress With Contour And Let’s Encrypt The GitOps Way](https://aws.amazon.com/blogs/containers/securing-eks-ingress-contour-lets-encrypt-gitops/)

### Install

Helm:

```bash
helm repo add authentication_handler_service https://github.com/blackspaceInc/BlackspacePlatform/authentication_handler_service

helm upgrade --install --wait frontend \
--namespace test \
--set replicaCount=2 \
--set backend=http://backend-authentication_handler_service:9898/echo \
authentication_handler_service/authentication_handler_service

# Test pods have hook-delete-policy: hook-succeeded
helm test frontend

helm upgrade --install --wait backend \
--namespace test \
--set hpa.enabled=true \
authentication_handler_service/authentication_handler_service
```

Kustomize:

```bash
kubectl apply -k github.com/blackspaceInc/BlackspacePlatform/src/services/authentication_handler_service//kustomize
```

Docker:

```bash
docker run -dp 9898:9898 github.com/blackspaceInc/BlackspacePlatform/authentication_handler_service
```

---
### TODO (This Week)
#### Technical
- [X] pull authn client library and enhance it with the custom impl. in the authentication folder
- [X] emit metrics for this service and fix errors
- [X] ensure some api are auth protected (use jwt)
- [ ] __ implement circuit breaker [link](https://github.com/cep21/circuit) in core library & wrap around all remote calls
- [ ] __ implement retryable operation in core library & for all remote operations [link](https://github.com/avast/retry-go)
- [X] implement distributed request tracing & wrap around all remote calls
- [ ] define graffana views, circuit breaker views, traces, ... etc to visualize metrics, circuit breaker, and traces and ensure
      template service has this implementation
- [ ] __ implement unit tests for all api's and new authn library client
- [ ] __ implement load testing for service and automate this flow
- [ ] __ configure github actions to run all service tests, run load tests, and end to end tests via docker end to end setup
- [ ] define minikube and docker-compose local test flows and test local deployments
- [ ] __ spec out user details and company details data for the registration flow
    - [X] model after square space account registration, square business account registration, shopify account registration, quora and pinterest
     account registration
- [ ] ___ implement graphql api gateway
- [ ] Implement linkerd side care deployment config (k8 and docker-compose) [link](https://github.com/LensPlatform/linkerd-examples)

#### Non Technical
- [ ] Reach out to 15 potential customers and follow the rubric specified in your product markdown file.
    - [ ] talk to them about your solution and ask what needs they may have and how you can better solve them
    - [ ] go on linkedIn find black owned businesses, get their emails and contact the entrepreneurs (either through email or through phone)
        - [ ] make note of the trends
- [ ] Watch 15 YCombinator Videos (1 Hour Long) & think
- [ ] Plan out sprint and allocate tasks


