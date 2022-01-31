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

  parameters { 
    booleanParam(
      name: 'TEST_OCP_NEXT',
      description: 'Whether or not to run the pipeline against the next OCP version',
      defaultValue: false) 
  }

  stages {

    stage('Build client Docker image') {
      steps {
        sh './bin/build'
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
            stage('Enterprise in GKE with JWT authentication') {
              steps {
                sh 'cd bin/test-workflow && summon --environment gke ./start --enterprise --platform gke --jwt'
              }
            }
            stage('OSS in OpenShift v(current) with JWT authentication') {
              steps {
                sh 'cd bin/test-workflow && summon --environment openshift -D ENV=ci -D VER=current ./start --platform openshift --jwt'
              }
            }
            stage('OSS in OpenShift v(next) with JWT authentication') {
              when {
                expression { params.TEST_OCP_NEXT }
              }
              steps {
                sh 'cd bin/test-workflow && summon --environment openshift -D ENV=ci -D VER=next ./start --platform openshift --jwt'
              }
            }
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
