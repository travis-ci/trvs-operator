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

Unfortunately, getting this operator up and running in the cluster is a bit non-trivial.

1. Create SSH keys for the three repositories:

    ```sh
    $ ssh-keygen -t rsa
    Generating public/private rsa key pair.
    Enter file in which to save the key (/Users/matt/.ssh/id_rsa): trvs.key
    Enter passphrase (empty for no passphrase):
    Enter same passphrase again:
    Your identification has been saved in trvs.key.
    Your public key has been saved in trvs.key.pub.
    ...

    # Repeat for the other two repos: travis-keychain.key and travis-pro-keychain.key
    ```

2. Add the public keys for each repo as a deploy key in GitHub. Read-only permissions are sufficient for trvs-operator.

3. Create a Kubernetes secret which each private key as its own entry:

    ```sh
    $ kubectl create secret generic trvs-operator \
        --from-file=travis-keychain.key \
        --from-file=travis-pro-keychain.key \
        --from-file=trvs.key
    ```

4. Install the operator using Helm:

    ```sh
    $ helm install chart/trvs-operator --name=trvs-operator \
        --set 'ssh.secretName=trvs-operator' \
        --set 'keychains.org=<ssh URL to org keychain>' \
        --set 'keychains.com=<ssh URL to com keychain>' \
        --set 'trvsUrl=<ssh URL to trvs repo>'
    ```

You're done. Now the operator should be running and will be able to transform `TrvsSecret` resources into ordinary Kubernetes secrets.