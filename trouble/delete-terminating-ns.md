kubectl proxy


export TOKEN=$(kubectl describe secret $(kubectl get secrets | grep default | cut -f1 -d ' ') | grep -E '^token' | cut -f2 -d':' | tr -d '\t')
curl http://localhost:8001/api/v1/namespaces --header "Authorization: Bearer $TOKEN" --insecure

curl -X PUT --data-binary @logging.json http://localhost:8001/api/v1/namespaces/logging/finalize -H "Content-Type: application/json" --header "Authorization: Bearer $TOKEN" --insecure
