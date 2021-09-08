#!/usr/bin/env groovy

pipeline {
  agent { label 'executor-v2' }

  options {
    timestamps()
    buildDiscarder(logRotator(numToKeepStr: '30'))
  }
  
  triggers {
    cron(getDailyCronString())
  }

  stages {
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

    stage('Build client Docker image') {
      steps {
        sh './bin/build'
      }
    }

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
          /*
           * Helm dependency update is done before running E2E tests in parallel
           * since this is not currently thread-safe (Helm chart downloads use
           * a non-uniquely named 'tmpcharts' directory and fail if the directory
           * already exists).
           */
          steps {
            sh './bin/helm-dependency-update-in-docker'
          }
        }
        stage('Test app with') {
          parallel {
            stage('Enterprise in GKE') {
              steps {
                sh 'cd bin/test-workflow && summon --environment gke ./start --enterprise --platform gke'
              }
            }
            stage('OSS in OpenShift') {
              steps {
                sh 'cd bin/test-workflow && summon --environment openshift -D ENV=ci -D VER=current ./start --platform openshift'
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
                  cd bin/test-workflow && summon --environment gke ./start --enterprise --platform jenkins
                '''
              }
            }
            stage('Test app in OpenShift') {
              steps {
                sh '''
                  HOST_IP="$(curl http://169.254.169.254/latest/meta-data/public-ipv4)";
                  echo "HOST_IP=${HOST_IP}"
                  cd bin/test-workflow && summon --environment openshift -D ENV=ci -D VER=current ./start --enterprise --platform jenkins
                '''
              }
            }
          }
        }
      }
    }

    stage('Publish client Docker images') {
      parallel {
        stage('On a master build') {
          when { branch 'master' }

          steps {
            sh 'summon ./bin/publish --edge'
          }
        }
        stage('On a new tag') {
          when { tag "v*" }

          steps {
            sh 'summon ./bin/publish --latest'
          }
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
