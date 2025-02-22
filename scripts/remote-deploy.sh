#!/bin/bash
GOOS=linux GOARCH=arm GOARM=7 go build -o build/server ./cmd
scp build/server raspberrypi.baida.dev:~/apis/binder-server/build/new_server

# Connect to the Raspberry Pi and restart the service
ssh -t raspberrypi.baida.dev 'echo "== Connecting to remote server ==" \
    && cd ~/apis/binder-server \
    && echo "== Restarting systemctl service ==" \
    && sudo systemctl stop binder-server.service \
    && sudo mv build/new_server build/server \
    && sudo systemctl start binder-server.service \
    && echo "== Deployed successfully =="'