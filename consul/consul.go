package consul

import . "github.com/tj/go-debug"
import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/vicanso/gorequest"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type serviceNode struct {
	name   string
	ip     string
	port   int
	weight int
}

var debug = Debug("jt.haproxy-agent")

/**
 * [get description]
 * @param  {[type]} key string        [description]
 * @return {[type]}     [description]
 */
func get(key string) string {
	var result = ""
	switch key {
	case "consul":
		result = os.Getenv("CONSUL")
		if "" == result {
			// result = "http://localhost:8500"
			// result = "http://test-docker:8500"
			result = "http://black:8500"
		}
	case "serviceTag":
		result = os.Getenv("SERVICE_TAG")
		if "" == result {
			// result = "haproxy"
			result = "ytj-backend-product"
		}
	case "backendTag":
		result = os.Getenv("BACKEND_TAG")
		if "" == result {
			result = "varnish"
		}
	}
	debug("get %s:%s", key, result)
	return result
}

/**
 * [HttpBackends description]
 */
func HttpBackends() {
	url := get("consul") + "/v1/catalog/services"
	request := gorequest.New().Timeout(10 * 1000 * time.Millisecond)
	_, body, err := request.Get(url).EndBytes()
	if err != nil {
		fmt.Println(err)
		return
	}

	debug("service list:%s", body)
	json, _ := simplejson.NewJson(body)
	data, _ := json.Map()
	// nodes = make([]serviceNode, len(arr))
	services := make([]string, 0)
	backendTagArr := strings.Split(get("backendTag"), ",")
	for name, _ := range data {
		for _, backendTag := range backendTagArr {
			tags, _ := json.GetPath(name).Array()
			index := indexOf(backendTag, tags)
			if index == -1 {
				break
			}
			index = indexOf(name, services)
			if index == -1 {
				services = append(services, name)
			}
		}
	}
	for _, name := range services {
		nodes := getService(name)
		debug("ndoes:%s", nodes)
	}
}

/**
 * [indexOf description]
 * @param  {[type]} params ...interface{} [description]
 * @return {[type]}        [description]
 */
func indexOf(params ...interface{}) int {
	v := reflect.ValueOf(params[0])
	arr := reflect.ValueOf(params[1])

	var t = reflect.TypeOf(params[1]).Kind()

	if t != reflect.Slice && t != reflect.Array {
		panic("Type Error! Second argument must be an array or a slice.")
	}

	for i := 0; i < arr.Len()-1; i++ {
		if arr.Index(i).Interface() == v.Interface() {
			return i
		}
	}
	return -1
}

/**
 * [getService description]
 * @param  {[type]} service string        [description]
 * @return {[type]}         [description]
 */
func getService(service string) (nodes []*simplejson.Json) {
	url := get("consul") + "/v1/catalog/service/" + service
	debug("get service url:%s", url)
	request := gorequest.New().Timeout(2000 * time.Millisecond)
	_, body, err := request.Get(url).EndBytes()
	if err != nil {
		panic("get service from consul fail")
	}

	debug("get service:%s", body)
	json, _ := simplejson.NewJson(body)

	arr, _ := json.Array()
	nodes = make([]*simplejson.Json, len(arr))
	for i, _ := range arr {
		tmp := json.GetIndex(i)
		weightKey := "weight:"
		weight := 1
		for _, tag := range tmp.Get("ServiceTags").MustStringArray() {
			if strings.Index(tag, weightKey) == 0 {
				tmpWeight, _ := strconv.Atoi(tag[len(weightKey):])
				weight = tmpWeight
			}
		}

		t := simplejson.New()
		t.Set("name", tmp.Get("ServiceName").MustString())
		t.Set("ip", tmp.Get("ServiceAddress").MustString())
		t.Set("port", tmp.Get("ServicePort").MustInt())
		t.Set("weight", weight)
		nodes[i] = t
	}
	return
}
