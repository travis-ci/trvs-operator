#!/usr/bin/env bash
set -uo pipefail

helm del --purge trvs-operator

kubectl delete secret trvs-operator
kubectl delete customresourcedefinitions.apiextensions.k8s.io trvssecrets.travisci.com
