#!/usr/bin/env groovy

// Automated release, promotion and dependencies
properties([
  // Include the automated release parameters for the build
  release.addParams(),
  // Dependencies of the project that should trigger builds
  dependencies(['cyberark/conjur-opentelemetry-tracer'])
])

// Performs release promotion.  No other stages will be run
if (params.MODE == "PROMOTE") {
  release.promote(params.VERSION_TO_PROMOTE) { sourceVersion, targetVersion, assetDirectory ->
    // Any assets from sourceVersion Github release are available in assetDirectory
    // Any version number updates from sourceVersion to targetVersion occur here
    // Any publishing of targetVersion artifacts occur here
    // Anything added to assetDirectory will be attached to the Github Release

    sh "docker pull cyberark/conjur-authn-k8s-client:${sourceVersion}"
    sh "summon ./bin/publish --source ${sourceVersion} --target ${targetVersion}"
  }
  return
}

pipeline {
  agent { label 'executor-v2' }

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

    stage('Validate') {
      parallel {
        stage('Changelog') {
          steps { sh './bin/parse-changelog.sh' }
        }

        stage('Cluster-Prep Schema') {
          steps { sh './bin/validate-schema ./helm/conjur-config-cluster-prep/values.schema.json'}
        }

        stage('Application Namespace-Prep Schema') {
          steps { sh './bin/validate-schema ./helm/conjur-config-namespace-prep/values.schema.json'}
        }

        stage('Helm Chart Unit Tests') {
          steps { sh './bin/test-helm-unit-in-docker' }
        }
      }
    }

    // Generates a VERSION file based on the current build number and latest version in CHANGELOG.md
    stage('Validate Changelog and set version') {
      steps {
        updateVersion("CHANGELOG.md", "${BUILD_NUMBER}")
      }
    }

    stage('Build client Docker image') {
      steps {
        sh './bin/build'
      }
    }

  /*
    stage('Run Tests') {
      steps {
        sh './bin/test'

        junit 'test/junit.xml'
      }
    }

    stage("Scan images") {
      parallel {
        stage ("Scan main image for fixable vulns") {
          steps {
            scanAndReport("conjur-authn-k8s-client:dev", "HIGH", false)
          }
        }
        stage ("Scan main image for total vulns") {
          steps {
            scanAndReport("conjur-authn-k8s-client:dev", "NONE", true)
          }
        }
        stage ("Scan redhat image for fixable vulns") {
          steps {
            scanAndReport("conjur-authn-k8s-client-redhat:dev", "HIGH", false)
          }
        }
        stage ("Scan redhat image for total vulns") {
          steps {
            scanAndReport("conjur-authn-k8s-client-redhat:dev", "NONE", true)
          }
        }
        stage ("Scan helm test image for fixable vulns") {
          steps {
            scanAndReport("conjur-k8s-cluster-test:dev", "HIGH", false)
          }
        }
        stage ("Scan helm test image for total vulns") {
          steps {
            scanAndReport("conjur-k8s-cluster-test:dev", "NONE", true)
          }
        }
      }
    }

    stage('E2E Workflow Tests') {
      stages {
        stage('Update Helm dependencies') {
          steps {
            sh './bin/helm-dependency-update-in-docker'
          }
        }
        stage('Test app with') {
          parallel {
            stage('Enterprise in GKE') {
              steps {
                sh 'cd bin/test-workflow && summon --environment gke ./start --enterprise --platform gke --ci-apps'
              }
            }
            stage('OSS in OpenShift v(current)') {
              steps {
                sh 'cd bin/test-workflow && summon --environment openshift -D ENV=ci -D VER=current ./start --platform openshift --ci-apps'
              }
            }
            stage('OSS in OpenShift v(next)') {
              when {
                expression { params.TEST_OCP_NEXT }
              }
              steps {
                sh 'cd bin/test-workflow && summon --environment openshift -D ENV=ci -D VER=next ./start --platform openshift --ci-apps'
              }
            }
          }
        }
        stage('Enterprise in Jenkins') {
          stages {
            stage('Test app in GKE') {
              steps {
                sh '''
                  HOST_IP="$(curl http://169.254.169.254/latest/meta-data/public-ipv4)";
                  echo "HOST_IP=${HOST_IP}"
                  cd bin/test-workflow && summon --environment gke ./start --enterprise --platform jenkins --ci-apps
                '''
              }
            }
            stage('Test app in OpenShift v(current)') {
              steps {
                sh '''
                  HOST_IP="$(curl http://169.254.169.254/latest/meta-data/public-ipv4)";
                  echo "HOST_IP=${HOST_IP}"
                  cd bin/test-workflow && summon --environment openshift -D ENV=ci -D VER=current ./start --enterprise --platform jenkins --ci-apps
                '''
              }
            }
            stage('Test app in OpenShift v(next)') {
              when {
                expression { params.TEST_OCP_NEXT }
              }
              steps {
                sh '''
                  HOST_IP="$(curl http://169.254.169.254/latest/meta-data/public-ipv4)";
                  echo "HOST_IP=${HOST_IP}"
                  cd bin/test-workflow && summon --environment openshift -D ENV=ci -D VER=next ./start --enterprise --platform jenkins  --ci-apps
                '''
              }
            }
          }
        }
      }
    }
    */

    stage('Release') {
      when {
        expression {
          MODE == "RELEASE"
        }
      }

      steps {
        release { billOfMaterialsDirectory, assetDirectory, toolsDirectory ->
          // Publish release artifacts to all the appropriate locations
          // Copy any artifacts to assetDirectory to attach them to the Github release
          
          // Create Go application SBOM
          sh """go-bom --tools "${toolsDirectory}" --image "golang:1.17" --main "cmd/authenticator/" --output "${billOfMaterialsDirectory}/go-app-bom.json" """
          // Create Go module SBOM
          sh """go-bom --tools "${toolsDirectory}" --image "golang:1.17" --output "${billOfMaterialsDirectory}/go-mod-bom.json" """
          // Publish edge release
          sh 'summon ./bin/publish'
        }
      }
    }
  }

  post {
    always {
      cleanupAndNotify(currentBuild.currentResult)
    }
  }
}
