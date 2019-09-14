package gridscale

import (
	"github.com/nvthongswansea/gsclient-go"
	"log"
)

//Arrays can't be constants in Go, but these will be used as constants
var hardwareProfiles = []string{"default", "legacy", "nested", "cisco_csr", "sophos_utm", "f5_bigip", "q35"}
var storageTypes = []string{"storage", "storage_high", "storage_insane"}
var availabilityZones = []string{"a", "b", "c"}

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
		false,
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
