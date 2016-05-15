## Setting up Dev Environment

1. Use the docker-compose file in this directory to start a development container
2. Connect to it `docker exec -it cadvisorbuild /bin/bash`
3. `cd /opt/share`
4. `bash devsetup.sh` - this will install all dependencies and download the cAdvisor code

## Building
1. `cd /opt/share/src/github.com/google/cadvisor/deploy`
2. `bash build_wf.sh`

## Deploying to Docker Hub
1. From deploy directory: `docker login hub.docker.com/r/wavefronthq/cadvisor`
2. `docker push wavefronthq/cadvisor:latest`

## Running inside container as a standalone binary
1. `cd /opt/share/src/github.com/google/cadvisor`
2. Change "storage_driver_host" in the following command to your proxy URL: `./cadvisor -storage_driver=wavefront -storage_driver_host=172.17.0.3:2878 -storage_driver_db=$(hostname) --logtostderr=true`

## Running inside container as a container
```
docker run \
-e WF_INTERVAL=10 \
-e WF_ADD_TAGS="az=\"us-west-2\" app=\"cadvisortesting\"" \
--volume=/:/rootfs:ro \
--volume=/var/run:/var/run:rw \
--volume=/sys:/sys:ro \
--volume=/var/lib/docker/:/var/lib/docker:ro \
--publish=8081:8080 \
--detach=false \
--name=cadvisor \
--volume=/cgroup:/cgroup:ro \
wavefronthq/cadvisor --storage_driver=wavefront -storage_driver_host=172.17.0.3:2878 -storage_driver_db=$(hostname) --logtostderr=true
'''

2. See it:
http://192.168.99.100:8081/
http://192.168.99.100:8081/api/v1.3/docker
