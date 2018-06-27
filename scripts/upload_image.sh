#!/bin/bash
# Script constants
SERVICE_NAME="sysz"

# Global variables
FINAL_IMAGE=""

docker_build()
{
    REGISTRY="acrQat.azurecr.io"
    IMAGE_TAG="sysz-testing"
    DOCKER_FILE_DIR="."
    read -p 'CONTAINER REGISTRY (default:acrQat.azurecr.io): ' REGISTRY
    read -p 'IMAGE TAG (default:sysz-testing)': IMAGE_TAG
    if [ "$REGISTRY" = "" ]; then
        REGISTRY="acrQat.azurecr.io"
    fi
    if [ "$IMAGE_TAG" = "" ]; then
        IMAGE_TAG="sysz-testing"
    fi
    if [ "$DOCKER_FILE_DIR" = "" ]; then
        DOCKER_FILE_DIR="."
    fi
    IMAGE_NAME="$REGISTRY/$SERVICE_NAME:$IMAGE_TAG"
    echo "Building image $IMAGE_NAME"
    docker build -t $IMAGE_NAME $DOCKER_FILE_DIR
    FINAL_IMAGE=$IMAGE_NAME
}

docker_push()
{
    if [ "$FINAL_IMAGE" = "" ]; then
        exit
    else
        echo "Pushing image $FINAL_IMAGE"
    fi
    docker push $FINAL_IMAGE
}


# Run the functions in this order
docker_build
docker_push
