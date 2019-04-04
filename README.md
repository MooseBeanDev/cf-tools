## Table of Contents
- [Overview](#overview)
- [Installation and Usage](#installation-and-usage)
- [Planned Features](#planned-features)

## Overview
This project will query the cloud foundry api using the cf go libraries, copy the results to a local json cache, and allow you to pull useful information from it locally at a much faster pace. CF CLI does not give much visibility outside the targeted org and space so this application aims to increase global visiblity for platform operators.

## Installation and Usage
You will need a new-ish version of go.

Build and move to somewhere in your PATH
```
go build -o cf-tools main.go

cp cf-tools /usr/local/bin/
chmod 755 /usr/local/bin/cf-tools
chown root:root /usr/local/bin/cf-tools
```

Set Env Variables and pull the api data down to local json files. located in /home/USER/.cfcache/
```
export CF_API_ADDRESS=https://api.system.your-url.org
export CF_USERNAME=admin
export CF_PASSWORD=BLAHBLAHBLAH

cf-tools sync
```

Show all crashed/unhealthy apps, as well as app total, crashed app total, etc.
```
cf-tools app health-check
```

List services as in the command "cf marketplace"
```
cf-tools service list
```

Show all instances of a service's usage, in this example mysql
```
cf-tools service usage mysql-dev-shared
```

Get a service instance's guid by entering its name. here you can see this returns multiple results.
```
cf-tools service get-guid credential-db                                                                                         


Searching for service guid by service instance name:  credential-db

Org:  test
Space:  Development
Service Name:  test-db
Service Guid:  00ea075e-1a57-40f4-844d-a3fd5e35cb44

Org:  test
Space:  QA
Service Name:  test-db
Service Guid:  13076ff1-c357-464b-b7af-69ddc7da444d
```

Find all apps this service is bound to. this searches by service guid. you can also search by app guid.
```
cf-tools binding service 00ea075e-1a57-40f4-844d-a3fd5e35cb44                                                                   


Searching for binding by service instance guid:  00ea075e-1a57-40f4-844d-a3fd5e35cb44

Org:  test
Space:  Development
App Name:  test-service
App Guid:  8739d7c9-07f1-4a06-b8d1-8b6b4511da19
```

Find app guids by searching app name
```
cf-tools app get-guid spring-music
```

Search by app guid to show all services bound to it
```
cf-tools binding app BLAH_GUID_HERE

#example
cf-tools binding app d63d1c3f-d631-4027-80f9-ffc7186f260d
                                                                        
Searching for bindings by app guid:  d63d1c3f-d631-4027-80f9-ffc7186f260d
                                       
Org:  test                                                        
Space:  postgres-broker-test                      
Service Name:  my-postgresqltest                                
Service Guid:  5006a480-9fbf-4930-925e-67934850d641
```

Show help
```
cf-tools -h
cf-tools --help
```

You can probably add the executable to the appropriate go bin path and have it as a globally executable file.

This is still very much under development.

## Planned Features
- Org tree (show me all apps under a certain org, in all of its spaces)
- Service tree (show me all services under a certain org, in all of its spaces)
- App Info (search by name or guid to get app info)
- Service Instance Info (search by name or guid to get service info)
