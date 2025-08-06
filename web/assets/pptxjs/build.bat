@echo off

REM npm install -g uglify-js clean-css-cli

set BASEDIR=%~dp0
set JSDIR=%BASEDIR%\js

cd /d %JSDIR%\
call :minjs pptxjs


echo --------------------------------------
echo DONE.

cd /d %BASEDIR%
exit /b


:minjs
echo --------------------------------------
echo --  minify js: %1
call uglifyjs.cmd %1.js --warn --compress --mangle --source-map url=%1.min.js.map -o %1.min.js
exit /b

:mincss
echo --------------------------------------
echo --  minify css: %1
call cleancss.cmd -d -o %1.min.css %1.css
exit /b
