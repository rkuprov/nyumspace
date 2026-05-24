# nyumspace

## Local Development

Prerequisites: [minikube](https://minikube.sigs.k8s.io/docs/start/), [helm](https://helm.sh/docs/intro/install/), [kubectl](https://kubernetes.io/docs/tasks/tools/).

### Starting the stack

```bash
minikube start
make mk-build   # build the image into minikube's docker daemon
make helm-up    # deploy postgres + app, start tunnel on localhost:8080
```

`helm-up` starts `minikube tunnel` in the background (requires sudo for routing) and deploys the Helm chart. The app pod waits for postgres to be ready before starting.

Verify everything is up:

```bash
kubectl get pods -n nyumspace
curl localhost:8080/hello
```

### Shutting down

```bash
make helm-down      # uninstall the Helm release and stop the tunnel
minikube stop       # pause the cluster (state is preserved)
# or
minikube delete     # wipe the cluster entirely
```

### Redeploying after a code change

When you change the app and compile a new binary, rebuild the image and do a rolling update:

```bash
make mk-build                                         # rebuild image in minikube
kubectl rollout restart deployment/nyumspace -n nyumspace  # replace running pod
kubectl rollout status deployment/nyumspace -n nyumspace   # wait until ready
```

The Helm release doesn't need to be reinstalled — `rollout restart` pulls the updated `nyumspace:local` image and swaps the pod in place. Postgres and the tunnel are unaffected.
