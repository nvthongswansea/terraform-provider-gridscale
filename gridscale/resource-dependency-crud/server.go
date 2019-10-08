package resource_dependency_crud

import (
	"context"
	"fmt"
	"github.com/gridscale/gsclient-go"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

//ServerDependencyClient is an wrapper of gsclient which is used for
//CRUD dependency of a server in gridscale terraform provider
type ServerDependencyClient struct {
	gsc  *gsclient.Client
	data *schema.ResourceData
}

//NewServerDepClient creates a new instance DependencyClient
func NewServerDepClient(gsc *gsclient.Client, d *schema.ResourceData) *ServerDependencyClient {
	return &ServerDependencyClient{gsc, d}
}

//GetGSClient returns gsclient
func (c ServerDependencyClient) GetGSClient() *gsclient.Client {
	return c.gsc
}

//SetData sets data property
func (c *ServerDependencyClient) SetData(data *schema.ResourceData) {
	c.data = data
}

//GetData returns data
func (c ServerDependencyClient) GetData() *schema.ResourceData {
	return c.data
}

//LinkStorages links a boot storage to a server
func (c *ServerDependencyClient) LinkStorages(ctx context.Context) error {
	d := c.GetData()
	client := c.GetGSClient()
	//Bootable storage has to be attached first
	if attr, ok := d.GetOk("storage"); ok {
		for _, value := range attr.(*schema.Set).List() {
			storage := value.(map[string]interface{})
			err := client.LinkStorage(ctx, d.Id(), storage["object_uuid"].(string), storage["bootdevice"].(bool))
			if err != nil {
				return fmt.Errorf(
					"Error waiting for storage (%s) to be attached to server (%s): %s",
					storage["object_uuid"].(string),
					d.Id(),
					err,
				)
			}
		}
	}
	return nil
}

//LinkIPv4 links IPv4 address to a server
func (c *ServerDependencyClient) LinkIPv4(ctx context.Context) error {
	d := c.GetData()
	client := c.GetGSClient()
	if attr, ok := d.GetOk("ipv4"); ok {
		if client.GetIPVersion(ctx, attr.(string)) != 4 {
			return fmt.Errorf("The IP address with UUID %v is not version 4", attr.(string))
		}
		err := client.LinkIP(ctx, d.Id(), attr.(string))
		if err != nil {
			return fmt.Errorf(
				"Error waiting for IP address (%s) to be attached to server (%s): %s",
				attr.(string),
				d.Id(),
				err,
			)
		}
	}
	return nil
}

//LinkIPv6 link an IPv6 address to a server
func (c *ServerDependencyClient) LinkIPv6(ctx context.Context) error {
	d := c.GetData()
	client := c.GetGSClient()
	if attr, ok := d.GetOk("ipv6"); ok {
		if client.GetIPVersion(ctx, attr.(string)) != 6 {
			return fmt.Errorf("The IP address with UUID %v is not version 6", attr.(string))
		}
		err := client.LinkIP(ctx, d.Id(), attr.(string))
		if err != nil {
			return fmt.Errorf(
				"Error waiting for IP address (%s) to be attached to server (%s): %s",
				attr.(string),
				d.Id(),
				err,
			)
		}
	}
	return nil
}

//LinkISOImage links an ISO-image to a server
func (c *ServerDependencyClient) LinkISOImage(ctx context.Context) error {
	d := c.GetData()
	client := c.GetGSClient()
	if attr, ok := d.GetOk("isoimage"); ok {
		err := client.LinkIsoImage(ctx, d.Id(), attr.(string))
		if err != nil {
			return fmt.Errorf(
				"Error waiting for iso-image (%s) to be attached to server (%s): %s",
				attr.(string),
				d.Id(),
				err,
			)
		}
	}
	return nil
}

//LinkNetworks links networks to server
func (c *ServerDependencyClient) LinkNetworks(ctx context.Context, isPublic bool) error {
	d := c.GetData()
	client := c.GetGSClient()
	if isPublic {
		publicNetwork, err := client.GetNetworkPublic(ctx)
		if err != nil {
			return err
		}
		err = client.LinkNetwork(
			ctx,
			d.Id(),
			publicNetwork.Properties.ObjectUUID,
			"",
			false,
			0,
			nil,
			&gsclient.FirewallRules{},
		)
		if err != nil {
			return fmt.Errorf(
				"Error waiting for public network (%s) to be attached to server (%s): %s",
				publicNetwork.Properties.ObjectUUID,
				d.Id(),
				err,
			)
		}
		return nil
	}
	if attr, ok := d.GetOk("network"); ok {
		for i, value := range attr.([]interface{}) {
			network := value.(map[string]interface{})
			var fwRules gsclient.FirewallRules
			var rulesv4 []gsclient.FirewallRuleProperties
			if attr1, ok1 := d.GetOk(fmt.Sprintf("%s.%d.rules_v4_in", "network", i)); ok1 {
				for _, value1 := range attr1.([]interface{}) {
					rule := value1.(map[string]interface{})
					rulesv4 = append(rulesv4, gsclient.FirewallRuleProperties{
						Protocol: rule["protocol"].(string),
						DstPort:  rule["dst_port"].(string),
						SrcPort:  rule["src_port"].(string),
						SrcCidr:  rule["src_cidr"].(string),
						Action:   rule["action"].(string),
						Comment:  rule["comment"].(string),
						DstCidr:  rule["dst_cidr"].(string),
						Order:    rule["order"].(int),
					})
				}
			}
			fwRules.RulesV4In = rulesv4
			err := client.LinkNetwork(
				ctx,
				d.Id(),
				network["object_uuid"].(string),
				"",
				network["bootdevice"].(bool),
				0,
				nil,
				&fwRules,
			)
			if err != nil {
				return fmt.Errorf(
					"Error waiting for network (%s) to be attached to server (%s): %s",
					network["object_uuid"],
					d.Id(),
					err,
				)
			}
		}
	}
	return nil
}

//IsShutdownRequired checks if server is needed to be shutdown when updating
func (c *ServerDependencyClient) IsShutdownRequired(ctx context.Context) bool {
	var shutdownRequired bool
	d := c.GetData()
	if d.HasChange("cores") {
		old, new := d.GetChange("cores")
		if new.(int) < old.(int) || d.Get("legacy").(bool) { //Legacy systems don't support updating the memory while running
			shutdownRequired = true
		}
	}
	if d.HasChange("memory") {
		old, new := d.GetChange("memory")
		if new.(int) < old.(int) || d.Get("legacy").(bool) { //Legacy systems don't support updating the memory while running
			shutdownRequired = true
		}
	}
	if d.HasChange("ipv4") || d.HasChange("ipv6") || d.HasChange("storage") || d.HasChange("network") {
		shutdownRequired = true
	}
	return shutdownRequired
}

//UpdateISOImageRel updates relationship between a server and an ISO-image
func (c *ServerDependencyClient) UpdateISOImageRel(ctx context.Context) error {
	d := c.GetData()
	client := c.GetGSClient()
	var err error
	if d.HasChange("isoimage") {
		oldIso, newIso := d.GetChange("isoimage")
		if newIso == "" {
			err = client.UnlinkIsoImage(ctx, d.Id(), oldIso.(string))

		} else {
			err = client.LinkIsoImage(ctx, d.Id(), newIso.(string))
		}
		if err != nil {
			return err
		}
	}
	return nil
}

//UpdateIPv4Rel updates relationship between a server and an IPv4 address
func (c *ServerDependencyClient) UpdateIPv4Rel(ctx context.Context) (bool, error) {
	needsPublicNetwork := true
	d := c.GetData()
	client := c.GetGSClient()
	var err error
	if d.HasChange("ipv4") {
		oldIp, newIp := d.GetChange("ipv4")
		if newIp == "" {
			err = client.UnlinkIP(ctx, d.Id(), oldIp.(string))
		} else {
			err = client.LinkIP(ctx, d.Id(), newIp.(string))
		}
		if err != nil {
			return needsPublicNetwork, err
		}
		if oldIp != "" {
			needsPublicNetwork = false
		}
	}
	return needsPublicNetwork, err
}

//UpdateIPv6Rel updates realtionship between a server and an IPv6 address
func (c *ServerDependencyClient) UpdateIPv6Rel(ctx context.Context) (bool, error) {
	needsPublicNetwork := true
	d := c.GetData()
	client := c.GetGSClient()
	var err error
	if d.HasChange("ipv6") {
		oldIp, newIp := d.GetChange("ipv6")
		if newIp == "" {
			err = client.UnlinkIP(ctx, d.Id(), oldIp.(string))
		} else {
			err = client.LinkIP(ctx, d.Id(), newIp.(string))
		}
		if err != nil {
			return needsPublicNetwork, err
		}
		if oldIp != "" {
			needsPublicNetwork = false
		}
	}
	return needsPublicNetwork, err
}

//UpdatePublicNetworkRel updates relationship between a server and a oublic network
func (c *ServerDependencyClient) UpdatePublicNetworkRel(ctx context.Context, isToLink bool) error {
	d := c.GetData()
	client := c.GetGSClient()
	publicNetwork, err := client.GetNetworkPublic(ctx)
	if err != nil {
		return err
	}
	if isToLink {
		err = client.LinkNetwork(
			ctx,
			d.Id(),
			publicNetwork.Properties.ObjectUUID,
			"",
			false,
			0,
			[]string{},
			&gsclient.FirewallRules{},
		)
		if err != nil {
			return err
		}
	} else {
		err = client.UnlinkNetwork(ctx, d.Id(), publicNetwork.Properties.ObjectUUID)
		if err != nil {
			return err
		}
	}
	return nil
}

//UpdateOtherNetworkRel updates relationship between a server and networks
func (c *ServerDependencyClient) UpdateOtherNetworkRel(ctx context.Context) error {
	d := c.GetData()
	client := c.GetGSClient()
	var err error
	//It currently unlinks and relinks all networks if any network has changed. This could probably be done better, but this way is easy and works well
	if d.HasChange("network") {
		oldNetworks, newNetworks := d.GetChange("network")
		for _, value := range oldNetworks.(*schema.Set).List() {
			network := value.(map[string]interface{})
			if network["object_uuid"].(string) != "" {
				err = client.UnlinkNetwork(ctx, d.Id(), network["object_uuid"].(string))
				if err != nil {
					return err
				}
			}
		}
		for _, value := range newNetworks.(*schema.Set).List() {
			network := value.(map[string]interface{})
			if network["object_uuid"].(string) != "" {
				err = client.LinkNetwork(
					ctx,
					d.Id(),
					network["object_uuid"].(string),
					"", network["bootdevice"].(bool),
					0, []string{},
					&gsclient.FirewallRules{},
				)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

//UpdateStorageRel updates relationship between a server and storages
func (c *ServerDependencyClient) UpdateStorageRel(ctx context.Context) error {
	d := c.GetData()
	client := c.GetGSClient()
	var err error
	if d.HasChange("storage") {
		oldStorages, newStorages := d.GetChange("storage")
		//unlink old storages if needed
		for _, value := range oldStorages.(*schema.Set).List() {
			oldStorage := value.(map[string]interface{})
			unlink := true
			for _, value := range newStorages.(*schema.Set).List() {
				newStorage := value.(map[string]interface{})
				if oldStorage["object_uuid"].(string) == newStorage["object_uuid"].(string) {
					unlink = false
					break
				}
			}
			if unlink {
				err = client.UnlinkStorage(ctx, d.Id(), oldStorage["object_uuid"].(string))
				if err != nil {
					return err
				}
			}
		}

		//link new storages if needed
		for _, value := range newStorages.(*schema.Set).List() {
			newStorage := value.(map[string]interface{})
			link := true
			for _, value := range oldStorages.(*schema.Set).List() {
				oldStorage := value.(map[string]interface{})
				if oldStorage["object_uuid"].(string) == newStorage["object_uuid"].(string) {
					link = false
					break
				}
			}
			if link {
				err = client.LinkStorage(ctx, d.Id(), newStorage["object_uuid"].(string), newStorage["bootdevice"].(bool))
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
