#!/bin/bash -ex

docker build . -f build/Dockerfile -t voyager

kubectl apply -f build/simple.yaml

kubectl get po

kubectl port-forward voyager-admin-666fc55cfb-wg5xb 14891

cat resources/local.json| prototool grpc --address localhost:14891 --method voyager.Admin/StartProbe --stdin | jq

kubectl delete -f build/simple.yaml

istioctl kube-inject -f build/simple.yaml | kubectl apply -f -

kubectl get po

kubectl port-forward voyager-admin-666fc55cfb-wg5xb 14891

cat resources/local.json| prototool grpc --address localhost:14891 --method voyager.Admin/StartProbe --stdin | jq

kubectl delete -f build/simple.yaml

istioctl kube-inject -f build/simple.yaml | kubectl delete -f -
