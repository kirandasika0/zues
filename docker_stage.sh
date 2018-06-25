#!/bin/bash
docker build . -t "acrQat.azurecr.io/sys:v0.1"
docker push acrQat.azurecr.io/sys:v0.1