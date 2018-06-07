# backhub
 
[![Build Status](https://travis-ci.org/jakewarren/backhub.svg?branch=master)](https://travis-ci.org/jakewarren/backhub/)
[![GitHub release](http://img.shields.io/github/release/jakewarren/backhub.svg?style=flat-square)](https://github.com/jakewarren/backhub/releases])
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](https://github.com/jakewarren/backhub/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/jakewarren/backhub)](https://goreportcard.com/report/github.com/jakewarren/backhub)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=shields)](http://makeapullrequest.com)

> Back up starred repos & Gists from GitHub

Backs up:
* Personal repos
* Starred repos
* Personal Gists
* Starred Gists

## Install
### Option 1: Binary

Download the latest release from [https://github.com/jakewarren/backhub/releases/latest](https://github.com/jakewarren/backhub/releases/latest)

### Option 2: From source

```
go get github.com/jakewarren/backhub
```
## Usage

When you first run `backhub` the program will prompt you to create a token. The program will then create four directories in the current working directory and clone/pull all personal & starred repos, and personal & starred Gists.

## Acknowledgements

https://github.com/cinience/GitStalker  
https://github.com/aprescott/gist-backup  
https://github.com/iamsalnikov/gorespect  

## Changes

All notable changes to this project will be documented in the [changelog].

The format is based on [Keep a Changelog](http://keepachangelog.com/) and this project adheres to [Semantic Versioning](http://semver.org/).

## License

MIT © 2018 Jake Warren

[changelog]: https://github.com/jakewarren/backhub/blob/master/CHANGELOG.md
