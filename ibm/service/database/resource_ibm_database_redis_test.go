// Copyright IBM Corp. 2017, 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package database_test

import (
	"fmt"
	"testing"

	acc "github.com/IBM-Cloud/terraform-provider-ibm/ibm/acctest"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccIBMDatabaseInstance_Redis_Basic(t *testing.T) {
	t.Parallel()
	databaseResourceGroup := "default"
	var databaseInstanceOne string
	rnd := fmt.Sprintf("tf-redis-%d", acctest.RandIntRange(10, 100))
	testName := rnd
	name := "ibm_database." + testName

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acc.TestAccPreCheck(t) },
		Providers:    acc.TestAccProviders,
		CheckDestroy: testAccCheckIBMDatabaseInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckIBMDatabaseInstanceRedisBasic(databaseResourceGroup, testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIBMDatabaseInstanceExists(name, &databaseInstanceOne),
					resource.TestCheckResourceAttr(name, "name", testName),
					resource.TestCheckResourceAttr(name, "service", "databases-for-redis"),
					resource.TestCheckResourceAttr(name, "plan", "standard"),
					resource.TestCheckResourceAttr(name, "location", acc.Region()),
					resource.TestCheckResourceAttr(name, "adminuser", "admin"),
					resource.TestCheckResourceAttr(name, "groups.0.memory.0.allocation_mb", "8192"),
					resource.TestCheckResourceAttr(name, "groups.0.disk.0.allocation_mb", "2048"),
					resource.TestCheckResourceAttr(name, "allowlist.#", "1"),
				),
			},
			{
				Config: testAccCheckIBMDatabaseInstanceRedisFullyspecified(databaseResourceGroup, testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", testName),
					resource.TestCheckResourceAttr(name, "service", "databases-for-redis"),
					resource.TestCheckResourceAttr(name, "plan", "standard"),
					resource.TestCheckResourceAttr(name, "location", acc.Region()),
					resource.TestCheckResourceAttr(name, "groups.0.memory.0.allocation_mb", "10240"),
					resource.TestCheckResourceAttr(name, "groups.0.disk.0.allocation_mb", "4096"),
					resource.TestCheckResourceAttr(name, "allowlist.#", "2"),
				),
			},
			{
				Config: testAccCheckIBMDatabaseInstanceRedisReduced(databaseResourceGroup, testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", testName),
					resource.TestCheckResourceAttr(name, "service", "databases-for-redis"),
					resource.TestCheckResourceAttr(name, "plan", "standard"),
					resource.TestCheckResourceAttr(name, "location", acc.Region()),
					resource.TestCheckResourceAttr(name, "allowlist.#", "0"),
					resource.TestCheckResourceAttr(name, "groups.0.memory.0.allocation_mb", "8192"),
					resource.TestCheckResourceAttr(name, "groups.0.disk.0.allocation_mb", "4096"),
				),
			},
			{
				Config: testAccCheckIBMDatabaseInstanceRedisUserRole(databaseResourceGroup, testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", testName),
					resource.TestCheckResourceAttr(name, "service", "databases-for-redis"),
					resource.TestCheckResourceAttr(name, "plan", "standard"),
					resource.TestCheckResourceAttr(name, "location", acc.Region()),
					resource.TestCheckResourceAttr(name, "users.#", "1"),
					resource.TestCheckResourceAttr(name, "users.0.name", "coolguy"),
					resource.TestCheckResourceAttr(name, "users.0.role", "-@all +@read"),
					resource.TestCheckResourceAttr(name, "allowlist.#", "0"),
				),
			},
		},
	})
}

// TestAccIBMDatabaseInstance_CreateAfterManualDestroy not required as tested by resource_instance tests

func TestAccIBMDatabaseInstanceRedisImport(t *testing.T) {
	t.Parallel()
	databaseResourceGroup := "default"
	var databaseInstanceOne string
	serviceName := fmt.Sprintf("tf-redis-%d", acctest.RandIntRange(10, 100))
	resourceName := "ibm_database." + serviceName

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acc.TestAccPreCheck(t) },
		Providers:    acc.TestAccProviders,
		CheckDestroy: testAccCheckIBMDatabaseInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckIBMDatabaseInstanceRedisImport(databaseResourceGroup, serviceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIBMDatabaseInstanceExists(resourceName, &databaseInstanceOne),
					resource.TestCheckResourceAttr(resourceName, "name", serviceName),
					resource.TestCheckResourceAttr(resourceName, "service", "databases-for-redis"),
					resource.TestCheckResourceAttr(resourceName, "plan", "standard"),
					resource.TestCheckResourceAttr(resourceName, "location", acc.Region()),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling.0.disk.0.capacity_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling.0.memory.0.io_enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"wait_time_minutes", "deletion_protection"},
			},
		},
	})
}

func TestAccIBMDatabaseInstanceRedisKP_Encrypt(t *testing.T) {
	t.Parallel()
	databaseResourceGroup := "default"
	var databaseInstanceOne string
	rnd := fmt.Sprintf("tf-redis-%d", acctest.RandIntRange(10, 100))
	testName := rnd
	kpInstanceName := fmt.Sprintf("tf_kp_instance_%d", acctest.RandIntRange(10, 100))
	kpKeyName := fmt.Sprintf("tf_kp_key_%d", acctest.RandIntRange(10, 100))
	kpByokName := fmt.Sprintf("tf_kp_byok_key_%d", acctest.RandIntRange(10, 100))
	// name := "ibm_database." + testName

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acc.TestAccPreCheck(t) },
		Providers:    acc.TestAccProviders,
		CheckDestroy: testAccCheckIBMDatabaseInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckIBMDatabaseInstanceRedisKPEncrypt(databaseResourceGroup, kpInstanceName, kpKeyName, kpByokName, testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIBMDatabaseInstanceExists("ibm_database.database", &databaseInstanceOne),
					resource.TestCheckResourceAttr("ibm_database.database", "name", testName),
					resource.TestCheckResourceAttr("ibm_database.database", "service", "databases-for-redis"),
					resource.TestCheckResourceAttrSet("ibm_database.database", "key_protect_key"),
					resource.TestCheckResourceAttrSet("ibm_database.database", "backup_encryption_key_crn"),
				),
			},
		},
	})
}

// func testAccCheckIBMDatabaseInstanceDestroy(s *terraform.State) etc in resource_ibm_database_postgresql_test.go

func testAccCheckIBMDatabaseInstanceRedisBasic(databaseResourceGroup string, name string) string {
	return fmt.Sprintf(`
	data "ibm_resource_group" "test_acc" {
		is_default = true
		# name = "%[1]s"
	  }

	  resource "ibm_database" "%[2]s" {
		resource_group_id            = data.ibm_resource_group.test_acc.id
		name                         = "%[2]s"
		service                      = "databases-for-redis"
		plan                         = "standard"
		location                     = "%[3]s"
		adminpassword                = "secure-Password12345"
		service_endpoints            = "public"
		group {
			group_id = "member"
			memory {
				allocation_mb = 4096
			}
			host_flavor {
				id = "multitenant"
			}
			disk {
				allocation_mb = 1024
			}
		}
		allowlist {
		  address     = "172.168.1.2/32"
		  description = "desc1"
		}
		configuration                = <<CONFIGURATION
		{
		  "appendonly": "no",
		  "maxmemory": 0,
		  "maxmemory-policy": "noeviction",
		  "maxmemory-samples": 5,
		  "stop-writes-on-bgsave-error": "yes"
		}
		CONFIGURATION
	  }
				`, databaseResourceGroup, name, acc.Region())
}

func testAccCheckIBMDatabaseInstanceRedisFullyspecified(databaseResourceGroup string, name string) string {
	return fmt.Sprintf(`
	data "ibm_resource_group" "test_acc" {
		is_default = true
		# name = "%[1]s"
	}

	resource "ibm_database" "%[2]s" {
		resource_group_id            = data.ibm_resource_group.test_acc.id
		name                         = "%[2]s"
		service                      = "databases-for-redis"
		plan                         = "standard"
		location                     = "%[3]s"
		adminpassword                = "secure-Password12345"
		service_endpoints            = "public"
		group {
			group_id = "member"
			memory {
				allocation_mb = 5120
			}
			host_flavor {
				id = "multitenant"
			}
			disk {
				allocation_mb = 2048
			}
		}
		allowlist {
		  address     = "172.168.1.2/32"
		  description = "desc1"
		}
		allowlist {
		  address     = "172.168.1.1/32"
		  description = "desc"
		}
	}
				`, databaseResourceGroup, name, acc.Region())
}

func testAccCheckIBMDatabaseInstanceRedisReduced(databaseResourceGroup string, name string) string {
	return fmt.Sprintf(`
	data "ibm_resource_group" "test_acc" {
		is_default = true
		# name = "%[1]s"
	  }

	  resource "ibm_database" "%[2]s" {
		resource_group_id            = data.ibm_resource_group.test_acc.id
		name                         = "%[2]s"
		service                      = "databases-for-redis"
		plan                         = "standard"
		location                     = "%[3]s"
		adminpassword                = "secure-Password12345"
		service_endpoints            = "public"
		group {
			group_id = "member"
			memory {
				allocation_mb = 4096
			}
			host_flavor {
				id = "multitenant"
			}
			disk {
				allocation_mb = 2048
			}
		}
	  }
				`, databaseResourceGroup, name, acc.Region())
}

func testAccCheckIBMDatabaseInstanceRedisUserRole(databaseResourceGroup string, name string) string {
	return fmt.Sprintf(`
	data "ibm_resource_group" "test_acc" {
		is_default = true
	}

	resource "ibm_database" "%[2]s" {
		resource_group_id            = data.ibm_resource_group.test_acc.id
		name                         = "%[2]s"
		service                      = "databases-for-redis"
		plan                         = "standard"
		location                     = "%[3]s"
		adminpassword                = "secure-Password12345"
		service_endpoints            = "public"

		group {
			group_id = "member"

			memory {
				allocation_mb = 8192
			}
			host_flavor {
				id = "multitenant"
			}
			disk {
				allocation_mb = 2048
			}
		}

		users {
			name = "coolguy"
    		password = "securePassword123"
      		role     = "-@all +@read"
	 	}
  	}
				`, databaseResourceGroup, name, acc.Region())
}

func testAccCheckIBMDatabaseInstanceRedisImport(databaseResourceGroup string, name string) string {
	return fmt.Sprintf(`
	data "ibm_resource_group" "test_acc" {
		is_default = true
		# name = "%[1]s"
	  }

	  resource "ibm_database" "%[2]s" {
		resource_group_id = data.ibm_resource_group.test_acc.id
		name              = "%[2]s"
		service           = "databases-for-redis"
		plan              = "standard"
		location          = "%[3]s"
		service_endpoints            = "public"
		auto_scaling {
			disk {
			  capacity_enabled             = true
			  free_space_less_than_percent = 15
			  io_above_percent             = 85
			  io_enabled                   = true
			  io_over_period               = "15m"
			  rate_increase_percent        = 15
			  rate_limit_mb_per_member     = 3670016
			  rate_period_seconds          = 900
			  rate_units                   = "mb"
			}
		  memory {
			  io_above_percent         = 90
			  io_enabled               = true
			  io_over_period           = "15m"
			  rate_increase_percent    = 10
			  rate_limit_mb_per_member = 114688
			  rate_period_seconds      = 900
			  rate_units               = "mb"
			}
		}
	  }
				`, databaseResourceGroup, name, acc.Region())
}
func testAccCheckIBMDatabaseInstanceRedisKPEncrypt(databaseResourceGroup string, kpInstanceName, kpKeyName, kpByokName, name string) string {
	return fmt.Sprintf(`
	data "ibm_resource_group" "test_acc" {
		is_default = true
		# name = "%s"
	  }
	resource "ibm_resource_instance" "kp_instance" {
		name              = "%s"
		service           = "kms"
		plan              = "tiered-pricing"
		location          = "%[3]s"
	}
	resource "ibm_kp_key" "test" {
		key_protect_id 	 = ibm_resource_instance.kp_instance.guid
		key_name		 = "%s"
		force_delete	 = true
	}
	resource "ibm_kp_key" "test1" {
		key_protect_id 	= ibm_resource_instance.kp_instance.guid
		key_name 		= "%s"
		force_delete	= true
	}
	resource "ibm_database" "database" {
		resource_group_id 			= data.ibm_resource_group.test_acc.id
		name              			= "%s"
		service           			= "databases-for-redis"
		plan              			= "standard"
		location         			= "%[3]s"
		key_protect_instance        = ibm_resource_instance.kp_instance.guid
		key_protect_key             = ibm_kp_key.test.id
		backup_encryption_key_crn   = ibm_kp_key.test1.id
		service_endpoints           = "public"
		timeouts {
			create = "480m"
			update = "480m"
			delete = "15m"
		}
	}
				`, databaseResourceGroup, kpInstanceName, kpKeyName, kpByokName, name, acc.Region())
}
