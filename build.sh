#!/usr/bin/env bash

docker build -t doughyou/food_svc:latest .

docker push doughyou/food_svc:latest