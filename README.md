# Gemigit

A self-hosted gemini Git service

## Features

* Allow users to create and manage git repositories
* Private and public repositories
* Serving git repositories on the http and ssh protocols
* LDAP authentication
* 2FA with time-based one-time passwords
* Option to use token authentication when doing git operations
* Basic brute-force protection
* User groups
* Privilege system for read/write access
* Support for sqlite and mysql databases
* Support stateless mode for multiple instances and load balancing

## Setup

* Build the program using the command "go build"
* Copy config.yaml into either /etc/gemigit, /usr/local/etc/gemigit or the working directory where Gemigit will be executed
* Edit the configuration file to suit your needs
* Execute gemigit

## Demo

You can try a public instance of Gemigit at this address gemini://gemini.rmf-dev.com using a Gemini client or with a [gemini web proxy][0]

## Contact

For inquiries about this software or the instance running at gemini://gemini.rmf-dev.com, you can contact the main maintainer of this project at : rawmonk@firemail.cc

[0]: https://portal.mozz.us/gemini/gemini.rmf-dev.com/
