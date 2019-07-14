#!/usr/bin/env bash

docker build -t doughyou/food_svc:latest .

docker login -u $docker_user --password $docker_password

docker push doughyou/food_svc:latest