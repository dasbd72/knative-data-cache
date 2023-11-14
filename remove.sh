#!/bin/sh
# apps
kubectl delete -f yamls/app-image-chain.yaml
kubectl delete -f yamls/app-video-chain.yaml

# controller
kubectl delete -f yamls/manager.yaml
kubectl delete -f yamls/data-serve.yaml
kubectl delete -f yamls/cache-deleter.yaml
