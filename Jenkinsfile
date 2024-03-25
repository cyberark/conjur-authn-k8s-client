#!/usr/bin/env groovy

@Library("product-pipelines-shared-library") _

// Automated release, promotion and dependencies
properties([
  // Include the automated release parameters for the build
  release.addParams(),
  // Dependencies of the project that should trigger builds
  dependencies(['conjur-enterprise/conjur-opentelemetry-tracer'])
])

// Performs release promotion.  No other stages will be run
if (params.MODE == "PROMOTE") {
  release.promote(params.VERSION_TO_PROMOTE) { infrapool, sourceVersion, targetVersion, assetDirectory ->
    // Any assets from sourceVersion Github release are available in assetDirectory
    // Any version number updates from sourceVersion to targetVersion occur here
    // Any publishing of targetVersion artifacts occur here
    // Anything added to assetDirectory will be attached to the Github Release

    // Pull existing images from internal registry in order to promote
    infrapool.agentSh """
      export PATH="release-tools/bin:${PATH}"
      docker pull registry.tld/conjur-authn-k8s-client:${sourceVersion}
      docker pull registry.tld/conjur-authn-k8s-client-redhat:${sourceVersion}
      docker pull cyberark/conjur-k8s-cluster-test:${sourceVersion}
      docker tag cyberark/conjur-k8s-cluster-test:${sourceVersion} conjur-k8s-cluster-test:${sourceVersion}
      // Promote source version to target version.
      summon ./bin/publish --promote --source ${sourceVersion} --target ${targetVersion}
    """
  }

  // Copy Github Enterprise release to Github
  release.copyEnterpriseRelease(params.VERSION_TO_PROMOTE)
  return
}

pipeline {
  agent { label 'conjur-enterprise-common-agent' }

  options {
    timestamps()
    buildDiscarder(logRotator(numToKeepStr: '30'))
  }

  triggers {
    cron(getDailyCronString())
  }

  environment {
    // Sets the MODE to the specified or autocalculated value as appropriate
    MODE = release.canonicalizeMode()
  }

  parameters {
    booleanParam(
      name: 'TEST_OCP_NEXT',
      description: 'Whether or not to run the pipeline against the next OCP version',
      defaultValue: false)
  }

  stages {
    // Aborts any builds triggered by another project that wouldn't include any changes
    stage ("Skip build if triggering job didn't create a release") {
      when {
        expression {
          MODE == "SKIP"
        }
      }
      steps {
        script {
          currentBuild.result = 'ABORTED'
          error("Aborting build because this build was triggered from upstream, but no release was built")
        }
      }
    }

    stage('Get InfraPool AzureExecutorV2 and ExecutorV2ARM Agents') {
      steps {
        script {
          // Request ExecutorV2 agents for 2 hour(s)
          INFRAPOOL_AZURE_AGENT_0 = getInfraPoolAgent.connected(type: "AzureExecutorV2", quantity: 1, duration: 2)[0]
          // Request ExecutorV2ARM agents for 2 hour(s)
          INFRAPOOL_EXECUTORV2ARM_AGENT_0 = getInfraPoolAgent.connected(type: "ExecutorV2ARM", quantity: 1, duration: 2)[0]
        }
      }
    }

    stage('Get latest upstream dependencies') {
      steps {
        script {
          updatePrivateGoDependencies("${WORKSPACE}/go.mod")
          // Copy the vendor directory onto AMD64 infrapool
          INFRAPOOL_AZURE_AGENT_0.agentPut from: "vendor", to: "${WORKSPACE}"
          INFRAPOOL_AZURE_AGENT_0.agentPut from: "go.*", to: "${WORKSPACE}"
          // Add GOMODCACHE directory to infrapool allowing automated release to generate SBOMs
          INFRAPOOL_AZURE_AGENT_0.agentPut from: "/root/go", to: "/var/lib/jenkins/"
          // Add GOMODCACHE directory for Azure ubuntu 20.04 (can be removed after os upgrade in favor of the above line)
          INFRAPOOL_AZURE_AGENT_0.agentPut from: "/root/go", to: "/home/jenkins/"

          // Copy the vendor directory onto ARM64 infrapool
          INFRAPOOL_EXECUTORV2ARM_AGENT_0.agentPut from: "vendor", to: "${WORKSPACE}"
          INFRAPOOL_EXECUTORV2ARM_AGENT_0.agentPut from: "go.*", to: "${WORKSPACE}"
          // Add GOMODCACHE directory to infrapool allowing automated release to generate SBOMs
          INFRAPOOL_EXECUTORV2ARM_AGENT_0.agentPut from: "/root/go", to: "/var/lib/jenkins/"
          // Add GOMODCACHE directory for Azure ubuntu 20.04 (can be removed after os upgrade in favor of the above line)
          INFRAPOOL_EXECUTORV2ARM_AGENT_0.agentPut from: "/root/go", to: "/home/jenkins/"
        }
      }
    }

    stage('Validate') {
      parallel {
        stage('Changelog') {
          steps {
            script {
              parseChangelog(INFRAPOOL_AZURE_AGENT_0)
            }
          }
        }

        stage('Log messages') {
          steps {
            validateLogMessages()
          }
        }

        stage('Cluster-Prep Schema') {
          steps { script { INFRAPOOL_AZURE_AGENT_0.agentSh './bin/validate-schema ./helm/conjur-config-cluster-prep/values.schema.json'} }
        }

        stage('Application Namespace-Prep Schema') {
          steps { script { INFRAPOOL_AZURE_AGENT_0.agentSh './bin/validate-schema ./helm/conjur-config-namespace-prep/values.schema.json'} }
        }

        stage('Helm Chart Unit Tests') {
          steps { script { INFRAPOOL_AZURE_AGENT_0.agentSh './bin/test-helm-unit-in-docker' } }
        }
      }
    }

    // Generates a VERSION file based on the current build number and latest version in CHANGELOG.md
    stage('Validate Changelog and set version') {
      steps {
        updateVersion(INFRAPOOL_AZURE_AGENT_0, "CHANGELOG.md", "${BUILD_NUMBER}")
        updateVersion(INFRAPOOL_EXECUTORV2ARM_AGENT_0, "CHANGELOG.md", "${BUILD_NUMBER}")
      }
    }

    stage('Build client Docker image') {
      parallel {
        stage('Build AMD64 image') {
          steps {
            script {
              INFRAPOOL_AZURE_AGENT_0.agentSh './bin/build'
            }
          }
        }
        stage('Build ARM64 image') {
          steps {
            script {
              INFRAPOOL_EXECUTORV2ARM_AGENT_0.agentSh './bin/build'
            }
          }
        }
      }
    }

    stage('Run Tests') {
      steps {
        script {
          INFRAPOOL_AZURE_AGENT_0.agentSh './bin/test'
        }
      }
      post {
        always {
          script {
            INFRAPOOL_AZURE_AGENT_0.agentSh './bin/coverage'
            INFRAPOOL_AZURE_AGENT_0.agentSh 'cp ./test/c.out ./c.out'
            INFRAPOOL_AZURE_AGENT_0.agentStash name: 'coverage-report', includes: 'test/*'
            unstash 'coverage-report'
            junit 'test/junit.xml'
            cobertura autoUpdateHealth: true, autoUpdateStability: true, coberturaReportFile: 'test/coverage.xml', conditionalCoverageTargets: '70, 0, 70', failUnhealthy: true, failUnstable: true, maxNumberOfBuilds: 0, lineCoverageTargets: '70, 70, 70', methodCoverageTargets: '70, 0, 70', onlyStable: false, sourceEncoding: 'ASCII', zoomCoverageChart: false
            codacy action: 'reportCoverage', filePath: "test/coverage.xml"
          }
        }
      }
    }

    stage("Scan images") {
      parallel {
        stage ("Scan main image for fixable vulns") {
          steps {
            scanAndReport(INFRAPOOL_AZURE_AGENT_0, "conjur-authn-k8s-client:dev", "HIGH", false)
          }
        }
        stage ("Scan main image for total vulns") {
          steps {
            scanAndReport(INFRAPOOL_AZURE_AGENT_0, "conjur-authn-k8s-client:dev", "NONE", true)
          }
        }
        stage ("Scan redhat image for fixable vulns") {
          steps {
            scanAndReport(INFRAPOOL_AZURE_AGENT_0, "conjur-authn-k8s-client-redhat:dev", "HIGH", false)
          }
        }
        stage ("Scan redhat image for total vulns") {
          steps {
            scanAndReport(INFRAPOOL_AZURE_AGENT_0, "conjur-authn-k8s-client-redhat:dev", "NONE", true)
          }
        }
        stage ("Scan helm test image for fixable vulns") {
          steps {
            scanAndReport(INFRAPOOL_AZURE_AGENT_0, "conjur-k8s-cluster-test:dev", "HIGH", false)
          }
        }
        stage ("Scan helm test image for total vulns") {
          steps {
            scanAndReport(INFRAPOOL_AZURE_AGENT_0, "conjur-k8s-cluster-test:dev", "NONE", true)
          }
        }
      }
    }

    stage('E2E Workflow Tests') {
      stages {
        stage('Update Helm dependencies') {
          /*
           * Helm dependency update is done before running E2E tests in parallel
           * since this is not currently thread-safe (Helm chart downloads use
           * a non-uniquely named 'tmpcharts' directory and fail if the directory
           * already exists).
          */
          steps {
            script {
              INFRAPOOL_AZURE_AGENT_0.agentSh './bin/helm-dependency-update-in-docker'
            }
          }
        }
        stage('Test app with') {
          parallel {
            stage('Enterprise in GKE') {
              steps {
                script {
                  INFRAPOOL_AZURE_AGENT_0.agentSh 'cd bin/test-workflow && summon --environment gke ./start --enterprise --platform gke --ci-apps'
                }
              }
            }
            stage('OSS in OpenShift v(current)') {
              steps {
                script {
                  INFRAPOOL_AZURE_AGENT_0.agentSh 'cd bin/test-workflow && summon --environment openshift -D ENV=ci -D VER=current ./start --platform openshift --ci-apps'
                }
              }
            }
            stage('OSS in OpenShift v(next)') {
              when {
                expression { params.TEST_OCP_NEXT }
              }
              steps {
                script {
                  INFRAPOOL_AZURE_AGENT_0.agentSh 'cd bin/test-workflow && summon --environment openshift -D ENV=ci -D VER=next ./start --platform openshift --ci-apps'
                }
              }
            }
          }
        }
        stage('Enterprise in Jenkins') {
          stages {
            // stage('Test app in GKE') {
            //   steps {
            //     script {
            //       INFRAPOOL_AZURE_AGENT_0.agentSh '''
            //         HOST_IP="$(curl https://checkip.amazonaws.com)";
            //         echo "HOST_IP=${HOST_IP}"
            //         echo "CONJUR_APPLIANCE_TAG=${CONJUR_APPLIANCE_TAG}"
            //         cd bin/test-workflow && summon --environment gke ./start --enterprise --platform jenkins --ci-apps
            //       '''
            //     }
            //   }
            // }
            // stage('Test app in OpenShift v(current)') {
            //   steps {
            //     script {
            //       INFRAPOOL_AZURE_AGENT_0.agentSh '''
            //         HOST_IP="$(curl https://checkip.amazonaws.com)";
            //         echo "HOST_IP=${HOST_IP}"
            //         cd bin/test-workflow && summon --environment openshift -D ENV=ci -D VER=current ./start --enterprise --platform jenkins --ci-apps
            //       '''
            //     }
            //   }
            // }
            stage('Test app in OpenShift v(next)') {
              when {
                expression { params.TEST_OCP_NEXT }
              }
              steps {
                script {
                  INFRAPOOL_AZURE_AGENT_0.agentSh '''
                    HOST_IP="$(curl https://checkip.amazonaws.com)";
                    echo "HOST_IP=${HOST_IP}"
                    cd bin/test-workflow && summon --environment openshift -D ENV=ci -D VER=next ./start --enterprise --platform jenkins  --ci-apps
                  '''
                }
              }
            }
          }
        }
      }
    }

    // Internal images will be used for promoting releases.
    stage('Push images to internal registry') {
      parallel {
        stage('Push images AMD64 image') {
          steps {
            script {
              // Push images to the internal registry so that they can be used
              // by tests, even if the tests run on a different executor.
              INFRAPOOL_AZURE_AGENT_0.agentSh './bin/publish --internal'
            }
          }
        }

        stage('Push images ARM64 image') {
          steps {
            script {
              // Push images to the internal registry so that they can be used
              // by tests, even if the tests run on a different executor.
              INFRAPOOL_EXECUTORV2ARM_AGENT_0.agentSh './bin/publish --internal --arch arm64'
            }
          }
        }
      }
    }

    stage('Push multi-arch manifest to internal registry') {
      steps {
        script {
          // Push multi-architecture manifest to the internal registry.
          INFRAPOOL_AZURE_AGENT_0.agentSh './bin/publish --manifest'
        }
      }
    }

    stage('Release') {
      when {
        expression {
          MODE == "RELEASE"
        }
      }

      steps {
        script {
          release(INFRAPOOL_AZURE_AGENT_0) { billOfMaterialsDirectory, assetDirectory, toolsDirectory ->
            // Publish release artifacts to all the appropriate locations
            // Copy any artifacts to assetDirectory to attach them to the Github release

            // Create Go application SBOM using the go.mod version for the golang container image
            INFRAPOOL_AZURE_AGENT_0.agentSh """export PATH="${toolsDirectory}/bin:${PATH}" && go-bom --tools "${toolsDirectory}" --go-mod ./go.mod --image "golang" --main "cmd/authenticator/" --output "${billOfMaterialsDirectory}/go-app-bom.json" """
            // Create Go module SBOM
            INFRAPOOL_AZURE_AGENT_0.agentSh """export PATH="${toolsDirectory}/bin:${PATH}" && go-bom --tools "${toolsDirectory}" --go-mod ./go.mod --image "golang" --output "${billOfMaterialsDirectory}/go-mod-bom.json" """
            // Publish edge release
            INFRAPOOL_AZURE_AGENT_0.agentSh 'export PATH="${toolsDirectory}/bin:${PATH}" && summon ./bin/publish --edge'

            // Publish internal edge release
            INFRAPOOL_AZURE_AGENT_0.agentSh 'export PATH="${toolsDirectory}/bin:${PATH}" && summon ./bin/publish --internal-edge'
          }
        }
      }
    }
  }

  post {
    always {
      releaseInfraPoolAgent(".infrapool/release_agents")
    }
  }
}
