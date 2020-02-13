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
        sh './bin/publish'
      }
    }
    stage('Validate') {
      parallel {
        stage('Changelog') {
          steps { sh './bin/parse-changelog.sh' }
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
