#1 /usr/bin/env sh

set -e

# wait for business_account_service
kubectl rollout status deployment/business_account_service --timeout=3m

# test business_account_service
helm test business_account_service
