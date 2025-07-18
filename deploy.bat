set APPNAME=xdemo
set APPHOME=C:\app\%APPNAME%
set PREFIX=
set LOG_WRITERS=stdout, textfile, jsonfile, xalfile, xajfile, httpdump

@REM set LOG_OPENSEARCH_APPLOG=https://localhost:9200/xdemo_applog/_bulk
@REM set LOG_OPENSEARCH_ACCESS=https://localhost:9200/xdemo_access/_bulk


if "%LOG_LEVEL%." == "." set LOG_LEVEL=INFO

mkdir %APPHOME%\conf
rmdir /S /Q %APPHOME%\tpls
rmdir /S /Q %APPHOME%\txts
rmdir /S /Q %APPHOME%\web

powershell -command "(gc conf\app.ini -Encoding utf8) | %% { $_ -replace 'prefix =.*', 'prefix = %PREFIX%' } | Out-File %APPHOME%\conf\app.ini -Encoding utf8"

powershell -command "(gc conf\log.ini -Encoding utf8).Replace('DEBUG', '%LOG_LEVEL%') | Out-File %APPHOME%\conf\log.ini -Encoding utf8"

copy /Y conf\config.csv %APPHOME%\conf\
copy /Y conf\*.sql      %APPHOME%\conf\
copy /Y conf\xdemo.*    %APPHOME%\conf\

if not "%LOG_SLACK_WEBHOOK%." == "." (
	set LOG_WRITERS=%LOG_WRITERS%, slack
	powershell -command "(gc %APPHOME%\conf\log.ini -Encoding utf8).Replace('LOG_SLACK_WEBHOOK', '%LOG_SLACK_WEBHOOK%') | %% { $_ -replace 'writer =.*', 'writer = %LOG_WRITERS%' } | Out-File %APPHOME%\conf\log.ini -Encoding utf8"
)
if not "%LOG_TEAMS_WEBHOOK%." == "." (
	set LOG_WRITERS=%LOG_WRITERS%, teams
	powershell -command "(gc %APPHOME%\conf\log.ini -Encoding utf8).Replace('LOG_TEAMS_WEBHOOK', '%LOG_TEAMS_WEBHOOK%') | %% { $_ -replace 'writer =.*', 'writer = %LOG_WRITERS%' } | Out-File %APPHOME%\conf\log.ini -Encoding utf8"
)
if not "%LOG_OPENSEARCH_APPLOG%." == "." (
	set LOG_WRITERS=%LOG_WRITERS%, appos
	powershell -command "(gc %APPHOME%\conf\log.ini -Encoding utf8).Replace('LOG_OPENSEARCH_APPLOG', '%LOG_OPENSEARCH_APPLOG%') | %% { $_ -replace 'writer =.*', 'writer = %LOG_WRITERS%' } | Out-File %APPHOME%\conf\log.ini -Encoding utf8"
	powershell -command "(gc %APPHOME%\conf\log.ini -Encoding utf8).Replace('LOG_OPENSEARCH_USERNAME', '%LOG_OPENSEARCH_USERNAME%') | %% { $_ -replace 'LOG_OPENSEARCH_PASSWORD', '%LOG_OPENSEARCH_PASSWORD%' } | Out-File %APPHOME%\conf\log.ini -Encoding utf8"
)
if not "%LOG_OPENSEARCH_ACCESS%." == "." (
	set LOG_WRITERS=%LOG_WRITERS%, xajos
	powershell -command "(gc %APPHOME%\conf\log.ini -Encoding utf8).Replace('LOG_OPENSEARCH_ACCESS', '%LOG_OPENSEARCH_ACCESS%') | %% { $_ -replace 'writer =.*', 'writer = %LOG_WRITERS%' } | Out-File %APPHOME%\conf\log.ini -Encoding utf8"
	powershell -command "(gc %APPHOME%\conf\log.ini -Encoding utf8).Replace('LOG_OPENSEARCH_USERNAME', '%LOG_OPENSEARCH_USERNAME%') | %% { $_ -replace 'LOG_OPENSEARCH_PASSWORD', '%LOG_OPENSEARCH_PASSWORD%' } | Out-File %APPHOME%\conf\log.ini -Encoding utf8"
)


copy  /Y %APPNAME%*.exe %APPHOME%\
@REM xcopy /Y /I /E tpls     %APPHOME%\tpls
@REM xcopy /Y /I /E txts     %APPHOME%\txts
@REM xcopy /Y /I /E web      %APPHOME%\web
