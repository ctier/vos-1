/*
	Copyright 2021	https://github.com/LeoBest2/vos

	!! 未实际测试，仅做思路参考 !!
*/

package main

import (
	"encoding/json"
	"fmt"
	"vos3000"
)

func main() {
	/*	默认不执行修改操作, 设置false时执行实际操作
		vos3000.NotActuallyExecute = false
	*/
	gwObjs, err := vos3000.GetGatewayObject("127.0.0.1:8080", vos3000.GET_MAPPING)
	if err != nil {
		fmt.Println("获取对接网关失败: " + err.Error())
	} else {
		b, _ := json.Marshal(gwObjs)
		fmt.Println("获取到以下对接网关:\n" + string(b))
	}
	errs := vos3000.SyncGatewayObject("127.0.0.1:8080",
		[]string{"127.0.0.1:8081", "127.0.0.1:8082"},
		vos3000.SYNC_ROUTING,
		func(gw *vos3000.GatewayObject) bool { return true })
	if len(errs) != 0 {
		fmt.Printf("sync errors: %v\n", errs)
	}
}
