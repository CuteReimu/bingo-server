@echo off
for /f %%I in ('dir *.proto /A-D /B /ON') do (
  echo generating %%I...
  protoc --proto_path=. --go_out=. %%I
)
