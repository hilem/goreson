# Development Running on OSX:

0. Ensure boot2docker is running
  1. ./boot2docker init
  2. ./boot2docker start
1. Ensure rethinkdb is running
  1. ./rethinkdb --bind all
2. Build & Run Go app Binary (run following from <APP_ROOT>)
  1. ./go build
  2. ./gadder

# Generate Docs: request templates zip, then run...

- godoc -templates="<TEMPLATES_DIR>" -html=true . > doc.html

# Deploy Build

- ./docker build -t hilem/gadder .
- ./docker push hilem/gadder
###### Login to  web server
- ssh -A gadder

## Quick Notes for Recent Deploy
docker run -d --name web --link db:db -p 80:3000 hilem/gadder
docker run -d --name db -p 8080:8080 -p 28015:28015 -p 29015:29015 dockerfile/rethinkdb rethinkdb --bind all


###### Start/Stop web service
- fleetctl stop gadder@1.service
- fleetctl start gadder@1.service
###### May need to login to dockerhub on the server for the preceeding to work properly
- docker login
  - follow on screen instructions to input (username/password/email)
###### Confirm Service Restarted
- fleetctl status gadder@1
###### Repeat previous steps for remaining web servers

# Reading Logs on Prod

###### Read Entire Journal
- journalctl
###### Read Specific Service Logs
- journalctl -u gadder@1.service
