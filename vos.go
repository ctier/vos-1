/*
	Copyright 2021	https://github.com/LeoBest2/vos

	!! 未实际测试，仅做思路参考 !!
*/
package vos3000

import (
	"fmt"
	"log"
	"reflect"

	"github.com/parnurzeal/gorequest"
)

var NotActuallyExecute = true

type GetType int
type ChangeType int
type SyncType int

const (
	GET_MAPPING GetType = iota
	GET_MAPPING_ONLINE
	GET_ROUTING
	GET_ROUTING_ONLINE

	CREATE_MAPPING ChangeType = iota
	CREATE_MAPPING_ONLINE
	CREATE_ROUTING
	CREATE_ROUTING_ONLINE

	MODIFY_MAPPING ChangeType = iota
	MODIFY_MAPPING_ONLINE
	MODIFY_ROUTING
	MODIFY_ROUTING_ONLINE

	DELETE_MAPPING ChangeType = iota
	DELETE_MAPPING_ONLINE
	DELETE_ROUTING
	DELETE_ROUTING_ONLINE

	SYNC_MAPPING SyncType = SyncType(GET_MAPPING)
	SYNC_ROUTING SyncType = SyncType(GET_ROUTING)
)

type APIResult struct {
	RetCode                   int64           `json:"retCode"`
	Exception                 string          `json:"exception"`
	InfoGatewayMappings       []GatewayObject `json:"infoGatewayMappings"`
	InfoGatewayMappingOnlines []GatewayObject `json:"infoGatewayMappingOnlines"`
	InfoGatewayRoutings       []GatewayObject `json:"infoGatewayRoutings"`
	InfoGatewayRoutingOnlines []GatewayObject `json:"infoGatewayRoutingOnlines"`
}

type GatewayObject map[string]interface{}

func (g *GatewayObject) Diff(dstGwObj *GatewayObject, attrIgnoredFunc func(name string) bool) (diffObj *GatewayObject) {
	diffObj = &GatewayObject{}
	for k1, v1 := range *g {
		if attrIgnoredFunc != nil && attrIgnoredFunc(k1) {
			continue
		}
		v2, ok := (*dstGwObj)[k1]
		if !ok || !reflect.DeepEqual(v1, v2) {
			(*diffObj)[k1] = v1
		}
	}
	return
}

// GetGatewayObject 获取指定项目的所有配置
func GetGatewayObject(server string, action GetType) (gwObj *[]GatewayObject, err error) {
	url := "http://" + server + "/external/server/"
	switch action {
	case GET_MAPPING:
		url += "GetGatewayMapping"
	case GET_MAPPING_ONLINE:
		url += "GetGatewayMappingOnline"
	case GET_ROUTING:
		url += "GetGatewayRouting"
	case GET_ROUTING_ONLINE:
		url += "GetGatewayRoutingOnline"
	default:
		err = fmt.Errorf("get action doesn't existed")
		return
	}
	gorequest.DisableTransportSwap = true
	apiResult := APIResult{}
	resp, _, errs := gorequest.
		New().
		SetCurlCommand(true).
		Post(url).
		Set("Content-Type", "text/html;charset=UTF-8").
		Type("text").
		Send(`{}`).
		EndStruct(&apiResult)
	if errs != nil {
		err = fmt.Errorf("请求失败: %v", errs)
		return
	}
	defer resp.Body.Close()
	if apiResult.RetCode != 0 {
		err = fmt.Errorf("请求API: %s 失败: %d - %s", url, apiResult.RetCode, apiResult.Exception)
		return
	}
	switch action {
	case GET_MAPPING:
		gwObj = &apiResult.InfoGatewayMappings
	case GET_MAPPING_ONLINE:
		gwObj = &apiResult.InfoGatewayMappingOnlines
	case GET_ROUTING:
		gwObj = &apiResult.InfoGatewayRoutings
	case GET_ROUTING_ONLINE:
		gwObj = &apiResult.InfoGatewayRoutingOnlines
	}
	return
}

// ChangeGatewayObject 修改指定项目的配置
func ChangeGatewayObject(server string, action ChangeType, gwObj *GatewayObject) error {
	url := "http://" + server + "/external/server/"
	switch action {
	case CREATE_MAPPING:
		url += "CreateGatewayMapping"
	case CREATE_MAPPING_ONLINE:
		url += "CreateGatewayMappingOnline"
	case CREATE_ROUTING:
		url += "CreateGatewayRouting"
	case CREATE_ROUTING_ONLINE:
		url += "CreateGatewayRoutingOnline"
	case MODIFY_MAPPING:
		url += "ModifyGatewayMapping"
	case MODIFY_MAPPING_ONLINE:
		url += "ModifyGatewayMappingOnline"
	case MODIFY_ROUTING:
		url += "ModifyGatewayRouting"
	case MODIFY_ROUTING_ONLINE:
		url += "ModifyGatewayRoutingOnline"
	case DELETE_MAPPING:
		url += "DeleteGatewayMapping"
	case DELETE_MAPPING_ONLINE:
		url += "DeleteGatewayMappingOnline"
	case DELETE_ROUTING:
		url += "DeleteGatewayRouting"
	case DELETE_ROUTING_ONLINE:
		url += "DeleteewayRoutingOnline"
	default:
		return fmt.Errorf("change action doesn't existed")
	}
	gorequest.DisableTransportSwap = true
	request := gorequest.
		New().
		Post(url).
		Set("Content-Type", "text/html;charset=UTF-8").
		Type("text").
		Send(gwObj)
	log.Println(request.AsCurlCommand())
	if NotActuallyExecute {
		log.Println(`{retCode: 0, exception: "未实际执行, 仅打印"}`)
		return nil
	}
	apiResult := APIResult{}
	resp, _, errs := request.EndStruct(&apiResult)
	if errs != nil {
		return fmt.Errorf("请求失败: %v", errs)
	}
	defer resp.Body.Close()
	if apiResult.RetCode != 0 {
		return fmt.Errorf("请求API: %s 失败: %d - %s", url, apiResult.RetCode, apiResult.Exception)
	}
	return nil
}

// SyncGatewayObject 同步指定项配置到其他设备
func SyncGatewayObject(srcServer string, dstServers []string, syncType SyncType, objectIgnoredFunc func(gw *GatewayObject) bool, attrIgnoredFunc func(name string) bool) (errs []error) {
	srcGwObjs, err := GetGatewayObject(srcServer, GetType(syncType))
	if err != nil {
		errs = append(errs, err)
		return // 获取初始网关错误, 直接退出
	}
	toNameObjectMap := func(gwObjs *[]GatewayObject) map[string]GatewayObject {
		ret := make(map[string]GatewayObject)
		for _, gwObj := range *gwObjs {
			if objectIgnoredFunc(&gwObj) {
				ret[gwObj["name"].(string)] = gwObj
			}
		}
		return ret
	}

	srcNameObjectMap := toNameObjectMap(srcGwObjs)

	for _, dstServer := range dstServers {
		dstGwObjs, err := GetGatewayObject(dstServer, GetType(syncType))
		if err != nil {
			errs = append(errs, err)
			continue // 获取目标网关错误, 继续下一个
		}
		dstNameObjectMap := toNameObjectMap(dstGwObjs)
		// 先判断增加, 修改
		for name, srcGwObj := range srcNameObjectMap {
			var err error
			if dstGwObj, ok := dstNameObjectMap[name]; !ok {
				if syncType == SYNC_MAPPING {
					err = ChangeGatewayObject(dstServer, CREATE_MAPPING, &srcGwObj)
				} else if syncType == SYNC_ROUTING {
					err = ChangeGatewayObject(dstServer, CREATE_ROUTING, &srcGwObj)
				}
				errs = append(errs, err)
			} else {
				diffObj := srcGwObj.Diff(&dstGwObj, attrIgnoredFunc)
				if len(*diffObj) != 0 {
					if syncType == SYNC_MAPPING {
						err = ChangeGatewayObject(dstServer, MODIFY_MAPPING, diffObj)
					} else if syncType == SYNC_ROUTING {
						err = ChangeGatewayObject(dstServer, MODIFY_ROUTING, diffObj)
					}
					errs = append(errs, err)
				}
			}
		}
		// 再判断需要删除的
		for name := range dstNameObjectMap {
			var err error
			if _, ok := srcNameObjectMap[name]; !ok {
				var diffObj = GatewayObject{"name": name}
				if syncType == SYNC_MAPPING {
					err = ChangeGatewayObject(dstServer, DELETE_MAPPING, &diffObj)
				} else if syncType == SYNC_ROUTING {
					err = ChangeGatewayObject(dstServer, DELETE_ROUTING, &diffObj)
				}
				errs = append(errs, err)
			}
		}
	}
	return nil
}
