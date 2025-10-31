#!/bin/sh

. /data/local/docker/docker.env

# If DISK_ROOT is empty, exit immediately.
if [ -z "$DISK_ROOT" ]; then
    echo "ERROR: DISK_ROOT is empty; exiting." >&2
    exit 1
fi

# Wait up to 60s for DOCKER_DATA_ROOT to appear
WAIT_TIMEOUT=60
WAIT_INTERVAL=2
ELAPSED=0

if [ ! -d "$DOCKER_DATA_ROOT" ]; then
    echo "Waiting for DOCKER_DATA_ROOT ($DOCKER_DATA_ROOT) to appear..."
    while [ ! -d "$DOCKER_DATA_ROOT" ] && [ "$ELAPSED" -lt "$WAIT_TIMEOUT" ]; do
        sleep "$WAIT_INTERVAL"
        ELAPSED=$((ELAPSED + WAIT_INTERVAL))
    done
fi

if [ ! -d "$DOCKER_DATA_ROOT" ]; then
    echo "ERROR: DOCKER_DATA_ROOT $DOCKER_DATA_ROOT not found after $WAIT_TIMEOUT seconds, exiting" >&2
    exit 1
fi

# --- original start.sh body ---

[ -d "$HOME" ] || mkdir "$HOME"
[ -d "$DOCKER_ROOT/var" ] || mkdir "$DOCKER_ROOT/var"

mountpoint -q "$DOCKER_ROOT/var" || mount -t tmpfs -o size=4M,uid=0,gid=0,mode=0755 tmpfs "$DOCKER_ROOT/var"

mkdir -p "$DOCKER_ROOT/var/run"

# IP forward and NAT
ip rule list prio 21000 from all iif docker0 | grep -Fwqs docker0 || ndc ipfwd add docker0 eth0
ndc ipfwd enable dockerd
iptables -S tetherctrl_FORWARD | grep -Fqe '-i docker0 -o eth0 -g' || ndc nat enable docker0 eth0 1

# Route
# ndc network create oem1 # oem2 , oem3 ... oem50
# ndc network interface add oem1 docker0
# ndc network route add oem1 docker0 172.17.0.0/16

# ip rule add from all to 172.16.0.0/12 table main
ip rule list prio 20500 from all table main | grep -Fwqs main || ip rule add prio 20500 from all table main


# Container interconnection
iptables -C oem_fwd -i docker0 -o docker0 -j ACCEPT 2>/dev/null || iptables -A oem_fwd -i docker0 -o docker0 -j ACCEPT


unshare -m "$DOCKER_ROOT/scripts/exec_dockerd.sh"