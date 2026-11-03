@echo off
rem Starts Word Killer in a window big enough for every screen (82 x 32).
rem Windows Terminal ignores console resize requests, so it is asked for the
rem size at launch instead; the classic console is resized in place.

setlocal
set COLS=90
set ROWS=34
set GAME=%~dp0word-killer.exe

rem The game reads config.json and data\ from its own folder
pushd "%~dp0"

rem Already inside Windows Terminal: open a new window of the right size
if defined WT_SESSION goto terminal

rem Started by double-click: prefer Windows Terminal when it is installed
where wt >nul 2>nul
if %errorlevel%==0 goto terminal

rem Classic console: resize this window, then run in it. If resizing fails
rem the game still starts and asks for a bigger window itself.
mode con: cols=%COLS% lines=%ROWS% >nul 2>nul
"%GAME%"
popd
goto :eof

:terminal
wt --size %COLS%,%ROWS% -d "%CD%" "%GAME%"
popd
