#!/bin/bash

set -o errexit

main() {
  setup_dependencies

  echo "INFO:
  Done! Finished setting up travis machine.
  "
}

setup_dependencies() {
  echo "INFO:
  Setting up dependencies.
  "

  sudo apt update -y
  sudo apt install ipvsadm python -y
  sudo apt install --only-upgrade docker-ce -y

  sudo pip install ansible

  docker info
  ansible --version
}

main

