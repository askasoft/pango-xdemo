set APPNAME=xdemo
set APPHOME=C:\app\%APPNAME%
set PREFIX=

if "%LOG_LEVEL%." == "." set LOG_LEVEL=INFO

mkdir %APPHOME%\conf
rmdir /S /Q %APPHOME%\tpls
rmdir /S /Q %APPHOME%\txts
rmdir /S /Q %APPHOME%\web

powershell -command "(gc conf\app.ini -Encoding utf8) | %% { $_ -replace 'prefix =.*', 'prefix = %PREFIX%' } | Out-File %APPHOME%\conf\app.ini -Encoding utf8"

powershell -command "(gc conf\log.ini -Encoding utf8).Replace('DEBUG', '%LOG_LEVEL%').Replace('HOSTNAME', '%COMPUTERNAME%') | Out-File %APPHOME%\conf\log.ini -Encoding utf8"

if not "%LOG_SLACK_WEBHOOK%." == "." (
	set LOG_WRITERS=stdout, file, access, dump, slack
	powershell -command "(gc %APPHOME%\conf\log.ini -Encoding utf8).Replace('LOG_SLACK_WEBHOOK', '%LOG_SLACK_WEBHOOK%') | %% { $_ -replace 'writer =.*', 'writer = %LOG_WRITERS%' } | Out-File %APPHOME%\conf\log.ini -Encoding utf8"
)

copy  /Y %APPNAME%.exe  %APPHOME%\
xcopy /Y /I /E tpls     %APPHOME%\tpls
xcopy /Y /I /E txts     %APPHOME%\txts
xcopy /Y /I /E web      %APPHOME%\web
