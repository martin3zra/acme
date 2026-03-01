# Acme

---

Create environment file, you can copy the `.env.sample` as `.env`.

# Deploy as service
## Test your binary
```shell
cd /Users/amartinez/Developer/acme
./build.sh
chmod +x bin/acme
./acme
```

## Create system service file
Create a file `/etc/systemd/system/acme-service.service`

```
[Unit]
Description=Acme Service
ConditionPathExists=/home/<your-user>/acme
After=network.target

[Service]
Type=simple
User=<your-user>
Group=<your-user>

WorkingDirectory=/home/<your-user>/acme
Environment="YOUR_ENV_1=YOUR_ENV_1_VALUE"
ExecStart=/home/<your-user>/acme/acme
Restart=on-failure
RestartSec=10

StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=acme-service

[Install]
WantedBy=multi-user.target
```

# Restart syslog:
```shell
systemctl restart rsyslog.service
```
# Active service
```
systemctl daemon-reload
service acme-service start
service acme-service status
```