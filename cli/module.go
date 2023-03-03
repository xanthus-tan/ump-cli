package cli

type UmpModuleCli struct {
	ModuleName string `json:"module"`
	Group      string `json:"group"`
	Name       string `json:"name"`
	Action     string `json:"action"`
	Comment    string `json:"comment"`
}
type HostsModuleCli struct {
	UmpModuleCli
	User     string `json:"user"`
	Password string `json:"password"`
	Address  string `json:"address"`
}
type MonitorModuleCli struct {
	UmpModuleCli
	Freq      string `json:"freq"`
	Jobid     string `json:"jobid"`
	Auto      string `json:"auto"`
	Collector string `json:"collector"`
	Cpath     string `json:"cpath"`
	CmdType   string `json:"type"`
}
type ReleaseModuleCli struct {
	UmpModuleCli
	Tag          string `json:"tag"`
	Filename     string `json:"filename"`
	Size         uint   `json:"size"`
	OriginName   string `json:"originName"`
	OriginSuffix string `json:"originSuffix"`
}
type DeployModuleCli struct {
	UmpModuleCli
	App     string `json:"app"`
	Dest    string `json:"dest"`
	History string `json:"history"`
	Detail  string `json:"detail"`
	Health  string `json:"health"`
	Args    string `json:"args"`
}
type InstanceModuleCli struct {
	UmpModuleCli
	DeployName string `json:"deploy-name"`
	Control    string `json:"control"`
	Insid      string `json:"insid"`
}
