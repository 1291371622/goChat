package Common_tools

import (
	"github.com/garyburd/redigo/redis"
)

var Instance redis.Conn

type OptionRedis struct {

}

func (this *OptionRedis)  GetRedisConn() redis.Conn{
	if Instance==nil {
		conn, e := redis.Dial("tcp", "127.0.0.1:6379")
		HanddleError(e, "连接出错")
		//_, err := conn.Do("AUTH", "zxc86506859")
		//HanddleError(err, "密码出错")
		Instance=conn
	}
	return Instance
}

func (this *OptionRedis) Option(option string,key... interface{}) interface{} {
	conn:=this.GetRedisConn()
	defer this.Close()
	//执行redis命令
	replay, err := conn.Do(option, key...)
	HanddleError(err, "执行出错")
	//根据具体的业务类型进行数据类型转换

	switch option {
	case "GET":
		redis_string, _ := redis.String(replay, err) //将redis返回的值转换为string,它返回的是uint8

		return redis_string
	case "EXISTS":
		redis_int, err := redis.Int(replay, err)
		HanddleError(err,"redis,exits失败")
		return redis_int
	case "SET":
		redis_string, err := redis.String(replay, err) //将redis返回的值转换为string,它返回的是uint8
		HanddleError(err,"redis,Set失败")
		return redis_string
	}

	return nil
}

func (this *OptionRedis) Close() {
	Instance.Close()
	Instance=nil
}
