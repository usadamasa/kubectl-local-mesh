#!/bin/bash
set -euo pipefail

echo "Waiting for kubeconfig..."
while [ ! -f /output/kubeconfig.yaml ]; do
    sleep 1
done
echo "kubeconfig found"

# kubeconfigのサーバーアドレスをk3sに変更
echo "Updating kubeconfig server address..."
sed -i 's/127.0.0.1/k3s/g' /output/kubeconfig.yaml
export KUBECONFIG=/output/kubeconfig.yaml

# K8sクラスタが準備できるまで待機
echo "Waiting for K8s API server..."
until kubectl get nodes >/dev/null 2>&1; do
    sleep 2
done
echo "K8s API server ready"

# テストサービスをデプロイ
echo "Deploying test services..."
kubectl apply -k /manifests/

echo "Waiting for pods to be ready..."
kubectl wait --for=condition=ready pod -l app=http-test-server -n e2e-test --timeout=120s
kubectl wait --for=condition=ready pod -l app=grpc-test-server -n e2e-test --timeout=120s
echo "Test services ready"

# kubectl-localmesh を起動
echo "Starting kubectl-localmesh..."
exec kubectl-localmesh "$@"
