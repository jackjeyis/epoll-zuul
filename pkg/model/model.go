package model

type StatusType string

const (
	UP           StatusType = "UP"
	DOWN         StatusType = "DOWN"
	STARTING     StatusType = "STARTING"
	OUTOFSERVICE StatusType = "OUT_OF_SERVICE"
	UNKNOW       StatusType = "UNKNOW"
)

type AppCluster struct {
	ID   string
	Name string
}

type AppInstance struct {
	InstanceId string     `json:"instanceId"`
	HostName   string     `json:"hostName"`
	App        string     `json:"app"`
	IPAddr     string     `json:"ipAddr"`
	VipAddress string     `json:"vipAddress"`
	Status     StatusType `json:status`
	Port       int        `json:"port"`
}
