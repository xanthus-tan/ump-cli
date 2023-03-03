# UMP客户端
ump客户端
版本2.6
## 命令示例
``` ump-cli.exe [模块名称] --action [set/get/delete] --参数1 --参数2.. ```
### hosts模块命令示例
#### 添加节点
ump-cli.exe hosts --action set --group g1 --address 192.168.72.130,192.168.72.132 --user xanthus --password 1  
#### 查询节点
ump-cli.exe hosts --action get --group g1  
#### 删除节点
ump-cli.exe hosts --action delete --group g1  
ump-cli.exe hosts --action delete --group g1 --address 192.168.72.130  
### monitor模块命令示例
#### 查看监控状态
ump-cli.exe monitor --action get --group g1 --type status  
#### 部署采集器
ump-cli.exe monitor --action set --group g1 --collector true --cpath /home/xanthus/ump/collector  
#### 获取节点资源指标
ump-cli.exe monitor --action get --group g1 --type metrics --cpath /home/xanthus/ump/collector  
#### 自动采集资源指标
ump-cli.exe monitor --action set --group g1 --auto true --freq 5 --cpath /home/xanthus/ump/collector  
#### 停止自动采集
ump-cli.exe monitor --action delete --jobid 720d453f80eb11edaba6a4fc7733a40c  
# release模块命令示例
#### 发布应用
ump-cli.exe release --action set --name demo-app --tag 1.0 --src d:\demo-app.jar  
#### 查询部署
ump-cli.exe release --action get  
ump-cli.exe release --action get --name demo-app  
#### 删除部署
ump-cli.exe release --action delete --name demo-app --tag 1.0  
# deploy模块命令示例
#### 组g1的所有节点部署demo-app:1.0  
ump-cli.exe deploy --action set --group g1 --name demo-deploy --app demo-app:1.0 --dest /tmp/  
#### 查看demo-deploy状态
ump-cli.exe deploy --action get --name demo-deploy  
ump-cli.exe deploy --action get --name demo-deploy --detail true  
ump-cli.exe deploy --action get --name demo-deploy --history true  
#### 删除demo-deploy
ump-cli.exe deploy --action delete --name demo-deploy  
# instance模块命令示例
#### 查看实例部署状态
ump-cli.exe instance --action get --deploy-name demo-deploy
#### 部署启停控制
ump-cli.exe instance --action set --deploy-name demo-deploy --control start  
ump-cli.exe instance --action set --deploy-name demo-deploy --control stop  
#### 实例启停控制
ump-cli.exe instance --action set --deploy-name demo-deploy --insid fe7ff01b117411edbb69a4fc7733a40c --control start  
ump-cli.exe instance --action set --deploy-name demo-deploy --insid 70ff911003e511ed80aba4fc7733a40c --control stop  
