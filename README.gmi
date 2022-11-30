# Gemigit

A self-hosted gemini Git service

## Features

* Allow users to create and manage git repositories
* Private and public repositories
* Serving git repositories on the http protocol
* LDAP authentication
* Basic bruteforce protection
* User groups
* Privilege system for read/write access

## Setup

* Build the program using the command "go build"
* Copy config.yaml into either /etc/gemigit, /usr/local/etc/gemigit or the working directory where Gemigit will be executed
* Edit the config file to suit your needs
* Execute gemigit
