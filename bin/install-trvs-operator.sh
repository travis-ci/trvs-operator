#!/usr/bin/env bash
set -euo pipefail

# Create keys
mkdir -p keys
cd keys

FILENAME=travis-keychain.key
[ -f $FILENAME ] || ssh-keygen -t rsa -N '' -C kubernetes@travis -f $FILENAME
echo
echo "Add this key to travis-keychain deploy keys:"
echo "https://github.com/travis-pro/travis-keychain/settings/keys/new"
echo
cat $FILENAME.pub
echo
read -p "Done?" choice
case "$choice" in
  * ) echo "ok";;
esac


FILENAME=travis-pro-keychain.key
[ -f $FILENAME ] || ssh-keygen -t rsa -N '' -C kubernetes@travis -f $FILENAME
echo
echo "Add this key to travis-pro-keychain deploy keys:"
echo "https://github.com/travis-pro/travis-pro-keychain/settings/keys/new"
echo
cat $FILENAME.pub
echo
read -p "Done?" choice
case "$choice" in
  * ) echo "ok";;
esac


FILENAME=trvs.key
[ -f $FILENAME ] || ssh-keygen -t rsa -N '' -C kubernetes@travis -f $FILENAME
echo
echo "Add this key to trvs deploy keys:"
echo "https://github.com/travis-ci/trvs/settings/keys/new"
echo
cat $FILENAME.pub
echo
read -p "Done?" choice
case "$choice" in
  * ) echo "ok";;
esac

# Install keys
kubectl create secret generic trvs-operator \
    --from-file=travis-keychain.key \
    --from-file=travis-pro-keychain.key \
    --from-file=trvs.key

cd -

# Install trvs-operator chart
helm install chart/trvs-operator --name=trvs-operator \
    --set 'image.tag=v1.0.1' \
    --set 'ssh.secretName=trvs-operator' \
    --set 'keychains.org=git@github.com:travis-pro/travis-keychain.git' \
    --set 'keychains.com=git@github.com:travis-pro/travis-pro-keychain.git' \
    --set 'trvsUrl=git@github.com:travis-ci/trvs.git'

echo Done
