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

    // Pull existing images from internal registry in order to promote
    sh "docker pull registry.tld/conjur-authn-k8s-client:${sourceVersion}"
    sh "docker pull registry.tld/conjur-authn-k8s-client-redhat:${sourceVersion}"
    // Promote source version to target version.
    sh "summon ./bin/publish --promote --source ${sourceVersion} --target ${targetVersion}"
  }
  return
}

pipeline {
  agent { label 'azure-linux' }

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

        stage('Log messages') {
          steps {
            validateLogMessages()
          }
        }

        stage('Cluster-Prep Schema') {
          steps { sh './bin/validate-schema ./helm/conjur-config-cluster-prep/values.schema.json'}
        }

        stage('Application Namespace-Prep Schema') {
          steps { sh './bin/validate-schema ./helm/conjur-config-namespace-prep/values.schema.json'}
        }
      }
    }

    // Generates a VERSION file based on the current build number and latest version in CHANGELOG.md
    stage('Validate Changelog and set version') {
      steps {
        updateVersion("CHANGELOG.md", "${BUILD_NUMBER}")
      }
    }
    
    stage('Get Upstream Dependencies') {
      steps {
        sh './bin/update-go-mod --go-mod ./go.mod'
      }
    }

    stage('Get latest upstream dependencies') {
       steps {
         updateGoDependencies('${WORKSPACE}/go.mod')
       }
     }

    stage('Build client Docker image') {
      steps {
        sh './bin/build'
      }
    }

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
          
          // Create Go application SBOM using the go.mod version for the golang container image
          sh """go-bom --tools "${toolsDirectory}" --go-mod ./go.mod --image "golang" --main "cmd/authenticator/" --output "${billOfMaterialsDirectory}/go-app-bom.json" """
          // Create Go module SBOM
          sh """go-bom --tools "${toolsDirectory}" --go-mod ./go.mod --image "golang" --output "${billOfMaterialsDirectory}/go-mod-bom.json" """
          cat '${billOfMaterialsDirectory}/go-app-bom.json'
          cat '${billOfMaterialsDirectory}/go-mod-bom.json'
          sh 'exit 1'
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
