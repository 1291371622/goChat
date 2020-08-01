package main



import (
	"Chat/Common_tools"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

var wr_lock sync.RWMutex
//在线的用户
var Users =make(map[string]net.Conn)
//下线的用户
var OfflineUserChannel=make(chan map[string]net.Conn,1024)

//调度
type Transfer struct {

}

func main()  {
	conn,err := net.Listen("tcp", "0.0.0.0:6666")
	Common_tools.HanddleError(err, "监听错误，请检查端口地址是否正确!")
	defer conn.Close()
	var transferController=new(Transfer)
	var notify=new(Notify)
	//挂起协程发送心跳包
	notify.HeartBeat()
	//通知用户**下线了
	notify.NotifyOtherUserOffline()


	for {

		accept, err := conn.Accept()

		Common_tools.HanddleError(err,"接收数据出错")
		go Receive(accept,transferController)

	}
}


func Receive(accept net.Conn, transferController *Transfer) {
	for {
		_, read_slice, n,err := Common_tools.Read(accept)
		if err!=nil{

			break
		}
		var data Common_tools.Data
		err = json.Unmarshal(read_slice[:n], &data)
		Common_tools.HanddleError(err, "反序列化Data失败")
		switch data.DataType {
		case Common_tools.LoginType: //登录类型则校验密码账号
			var login = new(Common_tools.UserLogin)
			byte := []byte(data.Message)
			err = json.Unmarshal(byte, login)
			Common_tools.HanddleError(err, "反序列化账号信息失败")
			var redis = new(Common_tools.OptionRedis)
			res := redis.Option("GET", login.Id)
			var code, msg string
			var onlineUser []string
			if len(res.(string)) > 0 {
				if res == login.Pass {
					//登录成功
					code = "200"
					msg = "登录成功"
					//切片增加用户
					transferController.AddUser(login.Id, accept)
					//获取当前在线所有用户
					onlineUser = transferController.GetOnlineUser()
					//通知其他用户,有人上线了
					var notify = new(Notify)
					notify.NotifyOtherUserOnline(login.Id)
				}
			} else {
				//登录失败
				code = "500"
				msg = "登录失败,请检查您的账号密码"
			}
			status := Common_tools.Status{
				Code:       code,
				Msg:        msg,
				OnlineUser: onlineUser,
			}
			status_byte, err := json.Marshal(status)
			Common_tools.HanddleError(err, "序列化登录状态码失败")
			status_str := string(status_byte)
			data := Common_tools.Data{
				DataType: Common_tools.LoginStatus,
				Message:  status_str,
			}
			data_byte, err := json.Marshal(data)
			Common_tools.HanddleError(err, "序列化登录状态码失败")
			n, err = accept.Write(data_byte)
		case Common_tools.RegisterType: //注册类型
			var register = new(Common_tools.UserRegister)
			msg := string(data.Message)
			msg_byte := []byte(msg)
			err := json.Unmarshal(msg_byte, register)
			Common_tools.HanddleError(err, "注册反序列化失败")
			//检验redis中是否有注册信息
			redis := new(Common_tools.OptionRedis)
			res := redis.Option("EXISTS", register.Id)
			var status = new(Common_tools.Status)
			if res == 1 {
				status.Code = "500"
				status.Msg = "该账号被注册过了"
			} else {
				res := redis.Option("SET", register.Id, register.Pass)
				if res == "OK" {
					status.Code = "200"
					status.Msg = "注册成功"
				} else {
					fmt.Println("注册失败")
					os.Exit(1)
				}
			}
			byte, err := json.Marshal(status)
			Common_tools.HanddleError(err, "注册返回状态失败")
			data := Common_tools.Data{
				DataType: Common_tools.RegisterStatus,
				Message:  string(byte),
			}
			byte, err = json.Marshal(data)
			Common_tools.HanddleError(err, "注册返回数据状态失败")
			n, err = accept.Write(byte)
		case Common_tools.LookOnlineUser:
			//查看用户在线情况
			users := Users
			var usersId []string
			for id, _ := range users {
				usersId = append(usersId, id)
			}
			byte, err := json.Marshal(&usersId)
			Common_tools.HanddleError(err, "服务端,查看用户在线情况序列化失败1")
			data := new(Common_tools.Data)
			data.Message = string(byte)
			data.DataType = Common_tools.LookOnlineUser
			data_byte, err := json.Marshal(data)
			Common_tools.HanddleError(err, "服务端,查看用户在线情况序列化失败2")
			accept.Write(data_byte)
		case Common_tools.IsOnline:	//'
			fmt.Println("接收到探测消息")
			id := data.Message
			isOnline := false
			for uid,_ := range Users {
				if uid == id {
					isOnline=true
					break
				}
			}
			data := new(Common_tools.Data)
			data.DataType = Common_tools.IsOnline
			if isOnline {
				data.Message="1"
			}else{
				data.Message="0"
			}
			byte,_ := json.Marshal(data)
			n,err=accept.Write(byte)

			if err != nil{
				fmt.Println("探测消息返回失败")
			}else{
				fmt.Println("探测信息转发成功")
			}


		case Common_tools.MessageType:
			//普通消息类型
			message := data.Message
			message_byte := []byte(message)
			chatBox := new(Common_tools.ChatBox)
			err := json.Unmarshal(message_byte,chatBox)
			Common_tools.HanddleError(err,"messageType反序列化失败")
			data := new(Common_tools.Data)
			data.DataType = Common_tools.MessageType
			data.Message = chatBox.Message
			data_byte,_ := json.Marshal(data)
			for id,Common_tools := range Users  {
				if id == chatBox.To {
					_,err := Common_tools.Write(data_byte)
					if err != nil {
						fmt.Println(id,"发送不成功,但是不提示")
					}
					break
				}
			}

		}

	}
}


//添加用户进入在线列表
func (this *Transfer) AddUser(id string,user net.Conn){
	Users[id]=user
	fmt.Println(id,"已上线")
}

//该用户由于网络的原因,发送失败,扔进下线队列,然后删除该用户
func (this *Transfer) DelUser(id string,conn net.Conn){

	offlineUser:=make(map[string]net.Conn)
	offlineUser[id]=conn
	OfflineUserChannel<-offlineUser
	wr_lock.Lock()
	delete(Users,id)
	fmt.Println(id,"已下线")
	wr_lock.Unlock()
}

//获取当前用户在线列表
func (this *Transfer) GetOnlineUser() []string{
	var arr  []string
	for k,_:=range Users {
		arr=append(arr,k)
	}
	return  arr
}


/**
上线通知其他用户
*/

type Notify struct {

}

//心跳包
func (this *Notify) HeartBeat(){
	go func() {
		data:=new(Common_tools.Data)
		data.DataType=Common_tools.HeartBeat
		data.Message=""
		byte,err:=json.Marshal(data)
		transfer:=new(Transfer)
		for  {
			for uid,conn:=range Users{
				_,err=conn.Write(byte)
				if err!=nil {
					//由于网络原因,删除该用户
					transfer.DelUser(uid,conn)
				}
			}
			//每一秒发送一次心跳包
			time.Sleep(time.Second*3)
		}
	}()
}


//通知其他用户有人上限
func (this *Notify) NotifyOtherUserOnline(id string){
	data:=new(Common_tools.Data)
	data.DataType=Common_tools.FriendStatus
	data.Message="用户:"+id+"上线了"
	byte,err:=json.Marshal(data)
	Common_tools.HanddleError(err,"通知其他用户好友上线失败")
	wr_lock.RLock()
	for uid,conn:=range Users{
		if id==uid {
			continue
		}
		_,err=conn.Write(byte)
		if err!=nil {
			fmt.Println(err)
		}
	}
	wr_lock.RUnlock()

}


//通知其他用户有人下线
func (this *Notify) NotifyOtherUserOffline(){
	go func() {
		data:=new(Common_tools.Data)
		data.DataType=Common_tools.FriendStatus
		var message string
		for value:=range OfflineUserChannel{
			for id,_:=range value {
				message="用户:"+id+"下线了\n"
			}
			data.Message=message
			byte,err:=json.Marshal(data)
			Common_tools.HanddleError(err,"通知其他用户好友下线失败")
			for _,conn:=range Users{
				_,err=conn.Write(byte)
				if err!=nil {
					fmt.Println(err)
				}
			}
		}
	}()
}


func (this *Notify) ReceiveUserMenu(){

}







func init(){

}
