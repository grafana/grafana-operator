pipeline {
    agent {
        node {
            label 'docker'
        }
    }
    stages {
        stage('build') {
            steps {
                sh "make submodule"
                sh "make docker-build IMG=us.gcr.io/${PROJECT}/grafana-operator:4.0.0"
            }
        }
        stage('push') {
            when {
                branch 'master'
            }
            steps {
                sh "docker push us.gcr.io/${PROJECT}/grafana-operator:4.0.0
            }
        }
    }
}
