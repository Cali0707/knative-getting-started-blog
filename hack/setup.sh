#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

echo "Installing Knative Serving Custom Resources..."
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.12.3/serving-crds.yaml

echo "Installing Knative Serving Core Components..."
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.12.3/serving-core.yaml

echo "Installing Knative Kourier Network Layer..."
kubectl apply -f https://github.com/knative/net-kourier/releases/download/knative-v1.12.3/kourier.yaml

echo "Configuring Knative Serving to use Kourier..."
kubectl patch configmap/config-network \
  --namespace knative-serving \
  --type merge \
  --patch '{"data":{"ingress-class":"kourier.ingress.networking.knative.dev"}}'

echo "Waiting for all Knative Serving pods to be in the Running state..."

while IFS= read -r line; do
    POD_NAME=$(echo $line | awk '{print $1}')
    POD_STATUS=$(echo $line | awk '{print $3}')
    if [ "$POD_STATUS" != "Running" ]; then
        echo "Waiting for pod $POD_NAME to be in Running state..."
        sleep 5
        kubectl get pods --namespace knative-serving | grep -v NAME | while IFS= read -r line; do continue; done
    fi
done < <(kubectl get pods --namespace knative-serving | grep -v NAME)

echo "All pods are running."

echo "Verifying Knative Serving Installation..."
kubectl get pods --namespace knative-serving

echo "<Success!> Knative Serving Installation Completed Successfully!"



echo "Installing Knative Eventing Custom Resource Definitions..."
kubectl apply -f https://github.com/knative/eventing/releases/download/knative-v1.12.4/eventing-crds.yaml

echo "Installing Knative Eventing Core Components..."
kubectl apply -f https://github.com/knative/eventing/releases/download/knative-v1.12.4/eventing-core.yaml

echo "Waiting for all Knative Eventing pods to be in the Running or Completed state..."

while IFS= read -r line; do
    POD_NAME=$(echo $line | awk '{print $1}')
    POD_STATUS=$(echo $line | awk '{print $3}')
    if [ "$POD_STATUS" != "Running" ] && [ "$POD_STATUS" != "Completed" ]; then
        echo "Waiting for pod $POD_NAME to reach Running or Completed state..."
        sleep 5
        kubectl get pods -n knative-eventing | grep -v NAME | while IFS= read -r line; do continue; done
    fi
done < <(kubectl get pods -n knative-eventing | grep -v NAME)

echo "All Knative Eventing pods are in the desired state."

echo "<Success!>Knative Eventing Installation Completed Successfully!"


echo "Before we proceed with installing the config, we need to check whether ko is installed or not."

if command -v ko >/dev/null 2>&1; then
    echo "ko is installed, proceeding with the installation of the config."
else
    echo "ko is not installed. Please manually install ko before proceeding."
    exit 1
fi

echo "Installing the config..."
ko apply -f config/

echo "Waiting for all pods to be in the Running or Completed state..."

while IFS= read -r line; do
    POD_NAME=$(echo $line | awk '{print $1}')
    POD_STATUS=$(echo $line | awk '{print $3}')
    if [ "$POD_STATUS" != "Running" ] && [ "$POD_STATUS" != "Completed" ]; then
        echo "Waiting for pod $POD_NAME to reach Running or Completed state..."
        sleep 5
        kubectl get pods -n knative-eventing | grep -v NAME | while IFS= read -r line; do continue; done
    fi
done < <(kubectl get pods -n knative-eventing | grep -v NAME)

echo "All the pods are in the desired state."

echo "<Success!> Config Installation Completed Successfully!"

echo "Install the sample source and event display"
kubectl apply -f example/

echo "Waiting for all pods to be in the Running or Completed state..."

while IFS= read -r line; do
    POD_NAME=$(echo $line | awk '{print $1}')
    POD_STATUS=$(echo $line | awk '{print $3}')
    if [ "$POD_STATUS" != "Running" ] && [ "$POD_STATUS" != "Completed" ]; then
        echo "Waiting for pod $POD_NAME to reach Running or Completed state..."
        sleep 5
        kubectl get pods -n knative-eventing | grep -v NAME | while IFS= read -r line; do continue; done
    fi
done < <(kubectl get pods -n knative-samples | grep -v NAME)

echo "All the pods are in the desired state. You are all set."
