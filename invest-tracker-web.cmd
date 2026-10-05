@echo off
rem Arranca a interface web de Control de Investimentos.
rem Executa desde o cartafol do programa para usar sempre a mesma base de datos.
cd /d "%~dp0"
"%~dp0invest-tracker.exe" web %*
if errorlevel 1 pause
