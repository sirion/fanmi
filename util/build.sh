#!/bin/bash

if [[ ! -d bin || ! -f app/main.go ]]; then 
    echo "Error: Please invoke build script from main repository directory" >&2
    exit 1
fi

if [[ "$1" == "release" ]]; then
    command -v upx > /dev/null
    if [[ "$?" != "0" ]]; then
        echo "Command for compressing executable files ("upx") not found. Used only in release build" >&2
        exit 2
    fi
fi

echo "Creating output directory 'bin' or removing old binary..."
mkdir -p bin && sudo rm -f bin/fanmi
if [[ "$?" != "0" ]]; then
    echo "Error: Could not create/clean 'bin' directory" >&2
    exit 3
fi

echo "Building executable..."
go build -ldflags "-s -w" -o bin/fanmi app/*.go
if [[ "$?" != "0" ]]; then
    echo "Error: Could not build executable file" >&2
    exit 4
fi

if [[ "$1" == "release" ]]; then
    echo "Compressing executable bin/fanmi..."
    # upx --brute --lzma bin/fanmi
    # upx --ultra-brute --lzma bin/fanmi
    upx -q --best --lzma bin/fanmi > /dev/null
    if [[ "$?" != "0" ]]; then
        echo "Error: Could not compress fanmi executable file" >&2
        exit 5
    fi
fi

echo "Changing ownership of bin/fanmi to root:root..."
sudo chown root:root bin/fanmi
if [[ "$?" != "0" ]]; then
    echo "Error: Could not give ownership to root user" >&2
    exit 6
fi

echo "Setting SUID on bin/fanmi..."
sudo chmod u+s bin/fanmi
if [[ "$?" != "0" ]]; then
    echo "Error: Could not set suid" >&2
    exit 7
fi

echo "Done."
