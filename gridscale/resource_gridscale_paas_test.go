package gridscale

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/nvthongswansea/gsclient-go"
	"testing"
)

func TestAccDataSourceGridscalePaaS_Basic(t *testing.T) {
	var object gsclient.PaaSService
	name := fmt.Sprintf("object-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDataSourceGridscalePaaSDestroyCheck,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDataSourceGridscalePaaSConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceGridscalePaaSExists("gridscale_paas.foopaas", &object),
					resource.TestCheckResourceAttr(
						"gridscale_paas.foopaas", "name", name),
					resource.TestCheckResourceAttr(
						"gridscale_paas.foopaas", "service_template_uuid", "f9625726-5ca8-4d5c-b9bd-3257e1e2211a"),
				),
			},
			{
				Config: testAccCheckDataSourceGridscalePaaSConfig_basic_update(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceGridscalePaaSExists("gridscale_paas.foopaas", &object),
					resource.TestCheckResourceAttr(
						"gridscale_paas.foopaas", "name", "newname"),
					resource.TestCheckResourceAttr(
						"gridscale_paas.foopaas", "service_template_uuid", "f9625726-5ca8-4d5c-b9bd-3257e1e2211a"),
				),
			},
			{
				Config: testAccCheckDataSourceGridscalePaaSConfig_forcenew_update(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceGridscalePaaSExists("gridscale_paas.foopaas", &object),
					resource.TestCheckResourceAttr(
						"gridscale_paas.foopaas", "name", "newname"),
					resource.TestCheckResourceAttr(
						"gridscale_paas.foopaas", "service_template_uuid", "136c1446-13e0-4734-bdb6-ab0a15c1d680"),
				),
			},
		},
	})
}

func testAccCheckDataSourceGridscalePaaSExists(n string, object *gsclient.PaaSService) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No object UUID is set")
		}
		client := testAccProvider.Meta().(*gsclient.Client)
		id := rs.Primary.ID
		foundObject, err := client.GetPaaSService(id)
		if err != nil {
			return err
		}
		if foundObject.Properties.ObjectUUID != id {
			return fmt.Errorf("Object not found")
		}
		*object = foundObject
		return nil
	}
}

func testAccCheckDataSourceGridscalePaaSDestroyCheck(s *terraform.State) error {
	client := testAccProvider.Meta().(*gsclient.Client)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "gridscale_paas" {
			continue
		}

		_, err := client.GetPaaSService(rs.Primary.ID)
		if err != nil {
			if requestError, ok := err.(gsclient.RequestError); ok {
				if requestError.StatusCode != 404 {
					return fmt.Errorf("Object %s still exists", rs.Primary.ID)
				}
			} else {
				return fmt.Errorf("Unable to fetching object %s", rs.Primary.ID)
			}
		} else {
			return fmt.Errorf("Object %s still exists", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckDataSourceGridscalePaaSConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "gridscale_paas" "foopaas" {
  name = "%s"
  service_template_uuid = "f9625726-5ca8-4d5c-b9bd-3257e1e2211a"
}
`, name)
}

func testAccCheckDataSourceGridscalePaaSConfig_basic_update() string {
	return fmt.Sprintf(`
resource "gridscale_paas" "foopaas" {
  name = "newname"
  service_template_uuid = "f9625726-5ca8-4d5c-b9bd-3257e1e2211a"
  resource_limit {
	resource = "cores"
	limit = 16
  }
}
`)
}

func testAccCheckDataSourceGridscalePaaSConfig_forcenew_update() string {
	return fmt.Sprintf(`
resource "gridscale_paas" "foopaas" {
  name = "newname"
  service_template_uuid = "136c1446-13e0-4734-bdb6-ab0a15c1d680"
  resource_limit {
	resource = "cores"
	limit = 16
  }
}
`)
}
