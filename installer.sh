#!/bin/bash
# check if script is running as root
if [[ $EUID -ne 0 ]]; then
    echo "Error: Please run as root." >&2
    exit 1
fi
# check if required commands are installed
commands=("git" "go")
for cmd in "${commands[@]}"; do
    if command -v "$cmd" >/dev/null 2>&1; then
        echo "$cmd is installed at: $(command -v "$cmd")"
    else
        echo "$cmd is NOT installed."
    fi
done
# make temp path
mkdir /tmp/codegoy
# enter temp path
cd /tmp/codegoy || echo "failed to enter temp path" && exit 1
# clone repo
git clone https://github.com/CodeGoy/upload.git --depth 1
# enter repo path
cd upload || echo "failed to enter git path" && exit 1
# build
go mod init upload
go mod tidy
go build -o /usr/bin/upload .
# copy config to /etc
cp upload.conf upload.conf.tmp
confs=("endpoint" "path" "tlsport" "port" "cert" "key")
for c in "${confs[@]}"; do
    echo "enter ${c}:"
    read -r var
    echo "parsed: ${var} for ${c}"
    sed -i "s|^${c}=.*|${c}=${var}|" upload.conf.tmp
done
cp upload.conf.temp /etc/upload.conf
rm upload.conf.temp
# set config file permissions
chmod 0644 /etc/upload.conf
# install systemD unit file
cp upload.service /etc/systemd/system/
# set unit file permissions
chmod 0644 /etc/systemd/system/upload.service
# enable and start service
systemctl enable upload.service
systemctl start upload.service
systemctl status upload.service
# cleanup
cd / || exit 1 && echo "failed to change path to root"
rm -rf /tmp/codegoy
#
echo "upload program is now installed"
