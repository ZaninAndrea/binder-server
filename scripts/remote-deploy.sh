ssh -t raspberrypi.baida.dev 'echo "== Connecting to remote server ==" \
    && cd ~/apis/binder-server \
    && echo "== Fetching latest version from git ==" \
    && git pull \
    && echo "== Building the application ==" \
    && go build -o build/server ./cmd \
    && echo "== Restarting systemctl service ==" \
    && sudo systemctl restart binder-server.service \
    && echo "== Deployed successfully =="'