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

    stage('Run Kubernetes Conjur Demo') {
      steps {
        script {
          build (job:'../kubernetes-conjur-demo/split-local-authenticator-and-secretless', wait: true)
        }
      }
    }

    // Cannot scan dev image as it's based on busybox and trivy can't determine
    // the OS
    stage("Scan redhat image") {
      steps {
        scanAndReport("conjur-authn-k8s-client-redhat:dev", "NONE")
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
