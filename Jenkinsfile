pipeline {
    agent {
        label 'buster'
    }

    stages {
        stage('Build') {
            steps {
                sh 'go build'
            }
        }

        stage('Create feeds') {
            steps {
                sh './xpod sfutf'
                sh './xpod heavy-metal-sewing-circle'
                sh './xpod gothique-boutique'
            }
        }

        stage('Upload feeds') {
            steps {
                sshPublisher(publishers: [sshPublisherDesc(configName: 'proton-ieure-public',
                                                           transfers: [sshTransfer(sourceFiles: '*.xml')])])
            }
        }
    }
}
