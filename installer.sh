#!/bin/bash
set -e
# check if script is running as root
if [[ $EUID -ne 0 ]]; then
    echo "Error: Please run as root." >&2
    exit 1
fi
go mod init upload
go mod tidy
go build -o /usr/bin/upload .
# copy config to /etc
cp upload.conf upload.conf.temp
confs=("endpoint" "path" "tlsport" "port" "cert" "key")
for c in "${confs[@]}"; do
    echo "enter ${c}:"
    read -r var
    echo "parsed: ${var} for ${c}"
    sed -i "s|^${c}=.*|${c}=${var}|" upload.conf.temp
done
cp upload.conf.temp /etc/upload.conf
rm upload.conf.temp
# set config file permissions
chmod 0644 /etc/upload.conf
# install systemD unit file
cp upload.service /etc/systemd/system/multi-user.target.wants/upload.service
# set unit file permissions
chmod 0644 /etc/systemd/system/multi-user.target.wants/upload.service
# enable and start service
systemctl enable upload.service
systemctl start upload.service
systemctl status upload.service
# cleanup
cd / || exit 1 && echo "failed to change path to root"
rm -rf /tmp/codegoy
echo "upload program is now installed"
