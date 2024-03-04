def  GIT_BRANCH = 'main'
def  GIT_REPOSITORY_URL = 'https://github.com/gamepkw/google-oauth2-auth-service.git'
def  APP_NAME = 'go-app'
def  APP_VERSION = 'latest'
def  REMOTE_IMAGE_REPOSITORY = 'docker.io/gamepkw/oauth2-auth-service-image'

def  NAMESPACE_NAME = 'oauth2-auth-service-namespace'
def  DEPLOYMENT_NAME = 'oauth2-auth-service-deployment'
def  SERVICE_NAME = 'oauth2-auth-service-nodeport-service'

def  REMOTE_SERVER_HOST = '192.168.1.47'
def  REMOTE_SERVER_USER = 'pakawat'
def  REMOTE_SERVER_PASSWORD = '1234'

def remote = [:]
remote.name = 'server1'
remote.host = '192.168.1.47'
remote.user = 'pakawat'
remote.password = '1234'
remote.allowAnyHosts = true

pipeline {
    agent any

    tools {
        go 'go'
        dockerTool 'docker'
    }

    environment {
        KUBE_CONFIG = "/var/snap/microk8s/current/credentials/client.config"
    }

    stages {
        stage('Clean') {
            steps {
                script {
                    cleanWs()
                    sh "docker images --format '{{.Repository}}:{{.Tag}}' | grep '^${APP_NAME}' | xargs -I {} docker rmi -f {}"
                    sh "docker images --format '{{.Repository}}:{{.Tag}}' | grep '^${REMOTE_IMAGE_REPOSITORY}' | xargs -I {} docker rmi -f {}"
                    def unusedImages = sh(script: 'docker images | grep "<none>" | awk \'{print $3}\'', returnStdout: true).trim()
                    if (unusedImages) {
                        def imageIds = unusedImages.split()
                        imageIds.each { imageId ->
                            sh "docker rmi -f $imageId"
                        }
                    } else {
                        echo "No images with '<none>' tag found."
                    }
                }
            }
        }
        stage('Git Pull') {
            steps {
                script {
                    git branch: "${GIT_BRANCH}", url: "${GIT_REPOSITORY_URL}"
                    try {
                        appVersion = sh(returnStdout: true, script: 'git tag --contains | tail -1 | grep -E "^[0-9]+\\.[0-9]+\\.[0-9]+$"').trim()
                        if (appVersion) {
                            APP_VERSION = appVersion
                        }
                    } catch (Exception e) {
                        echo "No valid version tag found. Using default version."
                    }
                }
            }
        }
        stage('Install Dependencies') {
            steps {
                script {
                    sh "go mod tidy"
                }
            }
        }
        stage('Docker Authentication') {
            steps {
                script {
                    withCredentials([usernamePassword(credentialsId: 'docker-secret', passwordVariable: 'DOCKER_PASSWORD', usernameVariable: 'DOCKER_USERNAME')]) {
                        sh "docker login -u ${DOCKER_USERNAME} -p ${DOCKER_PASSWORD}"
                    }
                }
            }
        }
        stage('Build Image') {
            steps {
                script {
                    sh "docker build -t ${REMOTE_IMAGE_REPOSITORY}:${APP_VERSION} ."
                }
            }
        }
        stage('Push Image') {
            steps {
                script {
                    sh "docker tag ${LOCAL_IMAGE_REPOSITORY}:${APP_VERSION} ${REMOTE_IMAGE_REPOSITORY}:${APP_VERSION}"
                    def imageId = sh(script: "docker images -q ${REMOTE_IMAGE_REPOSITORY}:${APP_VERSION}", returnStdout: true).trim()
                    sh "docker push ${REMOTE_IMAGE_REPOSITORY}:${APP_VERSION}"
                }
            }
        }
        stage('Put manifest file onto remote server') {
            steps {
                script {
                    sshPut remote: remote, from: 'deployment/deployment.yaml', into: '.'
                    sshPut remote: remote, from: 'deployment/nodeport-service.yaml', into: '.'
                }
            }
        }
        stage('Deploy') {
            steps {
                script {
                    //sshCommand remote: remote, command: "kubectl get namespace oauth2-auth-service-namespace --kubeconfig=/var/snap/microk8s/current/credentials/client.config"
                    //kubectl create namespace oauth2-auth-service-namespace
                    
                    def applyDeploymentCommand = "kubectl apply -f deployment.yaml -n ${NAMESPACE_NAME} --kubeconfig=${KUBE_CONFIG}"
                    def applyNodePortServiceCommand = "kubectl apply -f nodeport-service.yaml -n ${NAMESPACE_NAME} --kubeconfig=${KUBE_CONFIG}"
                            
                    sshCommand remote: remote, command: "${applyDeploymentCommand}"
                    sshCommand remote: remote, command: "${applyNodePortServiceCommand}"
                }
            }
        }
    }
}