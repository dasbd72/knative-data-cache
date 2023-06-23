#!/bin/bash

kubectl apply -f yamls/pv.yaml
kubectl apply -f yamls/pvc.yaml

kubectl apply -f yamls/clusterrole.yaml
kubectl apply -f yamls/clusterrolebinding.yaml
kubectl apply -f yamls/manager.yaml

kubectl apply -f yamls/apps.yaml
