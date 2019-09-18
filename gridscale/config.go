package gridscale

import (
	"github.com/gridscale/gsclient-go"
	"log"
)

//Arrays can't be constants in Go, but these will be used as constants
var hardwareProfiles = []string{"default", "legacy", "nested", "cisco_csr", "sophos_utm", "f5_bigip", "q35", "q35_nested"}
var storageTypes = []string{"storage", "storage_high", "storage_insane"}
var availabilityZones = []string{"a", "b", "c"}
var loadbalancerAlgs = []string{"roundrobin", "leastconn"}

const timeLayout = "2006-01-02T15:04:05Z"

//Config config for go gsclient
type Config struct {
	UserUUID string
	APIToken string
	APIUrl   string
}

//Client create a new go gsclient
func (c Config) Client() (*gsclient.Client, error) {
	config := gsclient.NewConfiguration(
		c.APIUrl,
		c.UserUUID,
		c.APIToken,
		true,
		0,
		0,
		0,
	)
	client := gsclient.NewClient(config)
	log.Print("[INFO] gridscale client configured")
	//Make sure the credentials are correct
	_, err := client.GetServerList()
	return client, err
}
