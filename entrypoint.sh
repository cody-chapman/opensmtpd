#!/bin/bash

set -euo pipefail
ln -snf /usr/share/zoneinfo/$TZ /etc/localtime
echo $TZ > /etc/timezone

# creating logging files before user
touch /var/log/opensmtpd/opensmtpd.log
touch /var/log/opensmtpd/opensmtpd.err.log

# Start cron in the background and tail logs to keep container alive
printenv > /etc/environment # Ensure cron sees env variables
echo "Launching supervisor..."
/usr/bin/supervisord -c /etc/supervisor/supervisord.conf
