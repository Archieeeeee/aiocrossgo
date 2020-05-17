package data

import (
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/widget"
	"github.com/go-resty/resty/v2"
	"github.com/json-iterator/go"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//const API_BASE = "http://het.b.kda.io"
const API_BASE = "http://127.0.0.1"
var client = resty.New()
//const TROJAN_ROOT_PATH = "E:/tools/trojan"
//const TROJAN_CFG_ROOT_PATH = "E:/tools/trojan/aio"

const TROJAN_ROOT_PATH = "/trojan"
const TROJAN_CFG_ROOT_PATH = "/trojan/aio"

//const TROJAN_ROOT_PATH = "/trojan"
//const TROJAN_CFG_ROOT_PATH = "/trojan/aio"

var cmd *exec.Cmd
var cfg AioCrossConfig
var msgLabel *widget.Label
var portEntry *widget.Entry
var wd string
var localCfg CrossLocalConfig

type CrossLocalConfig struct {
	ClientPort string
}

type AioCrossConfig struct {
	TrojanServers []AioCrossTrojanServer `json:"trojanServers"`
	Stat string `json:"stat"`
}

type AioCrossTrojanServer struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	CertName   string `json:"certName"`
	KeyName    string `json:"keyName"`
	PortClient int `json:"portClient"`
	IsClient   bool `json:"isClient"`
	Name string `json:"name"`
	NameEn string `json:"nameEn"`
}

func Init() {
	wwd, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	wd = wwd

	ReloadLocalCfg()
}

func ReloadLocalCfg() {
	cfgPath := filepath.FromSlash(path.Join(wd, "/cfg.json.txt"))
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		log.Println("ReloadLocalCfg does not exist")
		localCfg = *new(CrossLocalConfig)
		localCfg.ClientPort = "1006"
		return
	}
	bs, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		log.Println(err)
	}
	log.Println("ReloadLocalCfg res=" + string(bs))
	jsoniter.Unmarshal(bs, &localCfg)
}

func SaveLocalCfg() {
	cfgPath := filepath.FromSlash(path.Join(wd, "/cfg.json.txt"))
	bs, err := jsoniter.Marshal(localCfg)
	if err != nil {
		log.Println(err)
	}
	err = ioutil.WriteFile(cfgPath, bs, 0777)
	if err != nil {
		log.Println(err)
	}
}

func GetConfig() {

	res, err := client.R().Get(API_BASE + "/aiocrosscfg")
	if err != nil {
		log.Println(err)
	}

	fmt.Printf("getConfig res=%s\n", res)


	err = jsoniter.Unmarshal(res.Body(), &cfg)
	if err != nil {
		log.Println(err)
	}
	fmt.Printf("cfg=%s\n", &cfg)



	//ConnectTrojan(cfg.TrojanServers[0])
}

func StartUI() {
	os.Setenv("FYNE_FONT", "d:/NotoSans-Regular.ttf");

	a := app.New()

	w := a.NewWindow("Cross程序")



	serversBox := widget.NewVBox()
	for idx, ss := range cfg.TrojanServers {
		log.Println("AAidx is ", idx)
		i := idx
		label := widget.NewButton(fmt.Sprintf("%s-%s", ss.NameEn, ss.Host), func() {
			log.Println("idx is ", i)
			ConnectTrojan(cfg.TrojanServers[i])
		})
		label.Resize(fyne.NewSize(800, 50))
		serversBox.Append(label)
	}

	//settings
	setForm := widget.NewForm()
	portEntry = widget.NewEntry()
	portEntry.SetPlaceHolder(localCfg.ClientPort)
	portEntry.SetText(localCfg.ClientPort)
	setForm.Append("Local Port", portEntry)
	serversBox.Append(setForm)

	//msg
	msgLabel = widget.NewLabel("")
	serversBox.Append(msgLabel)


	w.SetContent(serversBox)
	w.Resize(fyne.NewSize(800, 520))
	w.CenterOnScreen()


	//w.SetContent(widget.NewVBox(widget.NewLabel("hdalsdqw")))

	//go func() {
	//	time.Sleep(5 * time.Second)
	//	w.Hide()
	//}()

	w.ShowAndRun()

}

func ConnectTrojan(cfg AioCrossTrojanServer) {
	msgLabel.SetText("Connecting to " + cfg.NameEn + "...")
	cfg.IsClient = true

	//get trojan config
	p, err := jsoniter.Marshal(&cfg)
	if err != nil {
		log.Println(err)
	}
	ps := string(p)

	res, err := client.R().SetQueryParam("trojanConfig", ps).Get(API_BASE + "/aioBuildTrojanConfig")
	if err != nil {
		log.Println(err)
	}
	//fmt.Printf("ConnectTrojan res=%s\n", res)



	//trojanRootPath := wd + TROJAN_ROOT_PATH
	//trojanCfgPath := wd + TROJAN_CFG_ROOT_PATH
	//ccPath := trojanCfgPath + "/client.json"

	trojanRootPath := filepath.FromSlash(path.Join(wd, TROJAN_ROOT_PATH))
	trojanCfgPath := filepath.FromSlash(path.Join(wd, TROJAN_CFG_ROOT_PATH))
	ccPath := filepath.FromSlash(path.Join(trojanCfgPath, "/client.json"))
	exePath := filepath.FromSlash(path.Join(trojanRootPath, "/trojan.exe"))

	localCfg.ClientPort = portEntry.Text
	cfg.PortClient, err = strconv.Atoi(localCfg.ClientPort)
	cc := strings.ReplaceAll(string(res.Body()), "CLIENT_PORT", fmt.Sprintf("%d", cfg.PortClient))
	cc = strings.ReplaceAll(cc, "CERT_DIR", strings.ReplaceAll(trojanCfgPath, "\\", "/") )
	SaveLocalCfg()


	os.MkdirAll(trojanCfgPath, 0777)

	//write config
	fmt.Printf("ConnectTrojan clientJson=%s\n", cc)
	err = ioutil.WriteFile(ccPath, []byte(cc), 0777)
	if err != nil {
		log.Println(err)
	}
	//start trojan
	log.Printf("start trojan=%s", cfg.Name)
	StopCmd()
	cmd = exec.Command(exePath, "-c", ccPath)
	go RunCmd()
	go RunTestCmd()
	//msgLabel.SetText(fmt.Sprintf("Connected to %s %s, OK!", cfg.NameEn, cfg.Name))
	msgLabel.SetText(fmt.Sprintf("Connected to %s %s, OK!", cfg.NameEn, ""))
	//RunCmd()
}

func StopCmd() {
	if cmd != nil && cmd.Process != nil {
		log.Println("kill process")
		cmd.Process.Kill()
		cmd.Wait()
		time.Sleep(2 * time.Second)
		log.Println("kill process done")
	}
}

func RunCmd() {
	log.Println("start process")
	cmd.Run()
	defer StopCmd()
}

func RunTestCmd() {
	time.Sleep(2 * time.Second)
	//curl --socks5 127.0.0.1:1006 icanhazip.com
	exe := filepath.FromSlash(path.Join(wd, "/libs/curl.exe"))
	acmd := exec.Command(exe, "--socks5", "127.0.0.1:" + localCfg.ClientPort, "icanhazip.com")
	out, err := acmd.Output()
	if err != nil {
		log.Println(err)
	}
	acmd.Run()
	log.Printf("RunTestCmd res=%s", out)
	msgLabel.SetText(msgLabel.Text + "\nCurrent IP is: " + string(out))
}

