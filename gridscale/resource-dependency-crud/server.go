package resource_dependency_crud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/nvthongswansea/gsclient-go"
)

var firewallRuleType = []string{"rules_v4_in", "rules_v4_out", "rules_v6_in", "rules_v6_out"}

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
	if attr, ok := d.GetOk("storage"); ok {
		for _, value := range attr.([]interface{}) {
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
		//Check IP version
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
		//Check IP version
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
func (c *ServerDependencyClient) LinkNetworks(ctx context.Context) error {
	d := c.GetData()
	client := c.GetGSClient()
	if attrNetRel, ok := d.GetOk("network"); ok {
		for _, value := range attrNetRel.(*schema.Set).List() {
			network := value.(map[string]interface{})
			//Read custom firewall rules from `network` property (field)
			customFwRules := readCustomFirewallRules(network)
			err := client.LinkNetwork(
				ctx,
				d.Id(),
				network["object_uuid"].(string),
				"",
				network["bootdevice"].(bool),
				0,
				nil,
				&customFwRules,
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

//readCustomFirewallRules reads custom firewall rules from a specific network
//returns `gsclient.FirewallRules` type variable
func readCustomFirewallRules(netData map[string]interface{}) gsclient.FirewallRules {
	//Init firewall rule variable
	var fwRules gsclient.FirewallRules

	//Loop through all firewall rule types
	//there are 4 types: "rules_v4_in", "rules_v4_out", "rules_v6_in", "rules_v6_out".
	for _, ruleType := range firewallRuleType {
		//Init array of firewall rules
		var rules []gsclient.FirewallRuleProperties
		//Check if the firewall rule type is declared in the current network
		if rulesInTypeAttr, ok := netData[ruleType]; ok {
			//Loop through all rules in the current firewall type
			for _, rulesInType := range rulesInTypeAttr.([]interface{}) {
				ruleProps := rulesInType.(map[string]interface{})
				ruleProperties := gsclient.FirewallRuleProperties{
					Protocol: ruleProps["protocol"].(string),
					DstPort:  ruleProps["dst_port"].(string),
					SrcPort:  ruleProps["src_port"].(string),
					SrcCidr:  ruleProps["src_cidr"].(string),
					Action:   ruleProps["action"].(string),
					Comment:  ruleProps["comment"].(string),
					DstCidr:  ruleProps["dst_cidr"].(string),
					Order:    ruleProps["order"].(int),
				}
				//Add rule to the array of rules
				rules = append(rules, ruleProperties)
			}
		}

		//Based on rule type to place the rules in the right property of fwRules variable
		if ruleType == "rules_v4_in" {
			fwRules.RulesV4In = rules
		} else if ruleType == "rules_v4_out" {
			fwRules.RulesV4Out = rules
		} else if ruleType == "rules_v6_in" {
			fwRules.RulesV6In = rules
		} else if ruleType == "rules_v6_out" {
			fwRules.RulesV6Out = rules
		}
	}
	return fwRules
}

//IsShutdownRequired checks if server is needed to be shutdown when updating
func (c *ServerDependencyClient) IsShutdownRequired(ctx context.Context) bool {
	var shutdownRequired bool
	d := c.GetData()
	//If the number of cores is decreased, shutdown the server
	if d.HasChange("cores") {
		old, new := d.GetChange("cores")
		if new.(int) < old.(int) || d.Get("legacy").(bool) { //Legacy systems don't support updating the memory while running
			shutdownRequired = true
		}
	}
	//If the amount of memory is decreased, shutdown the server
	if d.HasChange("memory") {
		old, new := d.GetChange("memory")
		if new.(int) < old.(int) || d.Get("legacy").(bool) { //Legacy systems don't support updating the memory while running
			shutdownRequired = true
		}
	}
	//If IP address, storages, or networks are changed, shutdown the server
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
	//Check if ISO-image field is changed
	if d.HasChange("isoimage") {
		oldIso, _ := d.GetChange("isoimage")
		//If there is an ISO-image already linked to the server
		//Unlink it
		if oldIso != "" {
			err = client.UnlinkIsoImage(ctx, d.Id(), oldIso.(string))
			if err != nil {
				if requestError, ok := err.(gsclient.RequestError); ok {
					//If 404, that means ISO-image is already deleted => the relation between ISO-image and server is deleted automatically
					if requestError.StatusCode != 404 {
						return fmt.Errorf(
							"Error waiting for ISO-image (%s) to be detached from server (%s): %s",
							oldIso,
							d.Id(),
							err,
						)
					}
				} else {
					return err
				}
			}
		}
		//Link new ISO-image (if there is one)
		err = c.LinkISOImage(ctx)
	}
	return err
}

//UpdateIPv4Rel updates relationship between a server and an IPv4 address
func (c *ServerDependencyClient) UpdateIPv4Rel(ctx context.Context) error {
	d := c.GetData()
	client := c.GetGSClient()
	var err error
	//If IPv4 field is changed
	if d.HasChange("ipv4") {
		oldIp, _ := d.GetChange("ipv4")
		//If there is an IPv4 address already linked to the server
		//Unlink it
		if oldIp != "" {
			err = client.UnlinkIP(ctx, d.Id(), oldIp.(string))
			if err != nil {
				if requestError, ok := err.(gsclient.RequestError); ok {
					//If 404, that means IP is already deleted => the relation between IP and server is deleted automatically
					if requestError.StatusCode != 404 {
						return fmt.Errorf(
							"error waiting for IPv4 (%s) to be detached from server (%s): %s",
							oldIp,
							d.Id(),
							err,
						)
					}
				} else {
					return err
				}
			}
		}
		//Link new IPv4 (if there is one)
		err = c.LinkIPv4(ctx)
	}
	return err
}

//UpdateIPv6Rel updates relationship between a server and an IPv6 address
func (c *ServerDependencyClient) UpdateIPv6Rel(ctx context.Context) error {
	d := c.GetData()
	client := c.GetGSClient()
	var err error
	if d.HasChange("ipv6") {
		oldIp, _ := d.GetChange("ipv6")
		//If there is an IPv6 address already linked to the server
		//Unlink it
		if oldIp != "" {
			err = client.UnlinkIP(ctx, d.Id(), oldIp.(string))
			if err != nil {
				if requestError, ok := err.(gsclient.RequestError); ok {
					//If 404, that means IP is already deleted => the relation between IP and server is deleted automatically
					if requestError.StatusCode != 404 {
						return fmt.Errorf(
							"Error waiting for IPv6 (%s) to be detached from server (%s): %s",
							oldIp,
							d.Id(),
							err,
						)
					}
				} else {
					return err
				}
			}
		}
		//Link new IPv6 (if there is one)
		err = c.LinkIPv6(ctx)
	}
	return err
}

//UpdateNetworksRel updates relationship between a server and networks
func (c *ServerDependencyClient) UpdateNetworksRel(ctx context.Context) error {
	d := c.GetData()
	client := c.GetGSClient()
	var err error
	if d.HasChange("network") {
		oldNetworks, _ := d.GetChange("network")
		//Unlink all old networks if there are any networks linked to the server
		for _, value := range oldNetworks.(*schema.Set).List() {
			network := value.(map[string]interface{})
			if network["object_uuid"].(string) != "" {
				err = client.UnlinkNetwork(ctx, d.Id(), network["object_uuid"].(string))
				if err != nil {
					if requestError, ok := err.(gsclient.RequestError); ok {
						//If 404, that means network is already deleted => the relation between network and server is deleted automatically
						if requestError.StatusCode != 404 {
							return fmt.Errorf(
								"Error waiting for network (%s) to be detached from server (%s): %s",
								network["object_uuid"].(string),
								d.Id(),
								err,
							)
						}
					} else {
						return err
					}
				}
			}
		}
		//Links all new networks (if there are some)
		err = c.LinkNetworks(ctx)
	}
	return err
}

//UpdateStoragesRel updates relationship between a server and storages
func (c *ServerDependencyClient) UpdateStoragesRel(ctx context.Context) error {
	d := c.GetData()
	client := c.GetGSClient()
	var err error
	if d.HasChange("storage") {
		oldStorages, _ := d.GetChange("storage")
		for _, value := range oldStorages.([]interface{}) {
			storage := value.(map[string]interface{})
			if storage["object_uuid"].(string) != "" {
				err = client.UnlinkStorage(ctx, d.Id(), storage["object_uuid"].(string))
				if err != nil {
					if requestError, ok := err.(gsclient.RequestError); ok {
						//If 404, that means storage is already deleted => the relation between storage and server is deleted automatically
						if requestError.StatusCode != 404 {
							return fmt.Errorf(
								"Error waiting for storage (%s) to be detached from server (%s): %s",
								storage["object_uuid"].(string),
								d.Id(),
								err,
							)
						}
					} else {
						return err
					}
				}
			}
		}
		//Links all new storages (if there are some)
		err = c.LinkStorages(ctx)
	}
	return err
}
