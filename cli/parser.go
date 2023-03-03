package cli

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gosuri/uitable"
	"gopkg.in/ini.v1"
)

// 客户端支持模块名称
const (
	HOSTS    = "host"
	MONITOR  = "monitor"
	RELEASE  = "release"
	DEPLOY   = "deploy"
	INSTANCE = "instance"
	VERSION  = "--version"
)
const verMsg = "客户端版本 2.6"

// 通用变量
var (
	actionFlag  string
	groupFlag   string
	nameFlag    string
	commentFlag string
)

// ump服务器
type Conf struct {
	Server string
	Port   string
	ApiVer string
}

// 文件仓库服务器
type Registry struct {
	Server string
	Port   string
	Api    string
}

type UmpModule struct {
	ColumnNameList []string            `json:"column_name"`
	Display        []map[string]string `json:"display"`
	DisplayType    string              `json:"display_type"`
	ModuleName     string              `json:"module_name"`
	ModuleStatus   int                 `json:"module_status"`
	Parameter      map[string]string   `json:"parameter"`
}
type ResponseBody struct {
	Module UmpModule `json:"module"`
	Code   int       `json:"code"`
	Msg    string    `json:"msg"`
}

type UmpTable struct {
	Table *uitable.Table
}

var conf *Conf
var reg *Registry

func init() {
	binary, _ := os.Executable()
	fileRoot := filepath.Dir(binary)
	configFile := filepath.Join(fileRoot, "ump.cnf")
	cfg, err := ini.Load(configFile)
	if err != nil {
		fmt.Println("读取配置文件错误", err)
		os.Exit(1)
	}
	conf = new(Conf)
	conf.Server = cfg.Section("server").Key("host").String()
	conf.Port = cfg.Section("server").Key("port").String()
	conf.ApiVer = cfg.Section("server").Key("api_version").String()

	reg = new(Registry)
	reg.Server = cfg.Section("registry").Key("host").String()
	reg.Port = cfg.Section("registry").Key("port").String()
	reg.Api = cfg.Section("registry").Key("api").String()
}
func RunCMD() {
	var cmdJson []byte
	switch os.Args[1] {
	case VERSION:
		fmt.Println(verMsg)
		os.Exit(0)
	case HOSTS:
		hostscmd := flag.NewFlagSet("hosts", flag.ExitOnError)
		hostscmd.StringVar(&actionFlag, "action", "", "动作事件")
		hostscmd.StringVar(&groupFlag, "group", "", "组名称")
		hostscmd.StringVar(&nameFlag, "name", "", "名称")
		hostscmd.StringVar(&commentFlag, "comment", "", "备注")
		address := hostscmd.String("address", "", "节点IP地址")
		username := hostscmd.String("user", "", "节点服务器用户")
		password := hostscmd.String("password", "", "节点服务器密码")
		hostscmd.Parse(os.Args[2:])
		hostsModule := new(HostsModuleCli)
		hostsModule.ModuleName = HOSTS
		hostsModule.Name = nameFlag
		hostsModule.Action = actionFlag
		hostsModule.Group = groupFlag
		hostsModule.Comment = commentFlag
		hostsModule.User = *username
		hostsModule.Password = *password
		hostsModule.Address = *address
		j, err := json.Marshal(hostsModule)
		if err != nil {
			log.Fatalf("命令编译发生内部错误. Error: %s", err.Error())
		}
		cmdJson = j
	case MONITOR:
		monitorCmd := flag.NewFlagSet("monitor", flag.ExitOnError)
		monitorCmd.StringVar(&actionFlag, "action", "", "动作事件")
		monitorCmd.StringVar(&groupFlag, "group", "", "组名称")
		monitorCmd.StringVar(&commentFlag, "comment", "", "备注")
		freq := monitorCmd.String("freq", "", "监控频率")
		jobId := monitorCmd.String("jobid", "", "监控实例id")
		auto := monitorCmd.String("auto", "", "是否开启自动监控")
		collector := monitorCmd.String("collector", "", "部署资源采集器")
		cpath := monitorCmd.String("cpath", "", "资源采集器路径")
		monitorType := monitorCmd.String("type", "", "信息类型, status 查看监控运行状态,metrics 查看节点资源指标")
		monitorCmd.Parse(os.Args[2:])
		monitorModule := new(MonitorModuleCli)
		monitorModule.ModuleName = MONITOR
		monitorModule.Name = nameFlag
		monitorModule.Action = actionFlag
		monitorModule.Group = groupFlag
		monitorModule.Comment = commentFlag
		monitorModule.Freq = *freq
		monitorModule.Auto = *auto
		monitorModule.Collector = *collector
		monitorModule.Jobid = *jobId
		monitorModule.Cpath = *cpath
		monitorModule.CmdType = *monitorType
		j, err := json.Marshal(monitorModule)
		if err != nil {
			log.Fatalf("命令编译发生内部错误. Error: %s", err.Error())
		}
		cmdJson = j
	case RELEASE:
		releaseCmd := flag.NewFlagSet("release", flag.ExitOnError)
		releaseCmd.StringVar(&nameFlag, "name", "", "应用名称")
		releaseCmd.StringVar(&actionFlag, "action", "", "动作事件")
		releaseCmd.StringVar(&commentFlag, "comment", "", "备注")
		tag := releaseCmd.String("tag", "", "发布版本标签")
		src := releaseCmd.String("src", "", "应用本地路径")
		releaseCmd.Parse(os.Args[2:])
		releaseModule := new(ReleaseModuleCli)
		releaseModule.ModuleName = RELEASE
		releaseModule.Name = nameFlag
		releaseModule.Action = actionFlag
		releaseModule.Comment = commentFlag
		releaseModule.Tag = *tag
		if actionFlag == "set" {
			// 判断本地路径是否存在
			fileinfo, err := os.Stat(*src)
			if err != nil {
				fmt.Println("输入文件路径" + *src)
				fmt.Println("错误: 本地文件不存在")
				os.Exit(1)
			}
			if fileinfo.IsDir() {
				fmt.Println("输入文件路径" + *src)
				fmt.Println("错误: 路径是目录")
				os.Exit(1)
			}
			localfileName := fileinfo.Name()
			localfile := strings.Split(localfileName, ".")
			releaseModule.OriginName = localfile[0]
			releaseModule.OriginSuffix = localfile[1]
			releaseModule.Size = uint(fileinfo.Size())
			j, err := json.Marshal(releaseModule)
			if err != nil {
				log.Fatalf("命令编译发生内部错误. Error: %s", err.Error())
			}
			recvData := SendCmd(j)
			res := new(ResponseBody)
			json.Unmarshal([]byte(recvData), res)
			if res.Module.ModuleStatus >= 500 {
				display(res)
				os.Exit(1)
			}
			fileId := res.Module.Parameter["fileId"]
			// 上传应用文件
			result, err := UploadApp(fileId, *tag, nameFlag, localfileName, *src)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println(string(result))
			os.Exit(0)
		} else {
			j, err := json.Marshal(releaseModule)
			if err != nil {
				log.Fatalf("命令编译发生内部错误. Error: %s", err.Error())
			}
			cmdJson = j
		}
	case DEPLOY:
		deployCmd := flag.NewFlagSet(DEPLOY, flag.ExitOnError)
		deployCmd.StringVar(&actionFlag, "action", "", "动作事件")
		deployCmd.StringVar(&groupFlag, "group", "", "组名称")
		deployCmd.StringVar(&nameFlag, "name", "", "部署名称")
		deployCmd.StringVar(&commentFlag, "comment", "", "备注")
		app := deployCmd.String("app", "", "部署应用的名称,样例: appname:tag")
		dest := deployCmd.String("dest", "", "部署到节点的路径")
		history := deployCmd.String("history", "", "查看部署历史")
		detail := deployCmd.String("detail", "", "查看部署详细信息")
		health := deployCmd.String("health", "", "应用健康检测，默认开启")
		args := deployCmd.String("args", "", "应用启动参数")
		deployCmd.Parse(os.Args[2:])
		deployModule := new(DeployModuleCli)
		deployModule.ModuleName = DEPLOY
		deployModule.Name = nameFlag
		deployModule.Action = actionFlag
		deployModule.Group = groupFlag
		deployModule.Comment = commentFlag
		deployModule.App = *app
		deployModule.Dest = *dest
		deployModule.History = *history
		deployModule.Detail = *detail
		deployModule.Health = *health
		deployModule.Args = *args
		j, err := json.Marshal(deployModule)
		if err != nil {
			log.Fatalf("命令编译发生内部错误. Error: %s", err.Error())
		}
		cmdJson = j
	case INSTANCE:
		instanceCmd := flag.NewFlagSet(DEPLOY, flag.ExitOnError)
		instanceCmd.StringVar(&actionFlag, "action", "", "动作事件")
		instanceCmd.StringVar(&nameFlag, "name", "", "部署名称")
		instanceCmd.StringVar(&commentFlag, "comment", "", "备注")
		deployName := instanceCmd.String("deploy-name", "", "部署应用的名称")
		control := instanceCmd.String("control", "", "实例启停控制(start | stop)")
		insid := instanceCmd.String("insid", "", "实例id")
		instanceCmd.Parse(os.Args[2:])
		instanceModule := new(InstanceModuleCli)
		instanceModule.ModuleName = INSTANCE
		instanceModule.Action = actionFlag
		instanceModule.Comment = commentFlag
		instanceModule.DeployName = *deployName
		instanceModule.Control = *control
		instanceModule.Insid = *insid
		j, err := json.Marshal(instanceModule)
		if err != nil {
			log.Fatalf("命令编译发生内部错误. Error: %s", err.Error())
		}
		cmdJson = j
	}
	action := strings.ToLower(actionFlag)
	if action != "get" && action != "set" && action != "delete" {
		fmt.Println("action参数不正确")
		os.Exit(0)
	}
	recvData := SendCmd(cmdJson)
	UmpServerEcho(recvData)
}

func UploadApp(fileId, tag, name, filename, filepath string) ([]byte, error) {
	httpClient := &http.Client{
		Timeout: 5 * time.Second, // 请求超时时间
	}
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	// 在表单中创建一个文件字段
	formFile, err := writer.CreateFormFile("file", filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// 打开文件句柄
	f, err := os.Open(filepath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// 读取文件内容到表单文件字段
	_, err = io.Copy(formFile, f)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// 将其他参数写入到表单
	writer.WriteField("fileId", fileId)
	writer.WriteField("appTag", tag)
	writer.WriteField("appName", name)
	if err = writer.Close(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// 私有库地址
	regServer := reg.Server
	regServerPort := reg.Port
	regServerApi := reg.Api
	regURL := "http://" + regServer + ":" + regServerPort + "/" + regServerApi
	// 构造请求对象
	req, err := http.NewRequest("POST", regURL, body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	// 发送请求
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func SendCmd(cmd []byte) []uint8 {
	umpServer := conf.Server
	umpServerPort := conf.Port
	umpServerApiVer := conf.ApiVer
	umpCmdApiUrl := "http://" + umpServer + ":" + umpServerPort + "/" + "cmd" + "/" + umpServerApiVer
	request, error := http.NewRequest("POST", umpCmdApiUrl, bytes.NewBuffer(cmd))
	if error != nil {
		panic(error)
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	client := &http.Client{}
	response, error := client.Do(request)
	if error != nil {
		fmt.Println("错误, 无法连接UMP服务端!")
		os.Exit(1)
	}
	defer response.Body.Close()
	recvData, _ := ioutil.ReadAll(response.Body)
	fmt.Println("debug->response Body:", string(recvData))
	return recvData
}

func UmpServerEcho(data []byte) {
	res := new(ResponseBody)
	json.Unmarshal([]byte(data), res)
	display(res)
}

func display(res *ResponseBody) {
	umpTable := new(UmpTable)
	table := uitable.New()
	umpTable.Table = table
	if res.Module.DisplayType == "feedback" {
		// table.MaxColWidth = 100
		table.Wrap = true
		firstLineName := "服务器状态"
		firstLineValue := res.Code
		secondLineName := "模块名称"
		secondLineValue := res.Module.ModuleName
		thirdLineName := "模块返回值"
		moduleStatus := res.Module.ModuleStatus
		thirdLineValue := moduleStatus
		fourthLineName := "信息"
		displayInfo := res.Module.Display[0][strconv.Itoa(moduleStatus)]
		fourthLineValue := displayInfo
		table.AddRow(firstLineName, firstLineValue)
		table.AddRow(secondLineName, secondLineValue)
		table.AddRow(thirdLineName, thirdLineValue)
		table.AddRow(fourthLineName, fourthLineValue)
		table.AddRow("")
		fmt.Println(table)
	} else {
		//  table.MaxColWidth = 100
		table.Wrap = true
		columnDisplay := res.Module.ColumnNameList
		umpTable.AddRow(columnDisplay)
		displayRows := res.Module.Display
		for _, displayValue := range displayRows {
			row := make([]string, 0)
			for _, colName := range columnDisplay {
				row = append(row, displayValue[colName])
			}
			umpTable.AddRow(row)
		}

		fmt.Println(table)
	}

}

func (t *UmpTable) AddRow(data []string) *uitable.Table {
	r := &uitable.Row{Cells: make([]*uitable.Cell, len(data))}
	for i, d := range data {
		if i == (len(data) - 1) {
			// 最后一列显示长字符串
			r.Cells[i] = &uitable.Cell{Data: d, Width: 35, RightAlign: false}
		} else {
			r.Cells[i] = &uitable.Cell{Data: d, Width: 20, RightAlign: false}
		}
	}
	t.Table.Rows = append(t.Table.Rows, r)
	return t.Table
}
