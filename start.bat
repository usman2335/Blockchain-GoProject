@echo off
REM Shell script to open 4 prompts and run a command with different parameters

REM Command to run
set COMMAND=go run .\test.go

REM Parameters for each window
set PARAM1=1
set PARAM2=2
set PARAM3=3
set PARAM4=4

REM Open new Command Prompt windows with different parameters
start cmd /k "%COMMAND% %PARAM1%"
start cmd /k "%COMMAND% %PARAM2%"
start cmd /k "%COMMAND% %PARAM3%"
start cmd /k "%COMMAND% %PARAM4%"

REM Optional: Keep the main script window open
pause
