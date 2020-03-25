package gridscale

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceGridscaleObjectStorage() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGridscaleObjectStorageRead,
		Schema: map[string]*schema.Schema{
			"project": {
				Type:        schema.TypeString,
				Description: "The project name that set in `GRIDSCALE_PROJECTS_TOKENS` env or `projects_tokens` tf variable",
				Required:    true,
			},
			"resource_id": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				Description:  "ID of a resource",
				ValidateFunc: validation.NoZeroValues,
			},
			"access_key": {
				Type:        schema.TypeString,
				Description: "The object storage secret_key.",
				Computed:    true,
			},
			"secret_key": {
				Type:        schema.TypeString,
				Description: "The object storage access_key.",
				Computed:    true,
			},
		},
	}
}

func dataSourceGridscaleObjectStorageRead(d *schema.ResourceData, meta interface{}) error {
	projectName := d.Get("project").(string)
	client, err := getProjectClientFromMeta(projectName, meta)
	if err != nil {
		return err
	}

	id := d.Get("resource_id").(string)
	errorPrefix := fmt.Sprintf("read object storage (%s) datasource-", id)

	objectStorage, err := client.GetObjectStorageAccessKey(emptyCtx, id)
	if err != nil {
		return fmt.Errorf("%s error: %v", errorPrefix, err)
	}

	d.SetId(objectStorage.Properties.AccessKey)
	if err = d.Set("access_key", objectStorage.Properties.AccessKey); err != nil {
		return fmt.Errorf("%s error setting access_key: %v", errorPrefix, err)
	}
	if err = d.Set("secret_key", objectStorage.Properties.SecretKey); err != nil {
		return fmt.Errorf("%s error setting access_key: %v", errorPrefix, err)
	}

	return nil
}
