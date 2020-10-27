# NetStat

## Backend
### Dependencies
- docker
- docker-compose

### Run
    git clone https://github.com/tylerferrara/NetStat
    cd ./NetStat/backend
    # build the containers
    docker-compose build
    # start
    docker-compose --env-file .env up
    # stop
    docker-compose down

