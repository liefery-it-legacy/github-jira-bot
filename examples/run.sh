#!/bin/bash
# <PROJECT ROOT>/examples/run.sh
make build || exit 1
./build/github-jira-bot run -f examples/config/config.yaml -e examples/config/issue-comment.env --verbose
#./build/github-jira-bot run -f config/samples/config.yaml -e config/samples/issue-comment.env --verbose