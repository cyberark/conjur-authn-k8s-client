#!/usr/bin/env groovy

pipeline {
  agent { label 'executor-v2' }

  options {
    timestamps()
    buildDiscarder(logRotator(numToKeepStr: '30'))
  }

  stages {
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
      success {
          script {
               if (env.BRANCH_NAME == 'master') {
                      build (job:'../cyberark--secrets-provider-for-k8s/master', wait: false)
               }
          }
      }
      always {
        cleanupAndNotify(currentBuild.currentResult)
      }
  }
}
