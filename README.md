# NetStat

### Install
- [docker](https://docs.docker.com/engine/install/)
- [docker-compose](https://docs.docker.com/compose/install/)

### Deployment

Includes a Backend Service, Frontend Web Server and SQL Database
```sh
git clone https://github.com/tylerferrara/NetStat
cd ./NetStat/backend
# build containers
docker-compose build
# start
docker-compose --env-file .env up
# stop
docker-compose down
```
