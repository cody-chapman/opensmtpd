#!/bin/bash

# set -e

set -euo pipefail
ln -snf /usr/share/zoneinfo/$TZ /etc/localtime
echo $TZ > /etc/timezone


declare -p | grep -Ev 'BASHOPTS|BASH_VERSINFO|EUID|PPID|SHELLOPTS|UID' > /etc/container.env
chmod 0644 /etc/container.env

# Start cron in the background and tail logs to keep container alive
printenv > /etc/environment # Ensure cron sees env variables
echo "Launching supervisor..."
/usr/bin/supervisord -c /etc/supervisor/supervisord.conf
