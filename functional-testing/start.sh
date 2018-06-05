#!/bin/bash
echo "______________________________________________________________________________"
echo ""
echo "    [*] starting minikube..."
echo "______________________________________________________________________________"
echo ""
export CNI_BRIDGE_NETWORK_OFFSET="0.0.1.0"
dockerd --host=unix:///var/run/docker.sock --host=tcp://0.0.0.0:2375 &> /var/log/docker.log 2>&1 < /dev/null &
/minikube start --vm-driver=none --extra-config=apiserver.Admission.PluginNames=Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,GenericAdmissionWebhook,ResourceQuota < /dev/null

c=0
while [ $(cat /var/log/docker.log | grep -c 'docker-containerd-shim started') -lt 9 ] ; do
    if [ $c -gt 60 ]; then
        echo "ERROR: failed waiting for minikube to come up..."
        exit 1
    fi
    sleep 10
    c=$((c+1))
done
echo "______________________________________________________________________________"
echo ""
echo "    [*] starting tiller..."
echo "______________________________________________________________________________"
echo ""
kubectl config view --merge=true --flatten=true > /kubeconfig
kubectl create serviceaccount --namespace kube-system cluster-admin
kubectl create clusterrolebinding cluster-rule --clusterrole=cluster-admin --serviceaccount=kube-system:cluster-admin
kubectl create clusterrolebinding cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:default 
kubectl --namespace kube-system apply -f /ca.yml
helm init --wait
helm repo add svc-cat https://svc-catalog-charts.storage.googleapis.com 

c=0
while [ "$(kubectl get pods -n kube-system | grep tiller | grep -c 1/1)" != "1" ] ; do
    if [ $c -gt 60 ]; then
        echo "ERROR: failed waiting for tiller to come up..."
        exit 1
    fi
    sleep 10
    c=$((c+1))
done
echo "______________________________________________________________________________"
echo ""
echo "    [*] starting service catalog..."
echo "______________________________________________________________________________"
echo ""
helm install svc-cat/catalog --name catalog --namespace catalog --wait --timeout 1200 --version 0.1.13 

c=0
while [ "$(kubectl get pods -n catalog | grep -c '2/2\|1/1')" != "2" ] ; do
    if [ $c -gt 60 ]; then
        echo "ERROR: failed waiting for service catalog to come up..."
        exit 1
    fi
    sleep 10
    c=$((c+1))
done

cd /

/bin/bash