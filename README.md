# Overview

This is small demo showing got to setup gRPC behind an ingress controller (both legs encrypted by TLS).

# Run

```
./stage.sh build diff apply -- --context k8s-dev -n myns --image-prefix mkmik/tlsdemo
```
