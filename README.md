# Xdemo

## Linux

### install golang (1.21)
```sh
sudo yum -y install golang
```

or

```sh
wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
tar -xzvf go1.21.6.linux-amd64.tar.gz
sudo mv go /opt/
sudo ln -s /opt/go/bin/go /usr/bin/go
```

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

### bind privileged port
```sh
sudo setcap 'cap_net_bind_service=+ep' /app/xdemo/xdemo
```


## Windows

### install golang (1.21)
download https://go.dev/dl/go1.21.6.windows-amd64.zip and extract it.


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

### create mysql database (not supported yet)
```sql
CREATE USER 'xdemo'@'%' IDENTIFIED BY 'xdemo';
CREATE DATABASE xdemo CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
GRANT ALL PRIVILEGES ON xdemo.* TO 'xdemo'@'%';
```

### create postgresql database
```sql
CREATE USER xdemo PASSWORD 'xdemo';
CREATE DATABASE xdemo WITH OWNER=xdemo ENCODING='UTF-8';
GRANT ALL ON DATABASE xdemo TO xdemo;
```

### init database
```sh
cd /app/xdemo
./xdemo execsql conf/schema.sql
```

### install as windows service
Run As Administrator

```bat
xdemo.exe install
```


## apache proxy setting

### Virtual Host
```xml
<VirtualHost *:80 *:443>
	ServerName xdemo.local

	<If "%{HTTPS} == 'on'">
		RequestHeader set X-Forwarded-Proto "https"
		RequestHeader set X-Forwarded-Port  "443"
	</If>

	DocumentRoot /app/xdemo/web
	<Directory /app/xdemo/web>
		AllowOverride All
		Options FollowSymLinks Indexes
		Require all granted
	</Directory>

	AllowEncodedSlashes NoDecode

	ProxyTimeout      300
	ProxyRequests     Off
	ProxyPreserveHost On

	ProxyPass         /         http://localhost:6060/ nocanon retry=0
	ProxyPassReverse  /         http://localhost:6060/ nocanon
</VirtualHost>
```

### Directory
```ini
[server]
prefix = /xdemo
```

```xml
	Alias /xdemo /app/xdemo/web
	<Directory /app/xdemo/web>
		AllowOverride All
		Options FollowSymLinks Indexes
		Require all granted
	</Directory>

	ProxyTimeout      300
	ProxyRequests     Off
	ProxyPreserveHost On

	ProxyPass         /xdemo   http://localhost:6060/xdemo  nocanon retry=0
	ProxyPassReverse  /xdemo   http://localhost:6060/xdemo  nocanon
```


## nginx proxy setting
```xml
server {
	listen       80;
	listen       443 ssl;
	server_name  xdemo.local;

	charset utf-8;

	client_max_body_size 0;

	location / {
		proxy_pass              http://localhost:6060;
		proxy_http_version      1.1;
		proxy_set_header        X-Real-IP $remote_addr;
		proxy_set_header        X-Forwarded-Proto $scheme;
		proxy_set_header        X-Forwarded-Port  $server_port;
		proxy_set_header        X-Forwarded-For   $proxy_add_x_forwarded_for;
		proxy_set_header        Host $http_host;
		proxy_request_buffering off;
		proxy_buffering         off;
		proxy_connect_timeout   10;
		proxy_send_timeout      10;
		proxy_read_timeout      600;
	}
}
```


## SSL
```sh
openssl genrsa -out xdemo.key 2048
openssl req -new -x509 -sha256 -key xdemo.key -out xdemo.cer -days 3650

openssl req -x509 -newkey rsa:2048 -keyout xdemo.key -out xdemo.cer -days 3650 -nodes -subj "/CN=*.xdemo.local"
```


## OpenSearch

```
DELETE xdemo_applog

PUT xdemo_applog
{
	"mappings": {
		"properties": {
			"time": {
				"type": "date",
				"format": "date_time"
			}
		}
	}
}

GET xdemo_applog

GET xdemo_applog/_search
{
	"query": {
		"match_all": {}
	}
}

POST xdemo_applog/_delete_by_query
{
	"query": {
		"match_all": {}
	}
}
```


```
DELETE xdemo_access

PUT xdemo_access
{
	"mappings": {
		"properties": {
			"time": {
				"type": "date",
				"format": "date_time"
			}
		}
	}
}

GET xdemo_access

GET xdemo_access/_search
{
	"query": {
		"match_all": {}
	}
}

POST xdemo_access/_delete_by_query
{
	"query": {
		"match_all": {}
	}
}
```

