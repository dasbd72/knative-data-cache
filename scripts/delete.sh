#!/bin/bash
kubectl delete -f yamls/apps.yaml

kubectl delete -f yamls/manager.yaml
kubectl delete -f yamls/clusterrolebinding.yaml
kubectl delete -f yamls/clusterrole.yaml

kubectl delete -f yamls/pvc.yaml
kubectl delete -f yamls/pv.yaml
