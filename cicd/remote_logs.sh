# If current machine is srv560368 then run docker compose logs app --tail=50 other run ssh in.msheriff.com "docker compose logs app --tail=50"
# run it from the context of current directory in local and remote directory is /root/github.com/openagent
if [ "$(hostname)" = "srv560368" ]; then
    docker compose logs app --tail=50
else
    ssh in.msheriff.com "cd /root/github.com/openagent && docker compose logs app --tail=50"
fi
