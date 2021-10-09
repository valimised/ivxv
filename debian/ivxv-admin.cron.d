# IVXV Internet voting framework

# Crontab for Management Service
# /etc/cron.d/ivxv-admin

# Copy service log files to Log Monitor with 5 min interval
*/5 * * * *     ivxv-admin      if [ -x /usr/bin/ivxv-copy-log-to-logmon ]; then /usr/bin/ivxv-copy-log-to-logmon --log-level=WARNING; fi

# Download voter list changesets from VIS with 15 min interval
*/15 * * * *    ivxv-admin      if [ -x /usr/bin/ivxv-voter-list-download ]; then /usr/bin/ivxv-voter-list-download --log-level=WARNING; fi

# Import common voting stats from voting service and export it to VIS
5,20,35,50 * * * *    ivxv-admin      if [ -x /usr/bin/ivxv-voterstats ]; then /usr/bin/ivxv-voterstats common --log-level=WARNING; fi
