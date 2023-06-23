# Connecting
kubectl port-forward --namespace logging service/log-collector 8080:80
ssh -L localhost:8080:localhost:8080 [HOST_IP]
# log at http://localhost:8080/