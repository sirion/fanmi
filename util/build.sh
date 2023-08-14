#!/bin/bash

if [[ ! -d bin || ! -f app/main.go ]]; then 
    echo "Error: Please invoke build script from main repository directory" >&2
    exit 1
fi

mkdir -p bin && \
sudo rm -f bin/fanmi && \
go build -o bin/fanmi app/*.go && \
sudo chown root:root bin/fanmi && \
sudo chmod u+s bin/fanmi