# Contributing

Thanks for your interest in contributing to the Conjur authn-k8s client! In
this document you'll find guidelines for how to get started.

For general contribution and community guidelines, please see the [community repo](https://github.com/cyberark/community).

## Table of Contents

- [Development](#development)
- [Testing](#testing)
- [Releases](#releases)
- [Contributing](#contributing)

## Development

### Prerequisites

- To work in this codebase, you will want to have at least Go v1.13 installed. The
  code may work on older versions but it has not been tested nor evaluated to be
  compatible for those configurations. We expect at the very least that you will
  need Go modules support so you will at minimum need Go v1.11+.

## Testing

### Test Suites

To run the test suite, run `./bin/build` and `./bin/test`.

### Demo Workflow 

To run a sample deployment of the Helm charts located in `/helm`, run `./bin/test-workflow`. This
will download and run the `conjur-oss-helm-chart` project example, then consecutively install the 
`helm/conjur-config-cluster-prep`, `helm/conjur-config-namespace-prep`, and `helm/conjur-app-deploy` charts, in that order.

## Releases

Releases should be created by maintainers only. To create a tag and release,
follow the instructions in this section.

### Update the version, changelog, and notices
1. Create a new branch for the version bump.
1. Based on the unreleased content, determine the new version number and update
   the [version.go](pkg/authenticator/version.go) file.
1. Determine the new version number and update the Helm `Chart.yaml` files in the `helm/conjur-*/` directories.
1. Review the git log and ensure the [changelog](CHANGELOG.md) contains all
   relevant recent changes with references to GitHub issues or PRs, if possible.
1. Review the changes since the last tag, and if the dependencies have changed
   revise the [NOTICES](NOTICES.txt) to correctly capture the included
   dependencies and their licenses / copyrights.
1. Commit these changes - `Bump version to x.y.z` is an acceptable commit
   message - and open a PR for review.

### Add a git tag
1. Once your changes have been reviewed and merged into master, tag the version
   using `git tag -s v0.1.1`. Note this requires you to be  able to sign releases.
   Consult the [github documentation on signing commits](https://help.github.com/articles/signing-commits-with-gpg/)
   on how to set this up. `vx.y.z` is an acceptable tag message.
1. Push the tag: `git push vx.y.z` (or `git push origin vx.y.z` if you are working
   from your local machine).

### Publish the git release
1. In the GitHub UI, create a release from the new tag and copy the change log
   for the new version into the GitHub release description. The Jenkins pipeline 
   will auto-publish new images to DockerHub.
   
### Publish the Red Hat image
1. Visit the [Red Hat project page](https://connect.redhat.com/project/795581/view) once the images have
   been pushed and manually choose to publish the latest release.

## Contributing

1. [Fork the project](https://help.github.com/en/github/getting-started-with-github/fork-a-repo)
2. [Clone your fork](https://help.github.com/en/github/creating-cloning-and-archiving-repositories/cloning-a-repository)
3. Make local changes to your fork by editing files
3. [Commit your changes](https://help.github.com/en/github/managing-files-in-a-repository/adding-a-file-to-a-repository-using-the-command-line)
4. [Push your local changes to the remote server](https://help.github.com/en/github/using-git/pushing-commits-to-a-remote-repository)
5. [Create new Pull Request](https://help.github.com/en/github/collaborating-with-issues-and-pull-requests/creating-a-pull-request-from-a-fork)

From here your pull request will be reviewed and once you've responded to all
feedback it will be merged into the project. Congratulations, you're a
contributor!
