#!/bin/bash

# Used in Docker build to set platform dependent variables
case $TARGETARCH in
    "amd64")
	echo "x86_64-unknown-linux-gnu" > /.platform
	;;
    "arm64") 
	echo "aarch64-unknown-linux-gnu" > /.platform
	;;
esac
# copy from https://blog.container-solutions.com/building-multiplatform-container-images