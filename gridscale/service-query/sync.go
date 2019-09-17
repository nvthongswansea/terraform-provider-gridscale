package service_query

import (
	"errors"
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/nvthongswansea/gsclient-go"
	"time"
)

const (
	delayFetchingStatus  = 500 * time.Millisecond
	timeoutCheckDeletion = 3 * time.Minute
)

const (
	provivisoningStatus = "in-provisioning"
	activeStatus        = "active"
)

type gsService string

const (
	LoadbalancerService gsService = "loadbalancer"
	IPService           gsService = "IP"
	NetworkService      gsService = "network"
	ServerService       gsService = "server"
	SSHKeyService       gsService = "sshkey"
	StorageService      gsService = "storage"
	ISOImageService     gsService = "isoimage"
	PaaSService         gsService = "paas"
	SecurityZoneService gsService = "security"
	SnapshotService     gsService = "snapshot"
)

//RetryUntilResourceStatusIsActive blocks until the object's state is not in provisioning anymore
func RetryUntilResourceStatusIsActive(client *gsclient.Client, service gsService, timeout time.Duration, ids ...string) error {
	return resource.Retry(timeout, func() *resource.RetryError {
		time.Sleep(delayFetchingStatus)
		switch service {
		case LoadbalancerService:
			if len(ids) != 1 {
				return resource.NonRetryableError(errors.New("invalid number of ids"))
			}
			lb, err := client.GetLoadBalancer(ids[0])
			if err != nil {
				return resource.NonRetryableError(fmt.Errorf(
					"Error waiting for loadbalancer (%s) to be fetched: %s", ids[0], err))
			}
			if lb.Properties.Status != activeStatus {
				return resource.RetryableError(fmt.Errorf("Status of loadbalancer %s is not active", ids[0]))
			}
			return nil
		case IPService:
			if len(ids) != 1 {
				return resource.NonRetryableError(errors.New("invalid number of ids"))
			}
			ip, err := client.GetIP(ids[0])
			if err != nil {
				return resource.NonRetryableError(fmt.Errorf(
					"Error waiting for IP (%s) to be fetched: %s", ids[0], err))
			}
			if ip.Properties.Status != activeStatus {
				return resource.RetryableError(fmt.Errorf("Status of IP %s is not active", ids[0]))
			}
			return nil
		case NetworkService:
			if len(ids) != 1 {
				return resource.NonRetryableError(errors.New("invalid number of ids"))
			}
			net, err := client.GetNetwork(ids[0])
			if err != nil {
				return resource.NonRetryableError(fmt.Errorf(
					"Error waiting for network (%s) to be fetched: %s", ids[0], err))
			}
			if net.Properties.Status != activeStatus {
				return resource.RetryableError(fmt.Errorf("Status of network %s is not active", ids[0]))
			}
			return nil
		case ServerService:
			if len(ids) != 1 {
				return resource.NonRetryableError(errors.New("invalid number of ids"))
			}
			server, err := client.GetServer(ids[0])
			if err != nil {
				return resource.NonRetryableError(fmt.Errorf(
					"Error waiting for server (%s) to be fetched: %s", ids[0], err))
			}
			if server.Properties.Status != activeStatus {
				return resource.RetryableError(fmt.Errorf("Status of server %s is not active", ids[0]))
			}
			return nil
		case SSHKeyService:
			if len(ids) != 1 {
				return resource.NonRetryableError(errors.New("invalid number of ids"))
			}
			sshKey, err := client.GetSshkey(ids[0])
			if err != nil {
				return resource.NonRetryableError(fmt.Errorf(
					"Error waiting for sshKey (%s) to be fetched: %s", ids[0], err))
			}
			if sshKey.Properties.Status != activeStatus {
				return resource.RetryableError(fmt.Errorf("Status of sshKey %s is not active", ids[0]))
			}
			return nil
		case StorageService:
			if len(ids) != 1 {
				return resource.NonRetryableError(errors.New("invalid number of ids"))
			}
			storage, err := client.GetStorage(ids[0])
			if err != nil {
				return resource.NonRetryableError(fmt.Errorf(
					"Error waiting for storage (%s) to be fetched: %s", ids[0], err))
			}
			if storage.Properties.Status != activeStatus {
				return resource.RetryableError(fmt.Errorf("Status of storage %s is not active", ids[0]))
			}
			return nil
		case PaaSService:
			if len(ids) != 1 {
				return resource.NonRetryableError(errors.New("invalid number of ids"))
			}
			paas, err := client.GetPaaSService(ids[0])
			if err != nil {
				return resource.NonRetryableError(fmt.Errorf(
					"Error waiting for PaaS service (%s) to be fetched: %s", ids[0], err))
			}
			if paas.Properties.Status != activeStatus {
				return resource.RetryableError(fmt.Errorf("Status of PaaS service %s is not active", ids[0]))
			}
			return nil
		case SecurityZoneService:
			if len(ids) != 1 {
				return resource.NonRetryableError(errors.New("invalid number of ids"))
			}
			securityZone, err := client.GetPaaSSecurityZone(ids[0])
			if err != nil {
				return resource.NonRetryableError(fmt.Errorf(
					"Error waiting for security zone (%s) to be fetched: %s", ids[0], err))
			}
			if securityZone.Properties.Status != activeStatus {
				return resource.RetryableError(fmt.Errorf("Status of security zone %s is not active", ids[0]))
			}
			return nil
		case SnapshotService:
			if len(ids) != 2 {
				return resource.NonRetryableError(errors.New("invalid number of ids"))
			}
			snapshot, err := client.GetStorageSnapshot(ids[0], ids[1])
			if err != nil {
				return resource.NonRetryableError(fmt.Errorf(
					"Error waiting for snapshot (%s) to be fetched: %s", ids[1], err))
			}
			if snapshot.Properties.Status != activeStatus {
				return resource.RetryableError(fmt.Errorf("Status of snapshot %s is not active", ids[1]))
			}
			return nil
		default:
			return resource.NonRetryableError(fmt.Errorf("invalid service"))
		}
	})
}

//RetryUntilDeleted blocks until an object is deleted successfully
func RetryUntilDeleted(client *gsclient.Client, service gsService, timeout time.Duration, ids ...string) error {
	return resource.Retry(timeout, func() *resource.RetryError {
		var err error
		switch service {
		case LoadbalancerService:
			if len(ids) != 1 {
				return resource.NonRetryableError(errors.New("invalid number of ids"))
			}
			_, err = client.GetLoadBalancer(ids[0])
		case IPService:
			if len(ids) != 1 {
				return resource.NonRetryableError(errors.New("invalid number of ids"))
			}
			_, err = client.GetIP(ids[0])
		case NetworkService:
			if len(ids) != 1 {
				return resource.NonRetryableError(errors.New("invalid number of ids"))
			}
			_, err = client.GetNetwork(ids[0])
		case ServerService:
			if len(ids) != 1 {
				return resource.NonRetryableError(errors.New("invalid number of ids"))
			}
			_, err = client.GetServer(ids[0])
		case SSHKeyService:
			if len(ids) != 1 {
				return resource.NonRetryableError(errors.New("invalid number of ids"))
			}
			_, err = client.GetSshkey(ids[0])
		case StorageService:
			if len(ids) != 1 {
				return resource.NonRetryableError(errors.New("invalid number of ids"))
			}
			_, err = client.GetStorage(ids[0])
		case PaaSService:
			if len(ids) != 1 {
				return resource.NonRetryableError(errors.New("invalid number of ids"))
			}
			_, err = client.GetPaaSService(ids[0])
		case SecurityZoneService:
			if len(ids) != 1 {
				return resource.NonRetryableError(errors.New("invalid number of ids"))
			}
			_, err = client.GetPaaSSecurityZone(ids[0])
		case SnapshotService:
			if len(ids) != 2 {
				return resource.NonRetryableError(errors.New("invalid number of ids"))
			}
			_, err = client.GetStorageSnapshot(ids[0], ids[1])
		default:
			return resource.NonRetryableError(fmt.Errorf("invalid service"))
		}
		if err != nil {
			if requestError, ok := err.(gsclient.RequestError); ok {
				if requestError.StatusCode == 404 {
					return nil
				}
			}
			return resource.NonRetryableError(err)
		}
		return resource.RetryableError(fmt.Errorf("%v (%v) still exists", service, ids))
	})
}
