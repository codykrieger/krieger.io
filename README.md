# krieger.io

This repository contains the source for my personal website, http://krieger.io.
It is meant to be deployed to Heroku.

There are a couple of interesting things at play here:

- The Jekyll source; pretty standard, aside from a couple of custom plugins
- The source for a tiny server written in Go, which actually handles all of the
  requests to the Heroku dyno(s) (`main.go`)
- The assumption that the Heroku buildpack being used is https://github.com/ddollar/heroku-buildpack-multi
  (which then invokes the buildpacks in `.buildpacks`)
  - Buildpack number one is https://github.com/codykrieger/heroku-buildpack-jekyll,
    which builds the Jekyll site after being pushed to Heroku
  - Buildpack number two is https://github.com/kr/heroku-buildpack-go,
    which compiles and runs the server in `main.go`

This allows for an entirely `git push`-based workflow; no pre-compilation of
the Jekyll site or Go server necessary. Everything's done on-demand.

## Setup

```
% git clone https://github.com/codykrieger/krieger.io.git
% cd krieger.io

% heroku create -b https://github.com/ddollar/heroku-buildpack-multi.git
Creating damp-waters-7941... done, stack is cedar
BUILDPACK_URL=https://github.com/ddollar/heroku-buildpack-multi.git
http://damp-waters-7941.herokuapp.com/ | git@heroku.com:damp-waters-7941.git
Git remote heroku added

% git push heroku master
Initializing repository, done.
Counting objects: 247, done.
Delta compression using up to 8 threads.
Compressing objects: 100% (223/223), done.
Writing objects: 100% (247/247), 61.11 KiB | 0 bytes/s, done.
Total 247 (delta 93), reused 0 (delta 0)

-----> Fetching custom git buildpack... done
-----> Multipack app detected
=====> Downloading Buildpack: https://github.com/codykrieger/heroku-buildpack-jekyll.git
=====> Detected Framework: Ruby
-----> Compiling Ruby/Rack
-----> Using Ruby version: ruby-2.1.2
-----> Installing dependencies using 1.5.2
       Running: bundle install --without development:test --path vendor/bundle --binstubs vendor/bundle/bin -j4 --deployment
       ...
       Bundle completed (50.08s)
       Cleaning up the bundler cache.
-----> Generating Jekyll site
       Configuration file: /tmp/build_dbb79348-bf19-4930-8202-ba1094ce97fc/_config.yml
       Source: /tmp/build_dbb79348-bf19-4930-8202-ba1094ce97fc
       Destination: /tmp/build_dbb79348-bf19-4930-8202-ba1094ce97fc/_site
       Generating... Asset Pipeline: Processing 'css_asset_tag' manifest 'global'
       Asset Pipeline: Saved 'global-12661c839e4d11f81d0f851d43774147.css' to '/tmp/build_dbb79348-bf19-4930-8202-ba1094ce97fc/_site/assets'
       done.
=====> Downloading Buildpack: https://github.com/kr/heroku-buildpack-go.git
=====> Detected Framework: Go
-----> Installing go1.2.2... done
-----> Running: godep go install -tags heroku ./...
Using release configuration from last framework (Go).
-----> Discovering process types
       Procfile declares types -> web

-----> Compressing... done, 32.8MB
-----> Launching... done, v4
       http://damp-waters-7941.herokuapp.com/ deployed to Heroku

To git@heroku.com:damp-waters-7941.git
 * [new branch]      master -> master
```
