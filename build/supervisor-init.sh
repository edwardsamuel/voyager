#!/usr/bin/env bash
cat <<EOF > /etc/supervisor/conf.d/app.conf
[program:app]
command=$*
autorestart=true
autostart=true
stderr_logfile_maxbytes=20MB   ; max # logfile bytes b4 rotation (default 50MB)
stderr_logfile_backups=5       ; # of stderr logfile backups (default 10)
stdout_logfile_maxbytes=20MB   ; max # logfile bytes b4 rotation (default 50MB)
stdout_logfile_backups=5       ; # of stdout logfile backups (default 10)
EOF
exec supervisord -n -c /etc/supervisor/supervisord.conf
