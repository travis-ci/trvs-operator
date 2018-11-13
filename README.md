# trvs-operator

trvs-operator is a Kubernetes controller that makes it easy to manage secrets for Travis CI services.

With this operator running in your cluster, you can create a resource like this:

```
apiVersion: travisci.com/v1
kind: TrvsSecret
metadata:
  name: worker-org
spec:
  app: macstadium-workers
  key: production-common
  prefix: TRAVIS_WORKER
```

And a Kubernetes `Secret` resource with the appropriate secret data will be automatically created and managed. The `TrvsSecret` resources can be committed to public repositories without exposing secret data.
