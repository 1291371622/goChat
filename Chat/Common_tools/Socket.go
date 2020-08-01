package Common_tools

import (
	"net"
)

const (
	LoginType="LoginType"
	RegisterType="RegisterType"
	MessageType="MessageType"
	LoginStatus="LoginStatus"
	RegisterStatus="RegisterStatus"
	FriendStatus="FriendStatus"
	HeartBeat="HeartBeat"
	LookOnlineUser="LookOnlineUser"
	IsOnline="IsOnline"	//探测是否在线类型
)

type Data struct {
	DataType string	`json:"data_type"`
	Message string	`json:"message"`
}

type ChatBox struct {
	From string	`json:"from"`	//发起消息的Id
	To string `json:"to"`	//对方id
	Message string	`json:"message"`	//消息内容
}

type UserLogin struct {
	Id string `json:"id"`
	Pass string `json:"pass"`
}

type UserRegister struct {
	Id string `json:"id"`
	Pass string `json:"pass"`
}

type Status struct {
	Code string	`json:"code"`
	Msg string	`json:"msg"`
	OnlineUser []string `json:"online_user"`
}

func Read(accept net.Conn) (read_str string,read_slice []byte,n int,err error){
	read_slice=make([]byte,2000)
	n,err=accept.Read(read_slice)
	read_str=string(read_slice[:n])

	return
}

