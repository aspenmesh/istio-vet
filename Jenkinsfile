node('docker') {
  timestamps {
    properties([disableConcurrentBuilds()])
    def img

    docker.withRegistry('https://quay.io', 'quay-infrajenkins-robot-creds') {
      stage('Build') {
        checkout scm

        img = docker.build("quay.io/aspenmesh/istio-vet:${env.BRANCH_NAME}-${env.BUILD_ID}")

      }

      stage('Push') {
        /* Push the container to the custom Registry */
        img.push()
        /* Tag image as latest for the branch */
        img.push("${env.BRANCH_NAME}")
        sha = sh(script: 'git rev-parse --short HEAD', returnStdout: true).trim()
        img.push("${env.BRANCH_NAME}-${sha}")
      }
    }
  }
}
