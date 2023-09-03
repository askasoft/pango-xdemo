# XDemo

## Linux

### install golang (1.20)
> sudo yum -y install golang

or

> wget https://go.dev/dl/go1.20.6.linux-amd64.tar.gz
> tar -xzvf go1.20.6.linux-amd64.tar.gz
> sudo mv go /opt/


### build
```sh
./build.sh
```

### deploy
```sh
export LOG_SLACK_WEBHOOK=https://hooks.slack.com/services/...

./deploy.sh
```

### install as system service
```sh
sudo useradd xdemo

echo '
[Unit]
Description=Pango Xdemo
After=syslog.target network.target local-fs.target

[Service]
Type=simple
WorkingDirectory=/app/xdemo
ExecStart=/app/xdemo/xdemo
ExecReload=/bin/kill -HUP $MAINPID
Restart=on-failure
User=xdemo
Group=xdemo

[Install]
WantedBy=multi-user.target
' | sudo tee /usr/lib/systemd/system/xdemo.service

sudo systemctl daemon-reload
sudo systemctl enable xdemo
sudo systemctl start xdemo
```


## Windows

### install golang (1.20)
download https://go.dev/dl/go1.20.6.windows-amd64.zip and extract it.


### get goversioninfo
```bat
go get github.com/josephspurrier/goversioninfo/cmd/goversioninfo
```

### build
```bat
build.bat
```

### deploy
```bat
SET LOG_SLACK_WEBHOOK=https://hooks.slack.com/services/...

deploy.bat
```

### init mysql database
```sql
CREATE USER 'xdemo'@'%' IDENTIFIED BY 'xdemo';
CREATE DATABASE xdemo CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
GRANT ALL PRIVILEGES ON xdemo.* TO 'xdemo'@'%';
```

### init mysql database
```sql
CREATE USER xdemo PASSWORD 'xdemo';
CREATE DATABASE xdemo WITH OWNER=xdemo ENCODING='UTF-8';
GRANT ALL ON DATABASE xdemo TO xdemo;
```


### install as windows service
Run As Administrator

```bat
xdemo.exe install
```


## apache proxy setting

```xml
<VirtualHost *:80>
	ServerName xdemo.local

	DocumentRoot /app/xdemo/web
	<Directory /app/xdemo/web>
		AllowOverride All
		Options FollowSymLinks Indexes
		Require all granted
	</Directory>

	ProxyTimeout 300
	ProxyRequests Off
	ProxyPreserveHost On

	ProxyPass         /         http://localhost:6060/ nocanon retry=0
	ProxyPassReverse  /         http://localhost:6060/ nocanon
</VirtualHost>
```

