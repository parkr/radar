workflow "Build on push" {
  on = "push"
  resolves = ["docker build"]
}

action "docker build" {
  uses = "actions/docker/cli@master"
  args = ["build", "."]
}

workflow "Publish on push to master" {
  on = "push"
  resolves = ["Build & publish"]
}

action "On branch master" {
  uses = "actions/bin/filter@25b7b846d5027eac3315b50a8055ea675e2abd89"
  args = "branch master"
}

action "Login to Docker Registry" {
  uses = "actions/docker/login@fe7ed3ce992160973df86480b83a2f8ed581cd50"
  needs = ["On branch master"]
  secrets = ["DOCKER_USERNAME", "DOCKER_PASSWORD"]
}

action "Build & publish" {
  uses = "actions/docker/cli-multi@fe7ed3ce992160973df86480b83a2f8ed581cd50"
  args = ["build -t parkr/radar:$GITHUB_SHA .", "push parkr/radar:$GITHUB_SHA"]
  needs = ["Login to Docker Registry"]
}
