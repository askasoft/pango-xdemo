set BASEDIR=%~dp0

pushd %BASEDIR%

SET EXE=xdemodb.exe

call ..\build.bat
move /Y %EXE% ..\

SET EXE=
