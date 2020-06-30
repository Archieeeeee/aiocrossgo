SET PARENT=F:\projects\projectsgo\src\AioCrossGo
SET BUILDP=D:\crossbuild
cd /d F:\projects\projectsgo\src\AioCrossGo

cd module
go build -o cross
F:\projects\projectsgo\bin\fyne package -icon D:\imgtemp\mouz.png

mkdir -p %BUILDP%\aio\
copy /Y %PARENT%\aio\* %BUILDP%\aio\
copy /Y %PARENT%\cfg.default.json.txt %BUILDP%\aio\cfg.json.txt
copy /Y %PARENT%\module\module.exe %BUILDP%\cross.exe
rm -f %BUILDP%\cross.zip
E:/toolsdev/7-Zip/7z.exe a %BUILDP%\cross.zip %PARENT%\libs %BUILDP%\cross.exe %BUILDP%\aio
