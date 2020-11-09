To run that sweet docker container,

docker build -t <DOCKER USERNAME>/election-frontend .  

docker run --name election-frontend -p 4680:3000 -d <DOCKER USERNAME>/election-frontend