# FIPS compliance for conjur-authn-k8s-client

## Background

The National Institute of Standards and Technology (NIST) issued the [FIPS
 140-2](https://csrc.nist.gov/publications/detail/fips/140/2/final) 
Publication Series to coordinate the requirements and standards for cryptography 
modules that include both hardware and software components.

By making the conjur-authn-k8s-client FIPS compliant, our customer will gain a 
cryptographic standard to protect their unclassified, but sensitive data.

It should be noted that this project isn't only available for use for big 
enterprise players - it's also a key part of our open source suite, and open 
source users may very well not care about FIPS and may not have purchased 
access to RedHat's tooling. Throughout the design and implementation
we should verify that non-enterprise users get the same quality from the project,
even if it is FIPS compliant.

## Useful Links

| **Name**                        | **Link**                                            |
|---------------------------------|-----------------------------------------------------|
| Epic - FIPS compliance for conjur-authn-k8s-client | https://github.com/cyberark/conjur-authn-k8s-client/issues/91 |
| FIPS compliant crypto in golang | https://kupczynski.info/2019/12/15/fips-golang.html |

## Solution

We looked into 3 options for making our product FIPS compliant:

1. Use cgo to call out to an existing certified library
   
    - This alternative disqualified due to its complexity and consuming time
     and effort as we will have to replace any usage of crypto.
     
1. Use [RHEL go toolchain](https://developers.redhat.com/blog/2019/06/24/go-and-fips-140-2-on-red-hat-enterprise-linux/).
   - RHEL takes the ownership on the encryption, and bridges between the
    Golang encryption and OpenSSL.
   - What needs to be done is build the Golang project on subscribed RHEL
    machine/UBI container with the relevant go-toolset installed on it. Then, 
    we can take the compiled binary and copy it into the containers as usual.
   - A big advantage of this approach is that we are aligning with Red Hat
    which is one of the big players in the field of certifications, enterprise, 
    and federal market.
   - The pipeline should use and subscribe RHEL / UBI on-the-fly.
   
1. Use BoringSSL based crypto
   Use [Google’s fork of Golang](https://github.com/golang/go/blob/dev.boringcrypto.go1.12/misc/boring/README.md) 
   that uses dev.boringcrypto as its crypto engine. dev.boringcrypto wraps BoringSSL which is FIPS compliant.
   More info can be found in the [Security section](#security).
   
Looking into the options above, it was clear that the best option is BoringSSL.
Using cgo (option 1) was disqualified due to its complexity and consuming time
 and effort as we will have to replace any usage of crypto.
 
Using the RHEL go toolchain has its infrastructure complexity as we will need
 to build a new pipeline for building the project on a subscribed RHEL machine.
 
Luckily, Google maintains the BoringSSL which is very easy to use and gives
 us FIPS compliancy with minimal work. All that needs to be done is:
 - Replace the base image from `golang` to the corresponding `goboring/golang
 `. For example, `golang:1.12` should turn into `goboring/golang:1.12.17b4`
 - Enable cgo by setting `CGO_ENABLED=1` in the env
 - Verify that the `authenticator` binary uses the BoringCrypto libraries by
  running `go tool nm authenticator` and verifying that it has symbols named
   *_Cfunc__goboringcrypto_*
   
More info on the required changes can be found [here](https://github.com/cyberark/conjur-authn-k8s-client/pull/97)

Note: the UX will not change in this solution.

### Non-FIPS compliant image

We should decide if we still want to release a version of the `conjur-authn-k8s-client`
that is not FIPS compliant (i.e that uses `golang` as the base image). Let's
look into some pros and cons of both approaches.
 
 | **Option**                             | **Implications**                                                                                                                                                                                                                                                                                                                                                        |
 |----------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
 | Release only `goboring/golang` version | - Customers use the client and don't require FIPS compliancy, will now miss libraries that are available in `golang` but are not available in `goboring/golang`                                                                                                                                                                                                       |
 | Release both versions                  | <ul><li>One more Dockerfile to maintain. Any change made to the original Dockerfile should be done also to `Dockerfile.fips` which can be easily forgotten</li><li>Another version to test. Our UTs will stay the same but we will need to run integration tests on both images, in order to release them. This adds complexity to our tests and doubles the tests run-time</li></ul> |

From the table above, I think that we should release only the FIPS compliant
version. The advantages that customers will get by using the `golang` image
as the base image is not worth the maintenance burden that releasing both
versions will bring to the development team.

Note: If the performance of the `goboring/golang` version will be much worse than that
      of the `golang` version (degradation of more than 5%) then we will release both versions.
      More info on the performance tests that will be done in this effort can be found in the
      [Performance section](#performance)

### Red Hat image

We will also release a FIPS compliant Red Hat image of the `conjur-authn-k8s
-client` to our Red Hat registry.

### Performance

Same SLA of performance should be kept. As of today, we don't have any
performance tests and an SLA (that I know of). We should decide on one and
verify that the change to GoBoring doesn't impact the performance.

We should write a performance test and run it twice - once on the `master` 
image and once on the new FIPS-compliance image - and verify that the
 performance is not affected by the change. We can live with a degradation of less
 than 5% between the versions.

This performance test should be (in high level):
  - Deploy a pod with the `conjur-authn-k8s-client` image and an application container
    that retrieves a secret from Conjur using the access token retrieved by the authenticator client.
  - Test how long the procedure above took
  - Run this for 1000 times and get the average time

## Test plan

Before we dig into which tests should run, we should decide where, and how, they
should run. Currently, the `conjur-authn-k8s-client` have UTs in its project
and a vanilla test that runs in `kubernetes-conjur-demo`. The `kubernetes-conjur-demo`
is triggered daily and when we push code to that repository., and pulls from `latest` tag of 
`conjur-authn-k8s-client`. 

Even while putting aside the fact that we have only a vanilla test, our current
 test flow is still not optimal. 
 Let's propose some options to improve our flow, and decide on the best one.

### Trigger `kubernetes-conjur-demo` to run after every build of `conjur-authn-k8s-client`

We can find a way to run `kubernetes-conjur-demo` using the currently built
 authenticator-client, utilizing the `LOCAL_AUTHENTICATOR` environment variable.
 If this is possible, we can trigger a `kubernetes-conjur-demo` build as part
  of our build.

This will give us full confidence in the green build as it also passed
 integration tests.
 
 However, the build time of `kubernetes-conjur-demo` is ~25 minutes (opposing
  2-3 minutes of the `conjur-authn-k8s-client`) and it is not very stable as it runs
  against multiple environments.
  . This means that it will harder to merge PRs into the `conjur-authn-k8s
  -client`.
  
Furthermore, it won't be easy to debug in case of a failure as we will need to jump between builds
and find the errors. In addition, `kubernetes-conjur-demo` runs also tests for
the `secretless-broker` so failures there will fail builds of the authn-client.
  
I am not sure that this can be implemented. By setting `LOCAL_AUTHENTICATOR
` to `true` we use the last built image (ran in the `./bin/build` step. In
 case we have builds running in parallel it may affect the latest built image
  so it is possible that we can't use a local authenticator in Jenkins.

### Trigger `kubernetes-conjur-demo` to run after every master build of `conjur-authn-k8s-client`

This option will still not fail `conjur-authn-k8s-client` builds that will
fail in `kubernetes-conjur-demo` but we will get _some_ feedback. 
After we will merge the PR into `master` we can follow the `master` build of 
`kubernetes-conjur-demo` to see if it passed. This is better than a nightly
build also because at the end of the day we will see which PR
introduced a failure and it will be easier to fix the error. 
  
### Add integration tests to `conjur-authn-k8s-client` as we have in the `secrets-provider-for-kubernetes`
  
In the `secrets-provider-for-kubernetes` we have integration tests
with different scenarios, that each one deploys its own pod and verifies
the expected output.

The caveat of this approach is that members of the community will not be able
to contribute to this repo as they can today as they will not be able to run
the integration tests (as they run with `summon` which requires `conjurops
` access). They will not be able to fix tests that broke because of their
change and will not be able to add tests for their contribution. However, this
is not really different than today where community members cannot run the
tests in `kubernetes-conjur-demo`, so it's better that tests will fail
before we merge their PRs.

### Summary

To summarize the options, let's look at the following table:
| **Option**                                                                             | **Build Time** | **Design and Implementation time**                                                                                                                  | **Feedback on Failure**                                                           | **Ease of adding tests**                                                                                   | **Community Contribution**                                                                                                                                                                                                      |
|----------------------------------------------------------------------------------------|----------------|------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| (a) Trigger `kubernetes-conjur-demo` to run after every build of `conjur-authn-k8s-client` | ~30 minutes    | Depends on whether we can utilize the LOCAL_AUTHENTICATOR environment variable. If we can then the implementation time should take 2-3 days | Immediate feedback as all integration tests will  run before the end of the build | Adding tests to this project is not simple. Especially error flows.                                        | The community member will not be able to investigate a failure in case  there was one so the build will never be green (unless they ask for help from a Conjur team member).  There is also no option to add integration tests. |
| (b) Trigger `kubernetes-conjur-demo` to run after every master build                       | ~3 minutes     | 1 hour                                                                                                                                   | After nightly build and only after the change is merged into `master`             | Adding tests to this project is not simple. Especially error flows.                                        | The community member will still be able to contribute as the tests will not run as part of the build                                                                                                                            |
| (c) secrets-provider model - deploy different scenarios and check output                   | ~30 minutes    | ~10 days                                                                                                                                  | Immediate feedback as all integration tests will run before the end of the build  | - Should be easier to add tests. Adding tests to the `secrets-provider-for-k8s` is pretty straight forward | The community member will not be able to investigate a failure in case there was one so the build will never be green (unless they askfor help from a Conjur team member). There is also no option to add integration tests.    |

### Decision

I would not implement option (a) as it has some work to do and the solution is not optimal.

Option (b) is the easiest one but it won't give us the required confidence before
merging to master so we should avoid that approach.

My suggestion is to implement option (c) for 2 main reasons:
  - Integration tests are maintained in the same repo as the application code 
  - Integration tests run as part of the build
  
We will need to further research and design the best solution for our needs, before
we can implement it. We should look into [`KInD`](https://github.com/kubernetes-sigs/kind)
that seems to be suitable. Once the design is ready we will perform a meeting including 
Architects and Repository Owner. 

We will not implement the tests in bash scripts like we do in the `secrets-provider-for-k8s`.

### Integration Tests

Regardless of how we will run our tests, it is not optimal that we have only
 a vanilla flow. We should add another test where in case the authenticator
 -client fails to authenticate with Conjur we don't provide an access token
  and the log shows `CAKC015E Login failed`.
  
We do not need to test different permutations of error flows (e.g host does
not exist, host is not permitted on the authn-k8s/prod authenticator) as
these test run in the `conjur` repository. As far as the authenticator-client
 is concerned, the output is binary - Authentication success or failure.
   
| **Scenario**            | **Given**                                                    | **When**                                                 | **Then**                                                                                               |
|-------------------------|--------------------------------------------------------------|----------------------------------------------------------|--------------------------------------------------------------------------------------------------------|
| Authentication succeeds | A Running Conjur cluster with a configured K8s Authenticator | I run the authenticator client with a valid k8s host     | An access token is provided to the application container and it can retrieve a secret with it          |
| Authentication fails    | A Running Conjur cluster with a configured K8s Authenticator | I run the authenticator client with a non-valid k8s host | An access token is not provided to the application container and the log shows `CAKC015E Login failed` |

## Docs

We should document any change that will affect the customer (e.g if we release 
both FIPS & Non-FIPS versions).

If no customer-facing changes are introduced, we will not need to add documentation.
The current design does not introduce any customer-facing changes.

## Version update

We should update the version of:
  - `conjur-authn-k8s-client`
  - `seed-fetcher`
  
We don't need to update the version of the `secretless-broker` and the
`secrets-provider-for-k8s` as they only consume libraries of the `conjur-authn-k8s-client`
and will not be affected by the base image change. 

## Security

Google is maintaining this 
fork of Go that uses BoringCrypto and they [intend to maintain in this branch 
the latest release plus BoringCrypto patches](https://go.googlesource.com/go/+/refs/heads/dev.boringcrypto/README.boringcrypto.md).
You can read more about [FIPS certification of the crypto library of dev.boringcrypto](https://csrc.nist.gov/CSRC/media/projects/cryptographic-module-validation-program/documents/security-policies/140sp3318.pdf).

Furthermore, the DoD of this effort is to be able to tell customers that we are using only 
FIPS compliant crypto libraries and not that our product is FIPS compliant by itself. 
Our solution meets that requirement so it meets the security needs.

## Delivery Plan

The delivery plan will include the following steps:
  - Test performance and compare between versions
    - EE: 2 days
  - Implement fips-compliancy
    - Finalize the open PR for the implementation. Most of the work is already done.
    - EE: 2 days
  - Implement tests
    - Research and design the solution 
    - Implement the tests
    - EE: 10 days (~5 for design & ~5 for implementation)
  - Update versions
    - EE: 1 day
  - Add documentation (if needed)
    - EE: 1 day

## Open Issues

- Will we release a Non-FIPS version?
  - Depends on the performance test outcome
- How will we test the project?
  - Depends on research 
  - Mitigation plan: in case the solution we decide on will require a long implementation
    time (the definition of that will be decided after the design) we can go with
    option (b) for the FIPS compliancy effort and improve the tests infrastructure
    straight after that. 

## DoD

- [ ] Solution design is approved by the following personas
  - [ ] Team Leader: Dekel Asaf (@Tovli)
  - [ ] Product Owner: Hila Gross (@hilagross)
  - [ ] Repository Owner: Geri Jennings (@izgeri)
  - [ ] QA Architect: Elad Kugman (@eladkug)
  - [ ] System Architect: Ofira Burstein (@oburstein-hub)
  - [ ] Security Architect: Shahar Glazner: (@shaharglazner)
- [ ] `conjur-authn-k8s-client` is FIPS compliant
- [ ] Tests are implemented according to Test Plan and are passing
- [ ] Required documentation changes are implemented
- [ ] Versions are bumped in all relevant projects


