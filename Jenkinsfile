pipeline {
  agent {
    docker {
      image 'golang:1.10.0'
    }
  }
  stages {
    stage('Build') {
      steps {
        sh 'ln -s `pwd` /go/src/octoprint-upload && go get octoprint-upload'
        sh 'GOOS=windows GOARCH=amd64 go build -v -o opu.exe octoprint-upload'
        sh 'GOOS=darwin GOARCH=amd64 go build -v -o opu-darwin octoprint-upload'
        sh 'GOOS=linux GOARCH=amd64 go build -v -o opu-linux octoprint-upload'
        archiveArtifacts 'opu*'
      }
    }
  }
}
