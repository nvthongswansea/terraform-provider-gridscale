package gridscale

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/nvthongswansea/gsclient-go"
	service_query "github.com/terraform-providers/terraform-provider-gridscale/gridscale/service-query"
	"log"
	"time"
)

func resourceGridscaleSecurityZone() *schema.Resource {
	return &schema.Resource{
		Create: resourceGridscaleSecurityZoneCreate,
		Read:   resourceGridscaleSecurityZoneRead,
		Delete: resourceGridscaleSecurityZoneDelete,
		Update: resourceGridscaleSecurityZoneUpdate,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The human-readable name of the object",
				ValidateFunc: validation.NoZeroValues,
			},
			"location_uuid": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Helps to identify which datacenter an object belongs to",
			},
			"location_country": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The human-readable name of the location",
			},
			"location_iata": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Uses IATA airport code, which works as a location identifier",
			},
			"location_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The human-readable name of the location",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Defines the date and time the object was initially created",
			},
			"change_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Defines the date and time of the last object change",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status indicates the status of the object",
			},
			"labels": {
				Type:        schema.TypeList,
				Description: "List of labels.",
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"relations": {
				Type:        schema.TypeList,
				Description: "List of PaaS services' UUIDs relating to the security zone",
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(time.Minute * 3),
			Create: schema.DefaultTimeout(time.Minute * 3),
			Update: schema.DefaultTimeout(time.Minute * 3),
		},
	}
}

func resourceGridscaleSecurityZoneRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gsclient.Client)
	secZone, err := client.GetPaaSSecurityZone(d.Id())
	if err != nil {
		if requestError, ok := err.(gsclient.RequestError); ok {
			if requestError.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}
	props := secZone.Properties
	d.Set("name", props.Name)
	d.Set("location_uuid", props.LocationUUID)
	d.Set("location_country", props.LocationCountry)
	d.Set("location_iata", props.LocationIata)
	d.Set("location_name", props.LocationName)
	d.Set("create_time", props.CreateTime)
	d.Set("change_time", props.ChangeTime)
	d.Set("status", props.Status)

	//Set labels
	if err = d.Set("labels", props.Labels); err != nil {
		return fmt.Errorf("Error setting labels: %v", err)
	}

	//Set relations
	var rels []string
	for _, val := range props.Relation.Services {
		rels = append(rels, val.ObjectUUID)
	}
	if err = d.Set("relations", rels); err != nil {
		return fmt.Errorf("Error setting relations: %v", err)
	}
	return nil
}

func resourceGridscaleSecurityZoneCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gsclient.Client)
	requestBody := gsclient.PaaSSecurityZoneCreateRequest{
		Name:         d.Get("name").(string),
		LocationUUID: d.Get("location_uuid").(string),
	}
	response, err := client.CreatePaaSSecurityZone(requestBody)
	if err != nil {
		return err
	}
	d.SetId(response.ObjectUUID)
	log.Printf("The id for security zone %s has been set to %v", requestBody.Name, response.ObjectUUID)
	return resourceGridscaleSecurityZoneRead(d, meta)
}

func resourceGridscaleSecurityZoneUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gsclient.Client)
	requestBody := gsclient.PaaSSecurityZoneUpdateRequest{
		Name:         d.Get("name").(string),
		LocationUUID: d.Get("location_uuid").(string),
	}
	err := client.UpdatePaaSSecurityZone(d.Id(), requestBody)
	if err != nil {
		return err
	}
	err = service_query.RetryUntilResourceStatusIsActive(client, service_query.SecurityZoneService, d.Id(), d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return err
	}
	return resourceGridscaleSecurityZoneRead(d, meta)
}

func resourceGridscaleSecurityZoneDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gsclient.Client)
	d.ConnInfo()
	err := resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		return resource.RetryableError(client.DeletePaaSSecurityZone(d.Id()))
	})
	if err != nil {
		return err
	}
	return service_query.RetryUntilDeleted(client, service_query.SecurityZoneService, d.Id())
}
