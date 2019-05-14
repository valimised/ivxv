# IVXV Internet voting framework

# Crontab for Management Service
# /etc/cron.d/ivxv-admin

# Copy service log files to Log Monitor with 5 min interval
*/5 * * * *     ivxv-admin      if [ -x /usr/bin/ivxv-copy-log-to-logmon ]; then /usr/bin/ivxv-copy-log-to-logmon; fi
