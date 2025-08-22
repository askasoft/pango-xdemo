REM go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest

SET GOARCH=amd64
SET GOOS=windows
SET GO111MODULE=on

SET COMPANY=Askasoft LLC.
SET PRODUCT=Pango Xdemo
SET VER_MAJOR=1
SET VER_MINOR=2
SET VER_PATCH=0

FOR /F "tokens=* USEBACKQ" %%i IN (`powershell -Command "Get-Date -date (Get-Date).ToUniversalTime()-uformat %%Y-%%m-%%dT%%H:%%M:%%SZ"`) DO (
	SET BUILD_TIME=%%i
)

SET YEAR=%BUILD_TIME:~0,4%
SET VERSION=%VER_MAJOR%.%VER_MINOR%.%VER_PATCH%
FOR /F "tokens=* USEBACKQ" %%i IN (`git rev-parse --short HEAD`) DO (
	SET REVISION=%%i
)

SET /A VER_BUILD=0x%REVISION%
SET VER_BUILD=%VER_BUILD:~0,4%

call :build .     xdemo    web/favicon.ico
call :build .\cmd xdemodb  ../web/favicon.ico
move .\cmd\*.exe  .\

go test ./...

EXIT /B %ERRORLEVEL%


:build
pushd %1
SET EXE=%2.exe
SET ICON=%3
(
echo {
echo 	"FixedFileInfo": {
echo 		"FileVersion": {
echo 			"Major": %VER_MAJOR%,
echo 			"Minor": %VER_MINOR%,
echo 			"Patch": %VER_PATCH%,
echo 			"Build": %VER_BUILD%
echo 		},
echo 		"FileFlagsMask": "3f",
echo 		"FileFlags ": "00",
echo 		"FileOS": "040004",
echo 		"FileType": "01",
echo 		"FileSubType": "00"
echo 	},
echo 	"StringFileInfo": {
echo 		"Comments": "",
echo 		"CompanyName": "%COMPANY%",
echo 		"FileDescription": "%PRODUCT% %VERSION%.%REVISION%",
echo 		"FileVersion": "",
echo 		"InternalName": "",
echo 		"LegalCopyright": "Copyright (c) %YEAR% %COMPANY%, All Rights Reserved",
echo 		"LegalTrademarks": "",
echo 		"OriginalFilename": "%EXE%",
echo 		"PrivateBuild": "",
echo 		"ProductName": "%PRODUCT%",
echo 		"ProductVersion": "%VERSION%.%REVISION%",
echo 		"SpecialBuild": ""
echo 	},
echo 	"VarFileInfo": {
echo 		"Translation": {
echo 			"LangID": "0409",
echo 			"CharsetID": "04B0"
echo 		}
echo 	},
echo 	"IconPath": "%ICON%",
echo 	"ManifestPath": ""
echo }
) > versioninfo.json

go generate

SET PKG=github.com/askasoft/pangox/xwa
SET LDF=-X %PKG%.Version=%VERSION% -X %PKG%.Revision=%REVISION% -X %PKG%.Buildtime=%BUILD_TIME%

go build -ldflags "%LDF%" -o %EXE%
popd
:EXIT /B 0
