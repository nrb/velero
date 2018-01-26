#!/usr/bin/env groovy
pipeline {
    agent any

    stages {
        stage("Test") {
            steps {
                sh 'make test'
            }
        }
        stage("Build executables") {
            steps {
                sh 'make all'
            }
        }
        stage("Build container") {
            when {
                // Only do release actions on branchs that explicitly label themselves as release branches
                branch.contains('release-')
            }
            steps {
                sh 'make container'
            }
        }
    }
}
