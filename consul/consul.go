package consul

import . "github.com/tj/go-debug"
import (
  "os"
  "time"
  "strings"
  "strconv"
  "github.com/vicanso/gorequest"
  "github.com/bitly/go-simplejson"
)

type serviceNode struct {
  name string
  ip string
  port int
  weight int
}



var debug = Debug("jt.haproxy-agent")

func get(key string) string {
  var result = ""
  switch key {
  case "consul":
    result = os.Getenv("CONSUL")
    if "" == result {
      // result = "http://localhost:8500"
      result = "http://test-docker:8500"
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


func GetService()  {
  service := get("serviceTag")
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
  result := make([]serviceNode, len(arr))
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
    name := tmp.Get("ServiceName").MustString()
    ip := tmp.Get("ServiceAddress").MustString()
    port := tmp.Get("ServicePort").MustInt()
    node := serviceNode{name : name, ip : ip, port : port, weight : weight}
    result[i] = node
  }
  debug("result%s", result)
}
