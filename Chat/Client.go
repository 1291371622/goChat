package main

import (
	"Chat/Common_tools"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"time"
)

type userPro struct {
	id string//账号
	pass string//密码
	conn net.Conn
}

func main() {
Menu:
	isFlag := true
	str := `
	***********
	* -1-登录 *
	* -2-注册 *
	* -3-退出 *
	***********`

	var choice int

	for isFlag {

		fmt.Println("\tLet's to Chat! ^_^")
		fmt.Println(str)
		fmt.Scanln(&choice)
		switch choice {
		case 1:
			choice = 1
			isFlag = false
		case 2:
			choice = 2
			isFlag = false
		case 3:
			isFlag = false
		default:
			fmt.Println("没有此选项")
		}

	}

	switch choice {
	case 1:
		RepeatLogin:
			user := new(userPro)
			err := user.Login()
			if err != nil {
				fmt.Println(err)
				goto RepeatLogin
			}
	case 2:
		RepeatReg:
			user := new(userPro)
			err := user.Register()
			if err != nil {
				fmt.Println(err)
				goto RepeatReg
			}else {
				fmt.Println("注册成功！")
				goto Menu
			}
	case 3:


	}




}


func (this *userPro) Conn(){
	conn,err:=net.Dial("tcp","127.0.0.1:6666")
	Common_tools.HanddleError(err,"连接客户端失败，请检查客户端拨号是否正确")
	this.conn=conn
}

func (this *userPro) Login() (error){
	fmt.Println("请输入账号")
	var id string
	fmt.Scanln(&id)
	fmt.Println("请输入密码")
	var password string
	fmt.Scanln(&password)
	this.Conn()
	defer this.conn.Close()
	conn := this.conn
	userLogin := Common_tools.UserLogin{
		Id: id,
		Pass: password,
	}
	userLogin_byte,err := json.Marshal(userLogin)
	Common_tools.HanddleError(err,"登录序列化失败")
	data := Common_tools.Data {
		DataType: Common_tools.LoginType,
		Message:  string(userLogin_byte),
	}
	data_byte,err:=json.Marshal(data)
	Common_tools.HanddleError(err,"数据序列化失败")
	_,err=conn.Write(data_byte)
	var statusData=new(Common_tools.Data)
	var status=new(Common_tools.Status)
	var byte_slice []byte
	for  {
		_,read_slice,n,err:=Common_tools.Read(this.conn)
		err=json.Unmarshal(read_slice[:n],statusData)
		Common_tools.HanddleError(err,"客户端读取数据序列化失败")
		byte_slice=[]byte(statusData.Message)
		err=json.Unmarshal(byte_slice,status)
		Common_tools.HanddleError(err,"客户端序列化状态失败")
		if status.Code!="200" {
			fmt.Println("invalid account,账号密码错误,请重新输入!")
			fmt.Println()
			main()
		}else{
			break
		}
	}

	if status.OnlineUser==nil{
		fmt.Println("当前暂时没有用户在线")
	}else {
		fmt.Println("当前在线用户如下")
		for _,id:=range status.OnlineUser{
			fmt.Println("用户:",id,"在线中")
		}
	}
	//挂起协程,随时监听用户上下线信息
	go keepReceive(conn)
	//显示登录成功的菜单
	loginSuccessMenu(conn,id)

	return nil
}

func (this *userPro) Register() (error){
	fmt.Println("->请输入账号<-")
	var id string
	fmt.Scanln(&id)
	fmt.Println("->请输入密码<-")
	var password string
	fmt.Scanln(&password)
	this.Conn()
	defer this.conn.Close()
	conn:=this.conn
	user:=Common_tools.UserRegister{
		Id:   id,
		Pass: password,
	}
	user_byte,err:=json.Marshal(user)
	Common_tools.HanddleError(err,"注册序列化失败")
	data:=Common_tools.Data{
		DataType: Common_tools.RegisterType,
		Message:  string(user_byte),
	}
	data_byte,err:=json.Marshal(data)
	Common_tools.HanddleError(err,"注册数据序列化失败")
	_,err=conn.Write(data_byte)
	var statusData=new(Common_tools.Data)
	var status=new(Common_tools.Status)
	var byte_slice []byte
	for   {
		_,read_slice,n,err:=Common_tools.Read(this.conn)
		err=json.Unmarshal(read_slice[:n],statusData)
		Common_tools.HanddleError(err,"客户端读取数据序列化失败")
		byte_slice=[]byte(statusData.Message)
		err=json.Unmarshal(byte_slice,status)
		Common_tools.HanddleError(err,"客户端序列化状态失败")
		if status.Code!="200" {
			return errors.New("注册失败,该账号被注册了")
		}else{

			break
		}
	}


	return nil
}


//查看在线用户
func lookOnlineUsers(conn net.Conn) {
	data:=Common_tools.Data{
		DataType: Common_tools.LookOnlineUser,
		Message:  "1",
	}
	data_byte,err:=json.Marshal(data)
	Common_tools.HanddleError(err,"查看在线用户序列化失败")
	_,err=conn.Write(data_byte)
	Common_tools.HanddleError(err,"查看在线用户发送失败")
}

//保持连接,接收用户上下线信息
func keepReceive(conn net.Conn)  {
	var data=new(Common_tools.Data)
	for  {
		_,read_slice,n,err:=Common_tools.Read(conn)
		err=json.Unmarshal(read_slice[:n],data)
		Common_tools.HanddleError(err,"keepReceive出错")
		switch data.DataType {
		case Common_tools.FriendStatus:
			//好友上下线信息从这里通知
			fmt.Println(data.Message)
		case Common_tools.LookOnlineUser:
			//查看所有在线好友列表
			var usersId []string
			err:=json.Unmarshal([]byte(data.Message),&usersId)
			Common_tools.HanddleError(err,"客户端查看好友在线列表序列化失败")
			fmt.Println("->在线用户列表<-")
			for k,v:=range usersId{
				k++
				fmt.Println("序号->",k,"账号->",v)
			}

		case Common_tools.MessageType:
			fmt.Println(data.Message)
		case Common_tools.IsOnline:
			isOnlineChannel<-data.Message

		}


	}
}

var isOnlineChannel=make(chan string,1)
func loginSuccessMenu(conn net.Conn,id string)  {

	fmt.Println("登录成功")
	str:=`
********************
* -1>查看在线好友 * 
* -2>开始聊天    *
* -3>退出       *
****************
`
	into:=true
	var input int
	for into {
		fmt.Println(str)
		fmt.Scanln(&input)
		switch input {
		case 1:
			//查看在线用户
			lookOnlineUsers(conn)
		case 2:
			//
			selectFriendToChat(conn,id)
		case 3:
			//退出系统
			into=false
		default:
			fmt.Println("输入有误,请重新输入")
		}
	}
}


func selectFriendToChat(conn net.Conn,id string){
IntoChat:
	bool:=true
	var fid string	//好友id
	chatbox:=new(Common_tools.ChatBox)
	data:=new(Common_tools.Data)
	for bool {
		//展现当前好友ID
		fmt.Println("当前好友在线情况如下:")
		lookOnlineUsers(conn)
		fmt.Println("请输入对方用户的账号,向他发起聊天,按回车发送聊天请求")
		fmt.Print("对方的账号:")
		fmt.Scanln(&fid)
		if fid==id {
			fmt.Println("您不能输入自己的ID,请重新输入")
		}else{

			chatbox.From=id
			chatbox.To=fid
			bool=false
		}
	}
	var message ,input string
	var loop =true
	fmt.Println("->请输入发送内容,按回车发送消息<-")
	for loop {
		fmt.Scanln(&message)
		chatbox.Message=time.Now().Format("2006-01-02 15:04:05")+"\t"+chatbox.From+"对"+chatbox.To+"说:"
		chatbox.Message+=message
		chatbox_byte,err:=json.Marshal(chatbox)
		Common_tools.HanddleError(err,"->消息盒子序列化失败<-")
		data.DataType=Common_tools.MessageType
		data.Message=string(chatbox_byte)
		data_byte,err:=json.Marshal(data)
		Common_tools.HanddleError(err,"->消息盒子数据序列化失败<-")
		//发送前来一次探测,对方Id是否在线,或者存在:
		online:=isOnline(conn,fid)
		if online {
			_,err=conn.Write(data_byte)
			if err!=nil {
				fmt.Println("发送失败!")
			}else{
				fmt.Println("发送成功!")
			}
		}else{
			fmt.Println("对方可能下线了,也可能是您输了一个不存在的账号!")
			for loop {
				fmt.Println("->输入1按回车,返回聊天大厅<-")
				fmt.Println("->输入2按回车,退出系统<-")
				fmt.Scanln(&input)
				switch input {
				case "1":
					goto IntoChat
				case "2":
					fmt.Println("bye bye!")
					loop=false
				default:
					fmt.Println("->输入错误,请重新输入<-")
				}
			}
		}

	}

}

//探测对方是否在线

func isOnline(conn net.Conn,id string) bool{
	online:=false
	data:=new(Common_tools.Data)
	data.DataType=Common_tools.IsOnline
	data.Message=id
	data_byte,err:=json.Marshal(data)
	Common_tools.HanddleError(err,"探测对方是否在线,序列化失败")
	_,err=conn.Write(data_byte)
	if err!=nil {
		fmt.Println("探测对方是否在线发送失败")
	}
	var begin=true
	for begin {
		read:=<-isOnlineChannel
		if read=="1" {
			online=true
			break
		}else{
			break
		}
	}

	return online
}