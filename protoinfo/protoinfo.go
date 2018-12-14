package protoinfo

import (
    "sync"
	"fmt"
	"github.com/topfreegames/pitaya/client"
	"github.com/sirupsen/logrus"
)

var instance *client.ProtoBufferInfo
var once sync.Once

func GetProtoInfo (host string, docs string) *client.ProtoBufferInfo {
    once.Do(func() {
    	if instance == nil {
    		cli := client.NewProto(docs, logrus.InfoLevel)
    		err := cli.LoadServoInfo(host)
			if err != nil {
				fmt.Println("Unable to load server documentation.")
				fmt.Println(err)
			} else {
				instance = cli.ExportInformation()
			}
    	}
    })
    return instance
}
