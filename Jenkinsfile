pipeline {
  agent {
    label "github-jira-bot"
  }
  stages {
    stage('Run Github Jira Bot') {
      steps {
        container('github-jira-bot') {
          sh "github-jira-bot run --from-file /home/jenkins/.github-jira-bot/config.yaml --verbose"
        }
      }
    }
  }
}
