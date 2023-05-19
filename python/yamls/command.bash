#!/bin/bash

# sample value for your variables
MANAGER_URL=$(kubectl get pod image-scale --template '{{.status.podIP}}')
LOCAL_STORAGE = $(curl -X POST $MANAGER_URL:9090 -H 'Content-Type: application/json' -d '{"Query":"local_storage"}')

# read the yml template from a file and substitute the string 
# {{MYVARNAME}} with the value of the MYVARVALUE variable
template=`cat "steps.yaml.template" | sed "s/tmp_manager_url/$MANAGER_URL/g" | sed "s/tmp_local_storage/$LOCAL_STORAGE/g"`

# apply the yml with the substituted value
echo "$template" | kubectl apply -f -
