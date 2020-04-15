#!/usr/bin/env groovy

pipeline {
  agent { label 'executor-v2' }

  options {
    timestamps()
    buildDiscarder(logRotator(numToKeepStr: '30'))
  }

  stages {
    stage('Validate') {
      parallel {
        stage('Changelog') {
          steps { sh './bin/parse-changelog.sh' }
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

    // Cannot scan dev image as it's based on busybox and trivy can't determine
    // the OS
    stage("Scan redhat image") {
      parallel {
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
      }
    }

    stage('Publish client Docker image') {
      when {
        branch 'master'
      }
      steps {
        sh 'summon ./bin/publish'
      }
    }
  }

  post {
    always {
      cleanupAndNotify(currentBuild.currentResult)
    }
  }
}
