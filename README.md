
# Slack Reservations Command

A simple way to manage reservations on shared resources


# Quick Start

    RESOURCES="comma, separated, list, of, resources" SLACK_VERIFICATION_TOKEN="xxxxxx" ./slack-reservations-command

# Setup

You need a server to run this command from.

Requires Go v1.7 or greater

Create a Go workspace if you haven't already

    cd /some/workspace/directory
    mkdir -p src/github.com/abhchand
    mkdir -p pkg
    mkdir -p bin

Set your `$GOPATH`. Best to add this to your `.bash_profile` or similar startup configuration

    export GOPATH="/some/workspace/directory"


Clone the repository

    cd $GOPATH/src/github.com/abhchand
    git clone git@github.com:abhchand/slack-reservations-command.git

Build the project

    go build

Follow [Slack's instructions](https://api.slack.com/apps) for setting up a new Slack App with a Slash Command. It will provide you a Verification token that's required below.

You'll fill out the following:

    Command: /reservations
    Request URL: http://your.host.here:8080/slack/commands/reservations
    Description: Manage reservations
    Usage Hint: help | list | reserve [resource] for [duration] | extend [resource] by [duration] | cancel [resource]


Run the app

    RESOURCES="comma, separated, list, of, resources" SLACK_VERIFICATION_TOKEN="xxxxxx" ./slack-reservations-command

