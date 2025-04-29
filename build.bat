@echo off
REM build.bat - Cross-compile Go static-file server for multiple OS/ARCH targets
REM Usage: build.bat

setlocal enabledelayedexpansion

REM Name of the output application (without extension)
set "APP_NAME=GinMiniServer"

REM List of OS/ARCH targets
set "TARGETS=windows/amd64 windows/386 windows/arm64 linux/amd64 linux/386 linux/arm linux/arm64 darwin/amd64 darwin/arm64"

REM Root output directory
set "OUTPUT_ROOT=build"

REM Clean previous builds
if exist "%OUTPUT_ROOT%" rd /s /q "%OUTPUT_ROOT%"
mkdir "%OUTPUT_ROOT%"

for %%T in (%TARGETS%) do (
    for /f "tokens=1,2 delims=/" %%A in ("%%T") do (
        set "GOOS=%%A"
        set "GOARCH=%%B"
        set "GOARM="

        REM Use GOARM=7 for ARM hard-float (armhf)
        if "%%A"=="linux" if "%%B"=="arm" set "GOARM=7"
        if "%%A"=="darwin" if "%%B"=="arm" set "GOARM=7"

        echo Building for !GOOS!-!GOARCH!!GOARM!

        REM Create output dir
        set "OUT_DIR=%OUTPUT_ROOT%\!GOOS!-!GOARCH!!GOARM!"
        if not exist "!OUT_DIR!" mkdir "!OUT_DIR!"

        REM Executable name
        set "OUT_NAME=%APP_NAME%"
        if "!GOOS!"=="windows" set "OUT_NAME=!OUT_NAME!.exe"

        REM Export env and build
        set GOOS=!GOOS!
        set GOARCH=!GOARCH!
        if defined GOARM set GOARM=!GOARM!
        go build -o "!OUT_DIR!\!OUT_NAME!" .
    )
)

echo All builds completed.
endlocal
