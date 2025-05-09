@echo off
set BUILD_DIR=build
set EXE_NAME=app.exe

if not exist %BUILD_DIR% mkdir %BUILD_DIR%

go build -o %BUILD_DIR%/%EXE_NAME% .
%BUILD_DIR%/%EXE_NAME%
