#!/bin/sh
# pv
kubectl apply -f yamls/pv.yaml
kubectl apply -f yamls/pvc.yaml

# controller
kubectl apply -f yamls/manager.yaml
kubectl apply -f yamls/data-serve.yaml
kubectl apply -f yamls/cache-deleter.yaml

# apps
kubectl apply -f yamls/app-image-chain.yaml
kubectl apply -f yamls/app-video-chain.yaml
