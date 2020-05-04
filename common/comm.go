package common

import (
	"crypto/md5"
	"errors"
	"fmt"
	"log"
	"net"
	"reflect"
	"strconv"
	"time"
)


func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

//struct -> map
func Struct2Map(obj interface{}) (data map[string]interface{}){
	reflectType := reflect.TypeOf(obj)
	reflectValue := reflect.ValueOf(obj)

	data = map[string]interface{}{}
	for i := 0; i < reflectType.NumField(); i++ {
		data[reflectType.Field(i).Name] = reflectValue.Field(i).Interface()
	}

	return data
}

//指向struct的指针 -> map
func StructPtr2Map(obj interface{}) (data map[string]interface{}){
	//通过指针取出结构体
	objElem := reflect.ValueOf(obj).Elem().Interface()
	//结构体转map
	return Struct2Map(objElem)
}

//[]struct -> []map, interface{}可以表示slice，不需要[]interface{}否则报错提示参数不能转换
func StructArray2MapArray(objArray interface{}) (data []map[string]interface{}){
	i := 0
	//slice长度
	iMax := reflect.ValueOf(objArray).Len() - 1

	data = []map[string]interface{}{}
	for {
		if i > iMax {
			break
		}
		//数组中第i个元素转换为map
		mapTemp := Struct2Map(reflect.ValueOf(objArray).Index(i).Interface())
		//mapTemp := structs.Map(reflect.ValueOf(objArray).Index(i).Interface())
		data = append(data, mapTemp)
		i++
	}

	return data
}

//slice中元素为指向struct的指针时，转换为map的slice
func StructPtrArray2MapArray(objArray interface{}) (data []map[string]interface{}){
	i := 0
	iMax := reflect.ValueOf(objArray).Len() - 1

	data = []map[string]interface{}{}
	for {
		if i > iMax {
			break
		}

		mapTemp := StructPtr2Map(reflect.ValueOf(objArray).Index(i).Interface())
		//mapTemp := structs.Map(reflect.ValueOf(objArray).Index(i).Interface())
		data = append(data, mapTemp)
		i++
	}

	return data
}



//根据结构体中sql标签映射数据到结构体中并且转换类型
func DataToStructByTagSql(data map[string]string, obj interface{}) {
	objValue := reflect.ValueOf(obj).Elem()
	for i := 0; i < objValue.NumField(); i++ {
		//获取sql对应的值
		value := data[objValue.Type().Field(i).Tag.Get("sql")]
		//获取对应字段的名称
		name := objValue.Type().Field(i).Name
		//获取对应字段类型
		structFieldType := objValue.Field(i).Type()
		//获取变量类型，也可以直接写"string类型"
		val := reflect.ValueOf(value)
		var err error
		if structFieldType != val.Type() {
			//类型转换
			val, err = TypeConversion(value, structFieldType.Name()) //类型转换
			if err != nil {

			}
		}
		//设置类型值
		objValue.FieldByName(name).Set(val)
	}
}

//类型转换
func TypeConversion(value string, ntype string) (reflect.Value, error) {
	if ntype == "string" {
		return reflect.ValueOf(value), nil
	} else if ntype == "time.Time" {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", value, time.Local)
		return reflect.ValueOf(t), err
	} else if ntype == "Time" {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", value, time.Local)
		return reflect.ValueOf(t), err
	} else if ntype == "int" {
		i, err := strconv.Atoi(value)
		return reflect.ValueOf(i), err
	} else if ntype == "int8" {
		i, err := strconv.ParseInt(value, 10, 64)
		return reflect.ValueOf(int8(i)), err
	} else if ntype == "int32" {
		i, err := strconv.ParseInt(value, 10, 64)
		return reflect.ValueOf(int64(i)), err
	} else if ntype == "int64" {
		i, err := strconv.ParseInt(value, 10, 64)
		return reflect.ValueOf(i), err
	} else if ntype == "float32" {
		i, err := strconv.ParseFloat(value, 64)
		return reflect.ValueOf(float32(i)), err
	} else if ntype == "float64" {
		i, err := strconv.ParseFloat(value, 64)
		return reflect.ValueOf(i), err
	}

	//else if .......增加其他一些类型的转换

	return reflect.ValueOf(value), errors.New("未知的类型：" + ntype)
}


func StringMd5(str string) string {
	md5Arr := md5.Sum([]byte(str))
	return fmt.Sprintf("%x", md5Arr[:])
}



func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		return ""
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return ""
}

