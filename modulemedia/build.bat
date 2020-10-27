SET PARENT=F:\projects\projectsgo\src\AioCrossGo
SET BUILDP=F:\tmp\aacfix
cd /d F:\projects\projectsgo\src\AioCrossGo

cd modulemedia
go build
F:\projects\projectsgo\bin\fyne package -icon %PARENT%\modulemedia\aac.png
copy /Y %PARENT%\modulemedia\modulemedia.exe %BUILDP%\xiufu.exe


