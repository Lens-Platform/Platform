#!/usr/bin/env bash
echo ">>>>>>>>>>>>>>>>>LENS PLATFORM<<<<<<<<<<<<<<<<<<<<<"
echo "Stopping All docker Containers"
docker stop $(docker ps -a -q)
echo ">>>>>>>>>>>>>>>>>LENS PLATFORM<<<<<<<<<<<<<<<<<<<<<"
echo "Removing All Stopped docker Containers"
docker rm $(docker ps -a -q)
echo "Removing Volumes"
docker volume rm $(docker volume ls -qf dangling=true)
docker volume ls -qf dangling=true | xargs -r docker volume rm
echo "Removing Networks"
docker network ls
docker network ls | grep "bridge"
docker network ls | awk '$3 == "bridge" && $2 != "bridge" { print $1 }'
echo "Removing Images"
docker images
docker rmi $(docker images --filter "dangling=true" -q --no-trunc)
docker images | grep "none"
docker rmi $(docker images | grep "none" | awk '/ / { print $3 }')
echo "Removing Containers"
docker ps
docker ps -a
docker rm $(docker ps -qa --no-trunc --filter "status=exited")
kill -9 $(lsof -t -i:9810)
echo ">>>>>>>>>>>>>>>>>LENS PLATFORM<<<<<<<<<<<<<<<<<<<<<"

