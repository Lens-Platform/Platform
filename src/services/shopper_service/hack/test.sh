#1 /usr/bin/env sh

set -e

# wait for shopper_service
kubectl rollout status deployment/shopper_service --timeout=3m

# test shopper_service
helm test shopper_service
