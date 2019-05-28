# trvs-operator

trvs-operator is a Kubernetes controller that makes it easy to manage secrets for Travis CI services.

With this operator running in your cluster, you can create a resource like this:

```yaml
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

When you push changes to the master branch of the keychain repos, the operator should see the change within a few minutes and update the secrets appropriately. Once this has happened, you'll need to delete any existing pods that are using the secrets as environment variables and let them be recreated in order to use the new secret values. Environment variables can't be updated in-place.

## Setting up

Unfortunately, getting this operator up and running in the cluster is a bit non-trivial. We've made a small script that will guide you through it.

   ```sh
   $ ./bin/install-trvs-operator.sh
   ```

It will:

1. Create SSH keys for the three repositories.

2. Ask you to add the public keys for each repo as a deploy key in GitHub. Read-only permissions are sufficient for trvs-operator.

3. Create a Kubernetes secret which each private key as its own entry.

4. Install the operator using Helm.

You're done. Now the operator should be running and will be able to transform `TrvsSecret` resources into ordinary Kubernetes secrets.

NOTE: The keys are now also on your machine, you likely want to remove them at some point.
