package gridscale

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/gridscale/gsclient-go"
)

func resourceGridscaleServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceGridscaleServerCreate,
		Read:   resourceGridscaleServerRead,
		Delete: resourceGridscaleServerDelete,
		Update: resourceGridscaleServerUpdate,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Description:  "The human-readable name of the object. It supports the full UTF-8 charset, with a maximum of 64 characters",
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"memory": {
				Type:         schema.TypeInt,
				Description:  "The amount of server memory in GB.",
				Required:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"cores": {
				Type:         schema.TypeInt,
				Description:  "The number of server cores.",
				Required:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"location_uuid": {
				Type:        schema.TypeString,
				Description: "Helps to identify which datacenter an object belongs to.",
				Optional:    true,
				ForceNew:    true,
				Default:     "45ed677b-3702-4b36-be2a-a2eab9827950",
			},
			"hardware_profile": {
				Type:        schema.TypeString,
				Description: "The number of server cores.",
				Optional:    true,
				ForceNew:    true,
				Default:     "default",
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					valid := false
					for _, profile := range hardwareProfiles {
						if v.(string) == profile {
							valid = true
							break
						}
					}
					if !valid {
						errors = append(errors, fmt.Errorf("%v is not a valid hardware profile. Valid hardware profiles are: %v", v.(string), strings.Join(hardwareProfiles, ",")))
					}
					return
				},
			},
			"power": {
				Type:        schema.TypeBool,
				Description: "The number of server cores.",
				Optional:    true,
				Computed:    true,
			},
			"current_price": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"auto_recovery": {
				Type:        schema.TypeInt,
				Description: "If the server should be auto-started in case of a failure (default=true).",
				Computed:    true,
			},
			"availability_zone": {
				Type:        schema.TypeString,
				Description: "Defines which Availability-Zone the Server is placed.",
				Optional:    true,
				Computed:    true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					valid := false
					for _, profile := range availabilityZones {
						if v.(string) == profile {
							valid = true
							break
						}
					}
					if !valid {
						errors = append(errors, fmt.Errorf("%v is not a valid availability zone. Valid availability zones are: %v", v.(string), strings.Join(availabilityZones, ",")))
					}
					return
				},
			},
			"console_token": {
				Type:        schema.TypeString,
				Description: "If the server should be auto-started in case of a failure (default=true).",
				Computed:    true,
			},
			"legacy": {
				Type:        schema.TypeBool,
				Description: "Legacy-Hardware emulation instead of virtio hardware. If enabled, hotplugging cores, memory, storage, network, etc. will not work, but the server will most likely run every x86 compatible operating system. This mode comes with a performance penalty, as emulated hardware does not benefit from the virtio driver infrastructure.",
				Computed:    true,
			},
			"usage_in_minutes_memory": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"usage_in_minutes_cores": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"labels": {
				Type:        schema.TypeSet,
				Description: "List of labels.",
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceGridscaleServerRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gsclient.Client)
	server, err := client.GetServer(emptyCtx, d.Id())
	if err != nil {
		if requestError, ok := err.(gsclient.RequestError); ok {
			if requestError.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.Set("name", server.Properties.Name)
	d.Set("memory", server.Properties.Memory)
	d.Set("cores", server.Properties.Cores)
	d.Set("hardware_profile", server.Properties.HardwareProfile)
	d.Set("location_uuid", server.Properties.LocationUUID)
	d.Set("power", server.Properties.Power)
	d.Set("current_price", server.Properties.CurrentPrice)
	d.Set("availability_zone", server.Properties.AvailabilityZone)
	d.Set("auto_recovery", server.Properties.AutoRecovery)
	d.Set("console_token", server.Properties.ConsoleToken)
	d.Set("legacy", server.Properties.Legacy)
	d.Set("usage_in_minutes_memory", server.Properties.UsageInMinutesMemory)
	d.Set("usage_in_minutes_cores", server.Properties.UsageInMinutesCores)

	if err = d.Set("labels", server.Properties.Labels); err != nil {
		return fmt.Errorf("Error setting labels: %v", err)
	}
	return nil
}

func resourceGridscaleServerCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gsclient.Client)

	requestBody := gsclient.ServerCreateRequest{
		Name:            d.Get("name").(string),
		Cores:           d.Get("cores").(int),
		Memory:          d.Get("memory").(int),
		LocationUUID:    d.Get("location_uuid").(string),
		AvailablityZone: d.Get("availability_zone").(string),
		Labels:          convSOStrings(d.Get("labels").(*schema.Set).List()),
	}

	profile := d.Get("hardware_profile").(string)
	if profile == "legacy" {
		requestBody.HardwareProfile = gsclient.LegacyServerHardware
	} else if profile == "nested" {
		requestBody.HardwareProfile = gsclient.NestedServerHardware

	} else if profile == "cisco_csr" {
		requestBody.HardwareProfile = gsclient.CiscoCSRServerHardware

	} else if profile == "sophos_utm" {
		requestBody.HardwareProfile = gsclient.SophosUTMServerHardware

	} else if profile == "f5_bigip" {
		requestBody.HardwareProfile = gsclient.F5BigipServerHardware

	} else if profile == "q35" {
		requestBody.HardwareProfile = gsclient.Q35ServerHardware

	} else if profile == "q35_nested" {
		requestBody.HardwareProfile = gsclient.Q35NestedServerHardware

	} else {
		requestBody.HardwareProfile = gsclient.DefaultServerHardware
	}

	response, err := client.CreateServer(emptyCtx, requestBody)

	if err != nil {
		return fmt.Errorf(
			"Error waiting for server (%s) to be created: %s", requestBody.Name, err)
	}

	d.SetId(response.ServerUUID)

	log.Printf("[DEBUG] The id for %s has been set to: %v", requestBody.Name, response.ServerUUID)

	//Set the power state if needed
	power := d.Get("power").(bool)
	if power {
		client.StartServer(emptyCtx, d.Id())
	}

	return resourceGridscaleServerRead(d, meta)
}

func resourceGridscaleServerDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gsclient.Client)
	id := d.Id()
	err := client.StopServer(emptyCtx, id)
	if err != nil {
		return err
	}
	err = client.DeleteServer(emptyCtx, id)

	return err
}

func resourceGridscaleServerUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gsclient.Client)
	shutdownRequired := false

	var err error

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

	requestBody := gsclient.ServerUpdateRequest{
		Name:            d.Get("name").(string),
		AvailablityZone: d.Get("availability_zone").(string),
		Labels:          convSOStrings(d.Get("labels").(*schema.Set).List()),
		Cores:           d.Get("cores").(int),
		Memory:          d.Get("memory").(int),
	}

	//The ShutdownServer command will check if the server is running and shut it down if it is running, so no extra checks are needed here
	if shutdownRequired {
		err = client.ShutdownServer(emptyCtx, d.Id())
		if err != nil {
			return err
		}
	}

	//Execute the update request
	err = client.UpdateServer(emptyCtx, d.Id(), requestBody)
	if err != nil {
		return err
	}

	// Make sure the server in is the expected power state.
	// The StartServer and ShutdownServer functions do a check to see if the server isn't already running, so we don't need to do that here.
	if d.Get("power").(bool) {
		err = client.StartServer(emptyCtx, d.Id())
	} else {
		err = client.ShutdownServer(emptyCtx, d.Id())
	}
	if err != nil {
		return err
	}

	return resourceGridscaleServerRead(d, meta)

}
