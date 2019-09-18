package gridscale

import (
	"fmt"
	"github.com/gridscale/gsclient-go"
	"github.com/hashicorp/terraform/helper/schema"
	service_query "github.com/terraform-providers/terraform-provider-gridscale/gridscale/service-query"
	"log"
	"time"
)

func resourceGridscaleStorageSnapshot() *schema.Resource {
	return &schema.Resource{
		Create: resourceGridscaleSnapshotCreate,
		Read:   resourceGridscaleSnapshotRead,
		Delete: resourceGridscaleSnapshotDelete,
		Update: resourceGridscaleSnapshotUpdate,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The human-readable name of the object",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status indicates the status of the object",
			},
			"location_country": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The human-readable name of the location",
			},
			"location_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The human-readable name of the location",
			},
			"location_iata": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Uses IATA airport code, which works as a location identifier",
			},
			"location_uuid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Helps to identify which datacenter an object belongs to",
			},
			"usage_in_minutes": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Total minutes the object has been running",
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
			"license_product_no": {
				Type:     schema.TypeString,
				Computed: true,
				Description: `If a template has been used that requires a license key (e.g. Windows Servers) this shows 
the product_no of the license (see the /prices endpoint for more details)`,
			},
			"current_price": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "The price for the current period since the last bill",
			},
			"capacity": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The capacity of a storage/ISO-Image/template/snapshot in GB",
			},
			"storage_uuid": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Uuid of the storage used to create this snapshot",
			},
			"labels": {
				Type:        schema.TypeList,
				Description: "List of labels.",
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(time.Minute * 3),
			Update: schema.DefaultTimeout(time.Minute * 3),
		},
	}
}

func resourceGridscaleSnapshotRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gsclient.Client)
	snapshot, err := client.GetStorageSnapshot(d.Get("storage_uuid").(string), d.Id())
	if err != nil {
		if requestError, ok := err.(gsclient.RequestError); ok {
			if requestError.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}
	props := snapshot.Properties
	d.Set("status", props.Status)
	d.Set("location_country", props.LocationCountry)
	d.Set("location_name", props.LocationName)
	d.Set("location_iata", props.LocationIata)
	d.Set("location_uuid", props.LocationUUID)
	d.Set("usage_in_minutes", props.UsageInMinutes)
	d.Set("create_time", props.CreateTime)
	d.Set("change_time", props.ChangeTime)
	d.Set("license_product_no", props.LicenseProductNo)
	d.Set("current_price", props.CurrentPrice)
	d.Set("capacity", props.Capacity)
	//Set labels
	if err = d.Set("labels", props.Labels); err != nil {
		return fmt.Errorf("Error setting labels: %v", err)
	}
	return nil
}

func resourceGridscaleSnapshotCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gsclient.Client)
	requestBody := gsclient.StorageSnapshotCreateRequest{
		Name:   d.Get("name").(string),
		Labels: convSOStrings(d.Get("labels").([]interface{})),
	}
	response, err := client.CreateStorageSnapshot(d.Get("storage_uuid").(string), requestBody)
	if err != nil {
		return err
	}
	d.SetId(response.ObjectUUID)
	log.Printf("The id for snapshot %s has been set to %v", requestBody.Name, response.ObjectUUID)
	return resourceGridscaleSnapshotRead(d, meta)
}

func resourceGridscaleSnapshotUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gsclient.Client)
	requestBody := gsclient.StorageSnapshotUpdateRequest{
		Name:   d.Get("name").(string),
		Labels: convSOStrings(d.Get("labels").([]interface{})),
	}
	err := client.UpdateStorageSnapshot(d.Get("storage_uuid").(string), d.Id(), requestBody)
	if err != nil {
		return err
	}
	err = service_query.RetryUntilResourceStatusIsActive(client,
		service_query.SnapshotService,
		d.Timeout(schema.TimeoutUpdate),
		d.Get("storage_uuid").(string),
		d.Id(),
	)
	if err != nil {
		return err
	}
	return resourceGridscaleSnapshotRead(d, meta)
}

func resourceGridscaleSnapshotDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gsclient.Client)
	err := client.DeleteStorageSnapshot(d.Get("storage_uuid").(string), d.Id())
	if err != nil {
		return err
	}
	return service_query.RetryUntilDeleted(client, service_query.SnapshotService, d.Timeout(schema.TimeoutDelete), d.Get("storage_uuid").(string), d.Id())
}
