package data

import (
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/go-resty/resty/v2"
	"github.com/json-iterator/go"
	"github.com/mitchellh/go-ps"
	"image/color"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

//const API_BASE = "http://het.b.kda.io"
var API_BASE = "http://127.0.0.1"
var client = resty.New()
//const TROJAN_ROOT_PATH = "E:/tools/trojan"
//const TROJAN_CFG_ROOT_PATH = "E:/tools/trojan/aio"

const TROJAN_ROOT_PATH = "/libs"
const TROJAN_CFG_ROOT_PATH = "/aio"

//const TROJAN_ROOT_PATH = "/trojan"
//const TROJAN_CFG_ROOT_PATH = "/trojan/aio"

var proxyCmd *exec.Cmd
var cfg AioCrossConfig
var msgLabel *widget.Label
var portEntry *widget.Entry
var portHttpEntry *widget.Entry
var enableHttpEntry *widget.Check
var wd string
var localCfg CrossLocalConfig

type CrossLocalConfig struct {
	ClientPort string
	ApiBase string
	LastTrojanPid int
	UseTrojan bool
	ClientPortHttp string
	EnableSocksToHttp bool
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

	PortSs int    `json:"portSs"`
	Method string `json:"method"`
	PasswordSs string `json:"passwordSs"`
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
	cfgPath := GetFilePathInWd("/aio/cfg.json.txt")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		log.Println("ReloadLocalCfg does not exist")
		localCfg = *new(CrossLocalConfig)
		localCfg.ClientPort = "1006"
		localCfg.ApiBase = "http://cross.kda.io:8080"
		API_BASE = localCfg.ApiBase
		localCfg.UseTrojan = true
		localCfg.ClientPortHttp = "8888"
		localCfg.EnableSocksToHttp = true
		return
	}
	bs, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		log.Println(err)
	}
	log.Println("ReloadLocalCfg res=" + string(bs))
	jsoniter.Unmarshal(bs, &localCfg)
	API_BASE = localCfg.ApiBase
}

func SaveLocalCfg() {
	cfgPath := GetFilePathInWd("/aio/cfg.json.txt")
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
	//os.Setenv("FYNE_FONT", "d:/NotoSans-Regular.ttf")
	os.Setenv("FYNE_FONT",  GetFilePathInWd("/libs/SourceHanSansCN-Normal.ttf"))
	os.Setenv("FYNE_SCALE", "1.2")

	a := app.New()
	serversBox := widget.NewVBox()

	//a.Settings().SetTheme(theme.LightTheme())
	a.Settings().SetTheme(theme.DarkTheme())

	w := a.NewWindow("Cross程序")

	//settings
	setForm := widget.NewForm()
	portEntry = widget.NewEntry()
	portEntry.SetPlaceHolder(localCfg.ClientPort)
	portEntry.SetText(localCfg.ClientPort)

	portHttpEntry = widget.NewEntry()
	portHttpEntry.SetPlaceHolder(localCfg.ClientPortHttp)
	portHttpEntry.SetText(localCfg.ClientPortHttp)

	enableHttpEntry = widget.NewCheck("一般不需要开", switchHttpPortEntry)
	enableHttpEntry.SetChecked(localCfg.EnableSocksToHttp)
	switchHttpPortEntry(localCfg.EnableSocksToHttp)

	typeBox := widget.NewVBox()
	radio := widget.NewRadio([]string{"Trojan", "Shadowsocks"}, func(s string) {
		if s == "Trojan" {
			localCfg.UseTrojan = true
		} else {
			localCfg.UseTrojan = false
		}
	})
	radio.Horizontal = true
	if localCfg.UseTrojan {
		radio.SetSelected("Trojan")
	} else {
		radio.SetSelected("Shadowsocks")
	}
	typeBox.Append(radio)

	setForm.Append("本地Socks端口(范围1-65535)", portEntry)
	setForm.Append("连接类型", typeBox)
	setForm.Append("是否启用Http代理", enableHttpEntry)
	setForm.Append("本地Http代理端口(范围1-65535)", portHttpEntry)

	serversBox.Append(setForm)

	hr := widget.NewGroup("服务器列表")
	serversBox.Append(hr)

	//servers
	for idx, ss := range cfg.TrojanServers {
		log.Println("AAidx is ", idx)
		i := idx
		label := widget.NewButton(fmt.Sprintf("%s    %s", ss.Name, ss.Host), func() {
			log.Println("idx is ", i)
			server := cfg.TrojanServers[i]
			t := "Shadowsocks"
			if localCfg.UseTrojan {
				t = "Trojan"
			}
			msgLabel.SetText(fmt.Sprintf("正在连接到 %s 连接类型 %s...", server.NameEn, t))
			localCfg.ClientPort = portEntry.Text
			localCfg.ClientPortHttp = portHttpEntry.Text
			localCfg.EnableSocksToHttp = enableHttpEntry.Checked
			SaveLocalCfg()

			StopCmd()
			if localCfg.UseTrojan {
				ConnectTrojan(server)
			} else {
				ConnectShadowsocks(server)
			}

		})
		serversBox.Append(label)
	}


	//serversBox.Append(setForm)


	spaceBox := canvas.NewRectangle(color.Transparent)
	spaceBox.SetMinSize(fyne.NewSize(200, 30))
	serversBox.Append(spaceBox)

	tipLabel := widget.NewLabel("")
	tipLabel.SetText("点击服务器列表直接连接,显示连接成功后关闭窗口即可,代理会在后台继续运行。")
	serversBox.Append(tipLabel)

	//msg
	msgLabel = widget.NewLabel("")
	serversBox.Append(msgLabel)


	w.SetContent(serversBox)
	w.Resize(fyne.NewSize(500, 600))
	w.CenterOnScreen()


	//w.SetContent(widget.NewVBox(widget.NewLabel("hdalsdqw")))

	//go func() {
	//	time.Sleep(5 * time.Second)
	//	w.Hide()
	//}()

	w.ShowAndRun()
	defer a.Quit()
}

func switchHttpPortEntry(b bool) {
	if b {
		portHttpEntry.Enable()
	} else {
		portHttpEntry.Disable()
	}
}
/**
proxy.exe sps -S socks -T tcp -P 127.0.0.1:1006 -t tcp -p :18080
 */
func SocksToHttpProxy() {
	if !localCfg.EnableSocksToHttp {
		return
	}
	//time.Sleep(2 * time.Second)
	exePath := GetFilePathInWd("/libs/proxy-windows-amd64/goproxy.exe")
	//start goproxy
	log.Printf("start goproxy=%s", "")
	cp := fmt.Sprintf("sps -S socks -T tcp -P 127.0.0.1:%s -t tcp -p :%s", localCfg.ClientPort, localCfg.ClientPortHttp)
	cmd := exec.Command(exePath, strings.Split(cp, " ")...)
	HideCmd(cmd)
	go RunCmd(cmd)
	//msgLabel.SetText(fmt.Sprintf("已连接到 %s %s, OK!", cfg.NameEn, ""))
}

/**
shadowsocks2-win64.exe -c ss://AES-128-GCM:ll@lintun.kda.io:51443 -verbose -socks :11888
 */
func ConnectShadowsocks(cfg AioCrossTrojanServer) {
	exePath := GetFilePathInWd("/libs/shadowsocks2-win64.exe")

	//start shadowsocks
	log.Printf("start shadowsocks=%s", cfg.Name)
	cp := fmt.Sprintf("ss://%s:%s@%s:%d", cfg.Method, cfg.PasswordSs, cfg.Host, cfg.PortSs)
	cmd := exec.Command(exePath, "-c", cp, "-socks", ":" + localCfg.ClientPort)
	HideCmd(cmd)
	proxyCmd = cmd
	go RunCmd(cmd)
	go RunTestCmd()
	go SocksToHttpProxy()
	//msgLabel.SetText(fmt.Sprintf("Connected to %s %s, OK!", cfg.NameEn, cfg.Name))
	msgLabel.SetText(fmt.Sprintf("已连接到 %s %s, OK!", cfg.NameEn, ""))
	//RunCmd()
}

func GetFilePathInWd(elem ...string) string {
	return filepath.FromSlash(path.Join(wd, path.Join(elem...)))
}

func ConnectTrojan(cfg AioCrossTrojanServer) {

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

	//trojanRootPath := GetFilePathInWd(TROJAN_ROOT_PATH)
	trojanCfgPath := GetFilePathInWd(TROJAN_CFG_ROOT_PATH)
	ccPath := GetFilePathInWd(TROJAN_CFG_ROOT_PATH, "/client.json")
	exePath := GetFilePathInWd(TROJAN_ROOT_PATH, "/trojan.exe")

	cfg.PortClient, err = strconv.Atoi(localCfg.ClientPort)
	cc := strings.ReplaceAll(string(res.Body()), "CLIENT_PORT", fmt.Sprintf("%d", cfg.PortClient))
	cc = strings.ReplaceAll(cc, "CERT_DIR", strings.ReplaceAll(trojanCfgPath, "\\", "/") )



	os.MkdirAll(trojanCfgPath, 0777)

	//write config
	fmt.Printf("ConnectTrojan clientJson=%s\n", cc)
	err = ioutil.WriteFile(ccPath, []byte(cc), 0777)
	if err != nil {
		log.Println(err)
	}
	//start trojan
	log.Printf("start trojan=%s", cfg.Name)
	cmd := exec.Command(exePath, "-c", ccPath)
	HideCmd(cmd)
	proxyCmd = cmd
	go RunCmd(cmd)
	go RunTestCmd()
	go SocksToHttpProxy()
	//msgLabel.SetText(fmt.Sprintf("Connected to %s %s, OK!", cfg.NameEn, cfg.Name))
	msgLabel.SetText(fmt.Sprintf("已连接到 %s %s, OK!", cfg.Name, ""))
	//RunCmd()
}

func HideCmd(cmdParam *exec.Cmd) {
	cmdParam.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}

func StopCmd() {
	if proxyCmd != nil && proxyCmd.Process != nil {
		log.Println("kill process")
		proxyCmd.Process.Kill()
		proxyCmd.Wait()
		time.Sleep(2 * time.Second)
		log.Println("kill proxyCmd done")
		return
	}

	processes, err := ps.Processes()
	if err != nil {
		log.Println(err)
	}
	for _, proc := range processes {
		if strings.Contains(proc.Executable(), "trojan") || strings.Contains(proc.Executable(), "shadowsocks2") || strings.Contains(proc.Executable(), "goproxy") {
			exe := "taskkill"

			acmd := exec.Command(exe, "/T",  "/F", "/PID", strconv.Itoa(proc.Pid()))
			HideCmd(acmd)
			out, err := acmd.Output()
			if err != nil {
				log.Println(err)
			}
			acmd.Run()
			log.Printf("taskkill res=%s", out)
		}
	}
	time.Sleep(2 * time.Second)
	log.Println("kill process done")

	//if localCfg.LastTrojanPid == 0 {
	//	return
	//}
	//proc, err := ps.FindProcess(localCfg.LastTrojanPid)
	//if err != nil {
	//	log.Println(err)
	//}



}

func RunCmd(cmdp *exec.Cmd) {
	log.Println("start process")
	cmdp.Run()
	//defer stopcmdp()
}

func RunTestCmd() {
	time.Sleep(2 * time.Second)
	//curl --socks5 127.0.0.1:1006 icanhazip.com
	exe := GetFilePathInWd("/libs/curl.exe")
	acmd := exec.Command(exe, "--socks5", "127.0.0.1:" + localCfg.ClientPort, "icanhazip.com")
	HideCmd(acmd)
	out, err := acmd.Output()
	if err != nil {
		log.Println(err)
	}
	acmd.Run()
	log.Printf("RunTestCmd res=%s", out)
	msgLabel.SetText(msgLabel.Text + "\n测试成功,当前IP地址是: " + string(out))
}

