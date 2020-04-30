package main

import (
	"fmt"
	"net/http"
	"os/exec"
)

//检查其他服务端口的服务
func healthCheck(w http.ResponseWriter, r *http.Request) {
	s := r.FormValue("s")

	//service-port的对应关系
	servicePortMap := map[string]string {
		"user" : "9090",
	}

	statusCode := 500

	port, isOk := servicePortMap[s]
	if !isOk {
		w.WriteHeader(statusCode)
		w.Write([]byte("s error"))
		return
	}

	//检查service的端口
	out, err := exec.Command("/bin/bash", "-c", "netstat -antlup | grep " + port).Output()
	if err == nil && len(out) > 0 {
		statusCode = 200
	}

	w.WriteHeader(statusCode)
	w.Write([]byte(""))
}


func main() {
	http.HandleFunc("/check", healthCheck)
	http.ListenAndServe(fmt.Sprintf(":%d", 9091), nil)
}

