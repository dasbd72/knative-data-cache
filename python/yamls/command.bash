#!/bin/bash

# sample value for your variables
MANAGER_URL=$(kubectl get ksvc manager --template '{{.status.url}}')

# read the yml template from a file and substitute the string 
template=`cat "steps.yaml.template" | sed "s/tmp_manager_url/$MANAGER_URL/g"`

# apply the yml with the substituted value
echo "$template" | kubectl apply -f -
