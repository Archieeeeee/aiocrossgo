package media

import (
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/widget"
	jsoniter "github.com/json-iterator/go"
	"image/color"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

var wd string
var mediaFile string
var ffmpegPath string
var ffmpegTextfield *widget.Entry
var chooseFileTextfield *widget.Entry
var outMediaTextfield *widget.Entry
var outMediaPath string
var msgLabel *widget.Label
var config AACFixConfig

func Init() {
	wwd, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	wd = wwd

}

type AACFixConfig struct {
	FfmpegPath string
}

func ReloadCfg() {
	cfgPath := GetFilePathInWd("/cfg/cfg.json.txt")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		config = *new(AACFixConfig)
		config.FfmpegPath = GetFilePathInWd("/ffmpeg/bin/ffmpeg.exe")
		return
	}
	bs, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		log.Println(err)
	}
	log.Println("ReloadLocalCfg res=" + string(bs))
	jsoniter.Unmarshal(bs, &config)
}

func SaveLocalCfg() {
	cfgPath := GetFilePathInWd("/cfg/cfg.json.txt")
	bs, err := jsoniter.Marshal(config)
	if err != nil {
		log.Println(err)
	}
	err = ioutil.WriteFile(cfgPath, bs, 0777)
	if err != nil {
		log.Println(err)
	}
}

func appendSpaceBox(box *widget.Box) {
	spaceBox := canvas.NewRectangle(color.Transparent)
	spaceBox.SetMinSize(fyne.NewSize(200, 30))
	box.Append(spaceBox)
}

func GetFilePath(s string) string {
	return strings.Replace(s, "file://", "", -1)
}

func GetFilePathInWd(elem ...string) string {
	return filepath.FromSlash(path.Join(wd, path.Join(elem...)))
}

func fileChooseCallback(rc fyne.URIReadCloser, err error) {

	mediaFile = rc.URI().String()
	chooseFileTextfield.Text = mediaFile
}

func StartUI() {
	os.Setenv("FYNE_FONT",  GetFilePathInWd("/libs/SourceHanSansCN-Normal.ttf"))
	os.Setenv("FYNE_SCALE", "1.2")
	a := app.New()
	w := a.NewWindow("视频AAC修复工具")

	serversBox := widget.NewVBox()

	ffmpegTextfield = widget.NewEntry()
	ffmpegTextfield.Text = config.FfmpegPath
	chooseFfmpegFile := widget.NewButton("ffmpeg位置", func() {
		mediaChooser := dialog.NewFileOpen(func(readCloser fyne.URIReadCloser, err error) {
			ffmpegPath = GetFilePath(readCloser.URI().String())
			ffmpegTextfield.Text = ffmpegPath
			ffmpegTextfield.Refresh()

			config.FfmpegPath = ffmpegPath
			SaveLocalCfg()
		}, w)
		mediaChooser.Show()
	})
	serversBox.Append(chooseFfmpegFile)
	serversBox.Append(ffmpegTextfield)
	appendSpaceBox(serversBox)


	//chooseFileLabel := widget.NewLabel("选择文件")
	chooseFileTextfield = widget.NewEntry()
	chooseFileTextfield.Text = ""
	chooseFile := widget.NewButton("选择视频文件", func() {
		mediaChooser := dialog.NewFileOpen(func(readCloser fyne.URIReadCloser, err error) {
			mediaFile = GetFilePath(readCloser.URI().String())
			chooseFileTextfield.Text = mediaFile
			chooseFileTextfield.Refresh()

			outMediaPath = getOutMediaPath(mediaFile)
			outMediaTextfield.Text = outMediaPath
			outMediaTextfield.Refresh()
		}, w)
		mediaChooser.Show()
	})
	//serversBox.Append(chooseFileLabel)
	serversBox.Append(chooseFile)
	serversBox.Append(chooseFileTextfield)
	appendSpaceBox(serversBox)

	outMediaTextfield = widget.NewEntry()
	outMediaTextfield.Text = ""
	chooseOutMediaFile := widget.NewButton("设置输出文件", func() {
		mediaChooser := dialog.NewFileSave(func(readCloser fyne.URIWriteCloser, err error) {
			outMediaPath = GetFilePath(readCloser.URI().String())
			outMediaTextfield.Text = outMediaPath
			outMediaTextfield.Refresh()
		}, w)
		mediaChooser.Show()
	})
	serversBox.Append(chooseOutMediaFile)
	serversBox.Append(outMediaTextfield)

	fixAACButton := widget.NewButton("开始修复AAC", func() {
		FixAAC()
	})
	serversBox.Append(fixAACButton)

	msgLabel = widget.NewLabel("")
	serversBox.Append(msgLabel)
	appendSpaceBox(serversBox)

	w.SetContent(serversBox)



	w.Resize(fyne.NewSize(1024, 400))
	w.CenterOnScreen()
	w.ShowAndRun()
	defer a.Quit()
}

func getOutMediaPath(s string) string {
	dir := path.Dir(s)
	ext := path.Ext(s)
	fileNameNoExt := strings.TrimSuffix(s, ext)
	return path.Join(dir, fileNameNoExt + "-aac" + ext)
}

func FixAAC() {
	fmt.Println("AAC fix start")
	ffmpegPath = ffmpegTextfield.Text
	config.FfmpegPath = ffmpegPath
	SaveLocalCfg()
	mediaFile = chooseFileTextfield.Text
	outMediaPath = outMediaTextfield.Text
	//ffmpegPath = "E:\\toolsdev\\ffmpeg-4.1.1-win64-static\\bin\\ffmpeg.exe"
	//mediaFile = "C:\\Users\\longk\\Downloads\\dcsgo.mp4"
	//outMediaPath = "C:\\Users\\longk\\Downloads\\dcsgoaac.mp4"
	var cmds []string
	cmds = append(cmds, "-i", mediaFile)
	cmds = append(cmds, strings.Split("-c:v copy -c:a copy -bsf:a aac_adtstoasc -y", " ")...)
	cmds = append(cmds, outMediaPath)
	//cmd := exec.Command(ffmpegPath, strings.Split("-c:v copy -c:a copy -bsf:a aac_adtstoasc", " ")...)
	cmd := exec.Command(ffmpegPath, cmds...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	slurp, _ := ioutil.ReadAll(stderr)
	fmt.Printf("%s\n", slurp)

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}

	msgLabel.SetText(msgLabel.Text + "\n处理完成: " + outMediaPath)
	fmt.Println("AAC fix done")

	OpenExplorer(outMediaPath)
}


func OpenExplorer(filePath string) {
	cmd := exec.Command(`explorer`, `/select,`, outMediaPath)
	cmd.Run()
}
