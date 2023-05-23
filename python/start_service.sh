#!/bin/bash

kubectl apply -f yamls/pv.yaml
kubectl apply -f yamls/pvc.yaml

kubectl apply -f yamls/clusterrole.yaml
kubectl apply -f yamls/clusterrolebinding.yaml
kubectl apply -f yamls/manager.yaml

# sample value for your variables
MANAGER_URL="http://manager:8080"

# read the yml template from a file and substitute the string 
# {{MYVARNAME}} with the value of the MYVARVALUE variable
template=`cat "yamls/steps_template.yaml" | sed "s,tmp_manager_url,$MANAGER_URL,g"`

# apply the yml with the substituted value
echo "$template" | kubectl apply -f -
