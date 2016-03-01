#/bin/bash

gox -os="darwin linux windows" -arch="386 amd64" -output="binaries/{{.Dir}}_{{.OS}}_{{.Arch}}"

cd binaries
echo compressing...
tar cvzf dockward_linux_386.tar.gz dockward_linux_386
tar cvzf dockward_linux_amd64.tar.gz dockward_linux_amd64
tar cvzf dockward_darwin_386.tar.gz dockward_darwin_386
tar cvzf dockward_darwin_amd64.tar.gz dockward_darwin_amd64
zip -r dockward_windows_386.zip dockward_windows_386.exe
zip -r dockward_windows_amd64.zip dockward_windows_amd64.exe

