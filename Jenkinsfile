pipeline {

    agent any

    tools {
        go 'Go'
    }
    environment {
        GO114MODULE = 'on'
        CGO_ENABLED = 0
        GOPATH = "${JENKINS_HOME}/jobs/${JOB_NAME}/builds/${BUILD_ID}"
        DB_PASSWD = 'Home@302017'
        DB_ADDR = '3.229.43.168'
        DB_PORT = 5432
        DB_NAME = 'postgres'
        DB_USER = 'postgres'
    }
    stages {
//          stage("Git Clone"){
//              steps{
//
//                 git credentialsId: 'GIT_HUB_CREDS', url: 'https://github.com/jitin03/qastack-workflow.git'
//              }
//          }
        stage('Pre Test') {
            steps {
                echo 'Installing dependencies'
                sh 'go version'
                sh 'go get -u golang.org/x/lint/golint'
            }
        }

        stage('Build') {
            steps {
                echo 'Compiling and building'
                sh 'go build'
            }
        }
        stage("Docker build"){
            steps{
            sh 'docker version'
            sh 'docker build -t stack-workflow .'
            sh 'docker image list'
            sh 'docker tag stack-workflow mehuljitin/stack-workflow:stack-workflow'
            sh 'docker run -d -e DB_USER=$DB_USER -e DB_PASSWD=$DB_PASSWD -e DB_ADDR=$DB_ADDR -e DB_NAME=$DB_NAME -p 8094:8094 stack-workflow'
            }
        }



    stage("Push Image to Docker Hub"){
        steps{
            script{
                 withCredentials([string(credentialsId: 'DOCKER_HUB_CREDS', variable: 'PASSWORD')]) {
        sh 'docker login -u mehuljitin -p $PASSWORD'
                 }
            sh 'docker push  mehuljitin/stack-workflow:stack-workflow'
            }
        }
    }


    }
    // post {
    //     always {
    //         emailext body: "${currentBuild.currentResult}: Job ${env.JOB_NAME} build ${env.BUILD_NUMBER}\n More info at: ${env.BUILD_URL}",
    //             recipientProviders: [[$class: 'DevelopersRecipientProvider'], [$class: 'RequesterRecipientProvider']],
    //             to: "${params.RECIPIENTS}",
    //             subject: "Jenkins Build ${currentBuild.currentResult}: Job ${env.JOB_NAME}"

    //     }
    // }
}