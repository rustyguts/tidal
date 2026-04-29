# Deploy Tidal to k3s

One-shot guide: build the image, push to GHCR, register Prefect deployments, run a flow.

## Prerequisites

- `docker` (or `podman`) with `buildx` for multi-arch builds
- `kubectl` configured with the `k3s` context (`kubectl config use-context k3s`)
- GHCR personal access token with `write:packages` scope
- Prefect CLI installed locally (`uv sync --extra dev`)

## Cluster facts

| Item | Value |
|------|-------|
| Context | `k3s` |
| Namespace | `prefect` |
| Work pool | `cirus` (Kubernetes type) |
| Worker | `deploy/prefect-worker` (already running) |
| Prefect API (in-cluster) | `http://prefect-server.prefect.svc.cluster.local:4200/api` |
| Registry | `ghcr.io/rustyguts/tidal` |
| Node arch | linux/amd64 |

The cluster already runs `prefect-server` and `prefect-worker` on the `cirus` work pool. You only need to publish a flow image and register deployments pointing at it.

## 1. Log in to GHCR

```bash
echo "$GHCR_TOKEN" | docker login ghcr.io -u <github-username> --password-stdin
```

## 2. Build and push the image

Use a git-sha tag for traceability plus `latest` for convenience.

```bash
TAG=$(git rev-parse --short HEAD)
IMAGE=ghcr.io/rustyguts/tidal

docker buildx build \
  --platform linux/amd64 \
  --tag "$IMAGE:$TAG" \
  --tag "$IMAGE:latest" \
  --push \
  .
```

First push? Make the package public (or add a pull secret — see Appendix A).

## 3. Port-forward Prefect API

The deploy script needs to talk to the Prefect API. Forward the in-cluster service to localhost:

```bash
kubectl --context k3s -n prefect port-forward svc/prefect-server 4200:4200
```

Leave that running in another terminal.

## 4. Register deployments

In a second shell:

```bash
export PREFECT_API_URL=http://localhost:4200/api
uv run deploy.py all --environment production
```

`deploy.py all` performs:
- `prefect_variables` — sets the `app_tidal` variable (env, data dir)
- `prefect_concurrency_limits` — creates task + global concurrency limits
- `prefect_flows` — registers the `pipeline/tidal-pipeline` and `transcode/simple-transcode` deployments

This stays running while it serves the deployments. Stop it with Ctrl+C once `Successfully deployed Prefect Flows!` prints — the deployments persist on the server.

## 5. Point deployments at your image

By default, deployments inherit the work-pool base job template. Override the image per deployment so Kubernetes jobs spawned by the worker pull your build:

```bash
prefect deployment set-job-variable transcode/simple-transcode \
  image="ghcr.io/rustyguts/tidal:$TAG" \
  image_pull_policy=Always

prefect deployment set-job-variable pipeline/tidal-pipeline \
  image="ghcr.io/rustyguts/tidal:$TAG" \
  image_pull_policy=Always
```

Verify:

```bash
prefect deployment inspect transcode/simple-transcode | grep -A3 job_variables
```

## 6. Trigger a run

```bash
prefect deployment run transcode/simple-transcode \
  --param config='{"input":"/data/in.mp4","output":"/data/out.mp4"}'
```

Watch the worker spawn a Kubernetes job:

```bash
kubectl --context k3s -n prefect get jobs -w
kubectl --context k3s -n prefect logs -l prefect.io/flow-run-id=<id> -f
```

Or follow worker logs:

```bash
kubectl --context k3s -n prefect logs deploy/prefect-worker -f
```

## 7. Promote a new build

```bash
TAG=$(git rev-parse --short HEAD)
docker buildx build --platform linux/amd64 -t ghcr.io/rustyguts/tidal:$TAG --push .

prefect deployment set-job-variable transcode/simple-transcode image="ghcr.io/rustyguts/tidal:$TAG"
prefect deployment set-job-variable pipeline/tidal-pipeline   image="ghcr.io/rustyguts/tidal:$TAG"
```

Re-run `uv run deploy.py all --environment production` whenever flow signatures, parameters, or concurrency limits change.

## Troubleshooting

| Symptom | Check |
|---------|-------|
| `ImagePullBackOff` on the spawned job | Package private — see Appendix A, or run `docker buildx imagetools inspect ghcr.io/rustyguts/tidal:$TAG` |
| Deployment stays in `Late` | Worker not polling: `kubectl -n prefect logs deploy/prefect-worker --tail=50` |
| `connection refused` from `deploy.py` | Port-forward dropped, restart step 3 |
| Wrong `PREFECT_API_URL` | Must end in `/api` (deploy.py asserts this) |
| Flow runs but can't read inputs | Mount needed PVC/hostPath via the work-pool base job template |

## Appendix A — Private GHCR pull secret

If you keep the package private:

```bash
kubectl --context k3s -n prefect create secret docker-registry ghcr-pull \
  --docker-server=ghcr.io \
  --docker-username=<github-username> \
  --docker-password=$GHCR_TOKEN
```

Then set `docker_credentials_secret_name=ghcr-pull` on the work pool's base job template (Prefect UI → Work Pools → cirus → Edit), or pass it per-deployment via `set-job-variable`.

## Appendix B — Quick verify script

```bash
kubectl --context k3s -n prefect get deploy,pod
kubectl --context k3s -n prefect exec deploy/prefect-server -- prefect deployment ls
kubectl --context k3s -n prefect exec deploy/prefect-server -- prefect work-pool inspect cirus
```
