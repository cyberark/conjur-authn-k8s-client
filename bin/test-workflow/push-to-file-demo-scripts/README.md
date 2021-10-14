# Scripts to run M1 E2E Demo

The scripts that are in this `push-to-file-demo-scripts` subdirectory are
intended to be used as a demonstration of the Secrets Provider "Push to File"
feature.

## Clone the `cyberark/conjur-authn-k8s-client` Repo

If you haven't already done so, clone the `cyberark/conjur-authn-k8s-client`
GitHub repository. For example:

```
cd
mkdir -p cyberark
cd cyberark
git clone https://github.com/cyberark/conjur-authn-k8s-client
cd conjur-authn-k8s-client
```

## Run the E2E Test Workflow Scripts to Spin Up a Push-to-File Demo Environment

Set up a Secrets Provider Push-to-File demo environment by running the start
script:

```
cd bin/test-workflow/push-to-file-demo-scripts
./start
```

## Use the Push-to-File Demo Scripts to Modify Annotations and Display Secret Files

After a demo environment has been set up, run any of the
`apply-patch\*` scripts in the `my-patch-scripts/` directory, and
the script will use `kubectl patch ...` to add Annotations to the
demo application deployment, and then show the resulting secrets files
that are generated in the application Pod's shared volume.

## Cleaning up the Demo Environment

To clean up the demo environment, run:

```
kubectl delete namespace app-test
kubectl delete namespace conjur-oss
```
