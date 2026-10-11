#!/bin/bash
set -e
# check if script is running as root
if [[ $EUID -ne 0 ]]; then
    echo "Error: Please run as root." >&2
    exit 1
fi

function buildProgram() {
    echo "build program"
    go build -o /usr/bin/upload .
}

function updateService() {
    systemctl stop upload.service
    git pull
    buildProgram
    systemctl start upload.service
}

function installService() {
    go mod init upload
    go mod tidy
    buildProgram
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
    cat >/etc/systemd/system/upload.service <<'EOF'
[Unit]
Description=Http based uploader
After=network.target
[Service]
Type=simple
ExecStart=/usr/bin/upload
Restart=on-failure
User=root
[Install]
WantedBy=multi-user.target
EOF
    # set unit file permissions
    chmod 0644 /etc/systemd/system/upload.service
    # enable and start service
    systemctl enable upload.service
    systemctl start upload.service
    systemctl status upload.service
    echo "upload program is now installed"
}

case "$1" in
    install)
        echo "installing service..."
        installService
        ;;
    update)
        echo "updating service..."
        updateService
        ;;
    *)
        echo "Invalid option: $1"
        echo "Valid options: install | update "
        exit 1
        ;;
esac