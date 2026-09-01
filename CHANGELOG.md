# Changelog

## [1.3.0](https://github.com/NItishSh/realestate-trust/compare/v1.2.0...v1.3.0) (2026-09-01)


### Features

* **architecture:** complete Phase 3 gaps (GAP-08 fail-closed KMS, GAP-10 reloader rotation, compose-test-e2e) ([#16](https://github.com/NItishSh/realestate-trust/issues/16)) ([6019e8f](https://github.com/NItishSh/realestate-trust/commit/6019e8f23be175ca019a0ab7cc0290940efab577))


### Bug Fixes

* **ci:** make compose E2E test in CI deterministic and fix postgres healthcheck ([#18](https://github.com/NItishSh/realestate-trust/issues/18)) ([8d5a067](https://github.com/NItishSh/realestate-trust/commit/8d5a067afbf417635fc5ed5eb7de55641417a3ba))

## [1.2.0](https://github.com/NItishSh/realestate-trust/compare/v1.1.2...v1.2.0) (2026-09-01)


### Features

* **architecture:** remediate critical data consistency and distributed concurrency gaps (Phase 1 & 2) ([#13](https://github.com/NItishSh/realestate-trust/issues/13)) ([2a43a1b](https://github.com/NItishSh/realestate-trust/commit/2a43a1bf164ee21f78e3b13f0c3c7cc0ed8e0f50))

## [1.1.2](https://github.com/NItishSh/realestate-trust/compare/v1.1.1...v1.1.2) (2026-09-01)


### Bug Fixes

* **infra:** add retry loop for ESO webhook warm-up and add production porting guide ([4ccae12](https://github.com/NItishSh/realestate-trust/commit/4ccae12beaad993e90a6cb7b1ff0013cbd064afb))

## [1.1.1](https://github.com/NItishSh/realestate-trust/compare/v1.1.0...v1.1.1) (2026-09-01)


### Bug Fixes

* **gitops:** update targetRevision from feature branch to main ([#9](https://github.com/NItishSh/realestate-trust/issues/9)) ([377087e](https://github.com/NItishSh/realestate-trust/commit/377087e04a19ec4a35dbee5020c5d0a61cedfe1a))

## [1.1.0](https://github.com/NItishSh/realestate-trust/compare/v1.0.1...v1.1.0) (2026-09-01)


### Features

* **infra:** terraform kind provisioning, keycloak IAM gitops integration, and dynamic smoke test auth ([bdfc35b](https://github.com/NItishSh/realestate-trust/commit/bdfc35b012d6a699ea87cee764886ba2722e7b46))

## [1.0.1](https://github.com/NItishSh/realestate-trust/compare/v1.0.0...v1.0.1) (2026-08-31)


### Bug Fixes

* update external-secrets apiVersion to v1beta1 ([#3](https://github.com/NItishSh/realestate-trust/issues/3)) ([8de4daf](https://github.com/NItishSh/realestate-trust/commit/8de4dafd2d27670fce4e2f1b72215beb6b0cf3ec))

## 1.0.0 (2026-08-25)


### Features

* add ArgoCD definitions for core infra (Istio, Vault, ESO) ([10ec887](https://github.com/NItishSh/realestate-trust/commit/10ec88739ea4ada0b4b1b53c8aa9c5d60b2ec657))
* add comprehensive system manual and update Dockerfile to Go 1.26.5 ([08b28be](https://github.com/NItishSh/realestate-trust/commit/08b28beaf8d99f004e87debcd6dedfe66d295170))
* **backend:** add seed data for non-production environments ([aeac8ab](https://github.com/NItishSh/realestate-trust/commit/aeac8ab2ca0b059e1f013f78a899bd8554f2230f))
* bootstrap next.js frontend application with tailwind v4, zustand, and dockercompose mapping ([04549b0](https://github.com/NItishSh/realestate-trust/commit/04549b0fbcf59656f6ce756e7803ea3c13439548))
* bootstrap real estate trust and escrow platform monorepo ([4de286c](https://github.com/NItishSh/realestate-trust/commit/4de286c07ee9cae35f4cafe1ed091fdc647682e8))
* extend property schema, add repository update methods, and implement robust startup health check polling for services. ([fcf7aab](https://github.com/NItishSh/realestate-trust/commit/fcf7aabd95662260ec38958f97a5f63fa4ee358f))
* **frontend:** add interactive guided user journeys widget and documentation spec ([3213285](https://github.com/NItishSh/realestate-trust/commit/3213285831daca8368d50678e410089be9a1a58e))
* **frontend:** change all currency displays and labels to INR with rupees symbol ([61ef818](https://github.com/NItishSh/realestate-trust/commit/61ef818462b086b64d711bbdbb61a3f199668fe2))
* implement AES-256-GCM encryption utility and document security compliance roadmap ([c28d06b](https://github.com/NItishSh/realestate-trust/commit/c28d06b44abdeaed94df71d8f6a8924fa957331f))
* implement asynchronous ledger updates using RabbitMQ event messaging and update frontend to reflect state changes ([9bae287](https://github.com/NItishSh/realestate-trust/commit/9bae2879b8046daa36a89eae04a13f6f25432413))
* implement audit logging in user registration and rebrand UI references to Portal ([514450c](https://github.com/NItishSh/realestate-trust/commit/514450c9de3f60ac3380b0e2fabec519d452c884))
* implement comprehensive security hardening including JWT authentication, infrastructure secrets management, and container security context controls. ([2a57974](https://github.com/NItishSh/realestate-trust/commit/2a57974ed7d46c7ed3c9b0dbf7629ef634dfe6c7))
* implement distributed correlation IDs across services, messaging, and frontend for improved observability. ([3c5f197](https://github.com/NItishSh/realestate-trust/commit/3c5f1973e3e85f3a32c729760f84b38006848191))
* implement feedback service with API handlers, repository, database migrations, and frontend widget ([d4b5a6f](https://github.com/NItishSh/realestate-trust/commit/d4b5a6fc241bf16f7750796e89a14a8ce6cca990))
* implement Istio observability stack with unified gateway routing and disable HPA in microservice chart ([b8c097d](https://github.com/NItishSh/realestate-trust/commit/b8c097df330e902d82a46cc9bf18174e293ff9a6))
* implement property-registry-service, add marketplace UI, and enable escrow funding workflow ([1d2d4fe](https://github.com/NItishSh/realestate-trust/commit/1d2d4fea9d84ece226a9a7f2cf453ceb3b6fc17b))
* implement RabbitMQ reliability with publisher confirms, manual ACKs, and DLQs, and add idempotency support to the ledger repository via event ID tracking. ([601b37f](https://github.com/NItishSh/realestate-trust/commit/601b37fc82ac0be4bf5b6e6b0207274a222dbc72))
* implement refresh token flow with session management, short-lived JWTs, and idle logout timer ([51a1cc4](https://github.com/NItishSh/realestate-trust/commit/51a1cc47681042d054e27442949c09ff97c35d05))
* implement SQL-based persistence repositories, migrate database schemas, and update infrastructure configurations for all core services. ([6bece76](https://github.com/NItishSh/realestate-trust/commit/6bece761ccca96b9bd893ca38147500e9f1ca6f2))
* implement user authentication system with login/signup pages, password hashing, and session management ([00b2a7b](https://github.com/NItishSh/realestate-trust/commit/00b2a7bc081c5e46ef243ef6c840008ae792628a))
* **infra:** add kind cluster configuration ([5e645ea](https://github.com/NItishSh/realestate-trust/commit/5e645eac69be685686fc01649621faa546743bab))
* **infra:** add local Helm mode to kind setup ([0ec8b99](https://github.com/NItishSh/realestate-trust/commit/0ec8b9971a5e9b3e72b3a25fe59f9a4716f1c483))
* infrastructure and testing overhaul ([c7a1de8](https://github.com/NItishSh/realestate-trust/commit/c7a1de858b7f88733ed9a832c2b746b0505ceaac))
* infrastructure, observability, KEDA, and Stitch UI redesign updates ([6b26078](https://github.com/NItishSh/realestate-trust/commit/6b260789a1693d3a7969c1012aaf7bb266b72be4))
* initialize local git repository and perform initial commit ([cb341cf](https://github.com/NItishSh/realestate-trust/commit/cb341cfdee3cd8019cd08307945553c58e27f586))
* inject JWT authentication into infrastructure smoke test verification scripts ([fa737b0](https://github.com/NItishSh/realestate-trust/commit/fa737b0f89555c754afc1942de09c28e896ad85b))
* integrate HashiCorp Vault and External Secrets Operator (ESO) for dynamic secret management ([5954579](https://github.com/NItishSh/realestate-trust/commit/5954579672411eb25361fd7e3aeeca0f627bc3db))
* integrate HashiCorp Vault Database secrets engine for dynamic database credential management and replace static secrets with ExternalSecrets. ([0c67ab0](https://github.com/NItishSh/realestate-trust/commit/0c67ab00dfdc1f68fb3637f33751d47cf34b0359))
* integrate Istio ingress gateway, migrate services to ClusterIP, and update API client for uniform routing ([00e49a2](https://github.com/NItishSh/realestate-trust/commit/00e49a2562c39f6eac40a8cfd723647ba851b910))
* **services:** implement list and getAll endpoints for all Go services to satisfy frontend fetches ([463363c](https://github.com/NItishSh/realestate-trust/commit/463363cf71863bb8046b6008c62078c25e92849b))
* support service-specific critical path smoke testing via helm tpl script injections ([7e028ee](https://github.com/NItishSh/realestate-trust/commit/7e028ee66a3603a08cf34d3680c2b796fca70987))


### Bug Fixes

* **frontend:** handle possibly undefined actionTx during update ([1ba760c](https://github.com/NItishSh/realestate-trust/commit/1ba760c7a0d6f00065a91d91b5f6b514a7607c20))
* **frontend:** upgrade node base image to v20 in dockerfile to satisfy next.js ([907f35f](https://github.com/NItishSh/realestate-trust/commit/907f35f62a706f2e1a00d5c6a20f3a7df717b417))
* **gitops:** correct createNamespace syntax in root-application.yaml ([1e41ab1](https://github.com/NItishSh/realestate-trust/commit/1e41ab12a53cbf4bd6138912ae1ee92f4b3ba971))
* **helm:** explicitly set service ports to match target ports ([d344d84](https://github.com/NItishSh/realestate-trust/commit/d344d84422b514ed1aeef2bda4ce256f050211d9))
* **helm:** expose services via NodePort and fix frontend probes ([f7e04fe](https://github.com/NItishSh/realestate-trust/commit/f7e04feef238d5f1fc678375e383e8214b1e6eaa))
* **helm:** resolve missing env vars and deadlocked migration hook ([396e717](https://github.com/NItishSh/realestate-trust/commit/396e7172fb789a43353c0f63e18989470e8595c8))
* **infra:** use existing ArgoCD and GitOps configuration for Kind ([66cb5f2](https://github.com/NItishSh/realestate-trust/commit/66cb5f29c09b9a702f9c022ba992175aec75c993))
* **infra:** use server-side apply for ArgoCD installation ([ef1b223](https://github.com/NItishSh/realestate-trust/commit/ef1b223c3477c09236ec2e421984eba8478414dd))
* resolve CRD ordering issue by delegating raw manifests to ArgoCD, and fix reset.sh args ([64e284a](https://github.com/NItishSh/realestate-trust/commit/64e284ac17356357ace1e210d115f0d63f4d6903))
* resolve remaining errcheck and govet issues ([#1](https://github.com/NItishSh/realestate-trust/issues/1)) ([c4055df](https://github.com/NItishSh/realestate-trust/commit/c4055df1f0d5cd11a55512568412f6af265c5c40))
* **services:** add CORS middleware to permit browser connections from port 3000 ([0430183](https://github.com/NItishSh/realestate-trust/commit/04301839f6496e284b109105da34caf8de0f6331))
