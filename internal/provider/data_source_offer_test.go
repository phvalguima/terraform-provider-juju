// Copyright 2023 Canonical Ltd.
// Licensed under the Apache License, Version 2.0, see LICENCE file for details.

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_DataSourceOffer_Edge(t *testing.T) {
	modelName := acctest.RandomWithPrefix("tf-datasource-offer-test-model")
	// ...-test-[0-9]+ is not a valid offer name, need to remove the dash before numbers
	offerName := fmt.Sprintf("tf-datasource-offer-test%d", acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: frameworkProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceOffer(modelName, offerName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.juju_offer.this", "model", modelName),
					resource.TestCheckResourceAttr("data.juju_offer.this", "name", offerName),
				),
			},
		},
	})
}

func TestAcc_DataSourceOffer_Stable(t *testing.T) {
	modelName := acctest.RandomWithPrefix("tf-datasource-offer-test-model")
	// ...-test-[0-9]+ is not a valid offer name, need to remove the dash before numbers
	offerName := fmt.Sprintf("tf-datasource-offer-test%d", acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		ExternalProviders: map[string]resource.ExternalProvider{
			"juju": {
				VersionConstraint: TestProviderStableVersion,
				Source:            "juju/juju",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceOffer(modelName, offerName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.juju_offer.this", "model", modelName),
					resource.TestCheckResourceAttr("data.juju_offer.this", "name", offerName),
				),
			},
		},
	})
}

func testAccDataSourceOffer(modelName string, offerName string) string {
	return fmt.Sprintf(`
resource "juju_model" "this" {
	name = %q
}

resource "juju_application" "this" {
	model = juju_model.this.name
	name  = "this"

	charm {
		name = "postgresql"
		series = "jammy"
	}
}

resource "juju_offer" "this" {
	model            = juju_model.this.name
	application_name = juju_application.this.name
	endpoint         = "db"
	name             = %q
}

data "juju_offer" "this" {
	url = juju_offer.this.url
}
`, modelName, offerName)
}
