#!/bin/sh

# Wrapper: start the real docker starter asynchronously so Android's boot
# process won't be blocked. The real starter is `start-dockerd.sh` which
# includes a wait-for-path logic (waits up to 60s for $DOCKER_ROOT).

REAL_START="/data/local/docker/start-dockerd.sh"
LOGFILE="/data/local/docker/start-dockerd.log"

if [ -f "$REAL_START" ]; then
	# Start in background and redirect output to a log file.
	sh "$REAL_START" >"$LOGFILE" 2>&1 &
else
	echo "Warning: $REAL_START not found, nothing started" >&2
fi

exit 0
