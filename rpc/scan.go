package rpc

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/hpb-project/go-hpb/common"
	"github.com/thedevsaddam/gojsonq/v2"
)

func GetNodeListFromHpbScan(endpoint string) (map[common.Address]Node, error) {
	req, err := http.NewRequest("POST", endpoint+"/HpbScan/node/list",
		strings.NewReader(`{"currentPage": 1, "pageSize": 2000, "nodeType": "hpbnode", "country": ""}`))
	if err != nil {
		return nil, err
	}

	req.Header.Add("content-type", "application/json;charset=UTF-8")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	page := gojsonq.New().FromString(string(body)).Nth(3)
	result := gojsonq.New().FromInterface(page).Find("list")
	if result == nil {
		return nil, errors.New("request GetNodeListFromHpbScan failed")
	}
	nodes := make(map[common.Address]Node)
	for _, v := range result.([]interface{}) {
		node := v.(map[string]interface{})
		nodes[common.HexToAddress(node["nodeAddress"].(string))] = Node{
			NodeName:       node["nodeName"].(string),
			NodeAddress:    common.HexToAddress(node["nodeAddress"].(string)),
			LockAmount:     node["lockAmount"].(float64),
			Country:        node["country"].(string),
			LocationDetail: node["locationDetail"].(string),
		}

	}
	return nodes, err
}
