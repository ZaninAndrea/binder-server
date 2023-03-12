ssh -t server 'echo "== Connecting to remote server ==" \
    && cd ~/server \
    && echo "== Fetching latest version from git ==" \
    && git pull \
    && echo "== Building the application ==" \
    && go build -o build/server ./cmd \
    && echo "== Restarting systemctl service ==" \
    && sudo systemctl restart server.service \
    && echo "== Deployed successfully =="'