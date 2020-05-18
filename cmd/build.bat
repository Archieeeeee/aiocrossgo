SET PARENT=F:\projects\projectsgo\src\AioCrossGo
cd /d F:\projects\projectsgo\src\AioCrossGo

cd cmd
go build
F:\projects\projectsgo\bin\fyne package -icon D:\imgtemp\mouz.png

copy /Y %PARENT%\cfg.default.json.txt d:\cfg.json.txt
rm -f d:/cross.zip
E:/toolsdev/7-Zip/7z.exe a d:/cross.zip %PARENT%\libs %PARENT%\cmd\cmd.exe d:\cfg.json.txt
