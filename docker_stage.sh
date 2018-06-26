#!/bin/bash
docker build . -t "acrQat.azurecr.io/sysz:v0.1"
docker push acrQat.azurecr.io/sysz:v0.1