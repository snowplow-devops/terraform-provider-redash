//
// Copyright (c) 2020 Snowplow Analytics Ltd. All rights reserved.
//
// This program is licensed to you under the Apache License Version 2.0,
// and you may not use this file except in compliance with the Apache License Version 2.0.
// You may obtain a copy of the Apache License Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0.
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the Apache License Version 2.0 is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the Apache License Version 2.0 for the specific language governing permissions and limitations there under.
//
package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/snowplow-devops/redash-client-go/redash"
)

func resourceRedashDataSource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRedashDataSourceCreate,
		ReadContext:   resourceRedashDataSourceRead,
		UpdateContext: resourceRedashDataSourceUpdate,
		DeleteContext: resourceRedashDataSourceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"last_updated": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"scheduled_queue_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"queue_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"paused": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"pause_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"syntax": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"groups": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"options": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"access_key": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"account": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"adhoc_query_group": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"allowed_schemas": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"api_key": {
							Type:      schema.TypeString,
							Sensitive: true,
							Optional:  true,
						},
						"api_server": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"api_version": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"apikey": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"aws_access_key": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"aws_secret_key": {
							Type:      schema.TypeString,
							Sensitive: true,
							Optional:  true,
						},
						"azure_ad_client_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"azure_ad_client_secret": {
							Type:      schema.TypeString,
							Sensitive: true,
							Optional:  true,
						},
						"azure_ad_tenant_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"basic_auth_password": {
							Type:      schema.TypeString,
							Sensitive: true,
							Optional:  true,
						},
						"basic_auth_user": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"catalog": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"charset": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"cluster": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"connection_string": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"connection_timeout": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"customer_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"database": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"db": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"db_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"dbname": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"dbpath": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"encryption_option": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"endpoint": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"expression": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"get_schema": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"glue": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"host": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"hostname": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"http_password": {
							Type:      schema.TypeString,
							Sensitive: true,
							Optional:  true,
						},
						"http_path": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"http_scheme": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"json_key_file": {
							Type:      schema.TypeString,
							Sensitive: true,
							Optional:  true,
						},
						"key": {
							Type:      schema.TypeString,
							Sensitive: true,
							Optional:  true,
						},
						"keyspace": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"kms_key": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ldap_password": {
							Type:      schema.TypeString,
							Sensitive: true,
							Optional:  true,
						},
						"ldap_user": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"limit": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"load_schema": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"location": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"maximum_billing_tier": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"min_insert_date": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"passwd": {
							Type:      schema.TypeString,
							Sensitive: true,
							Optional:  true,
						},
						"password": {
							Type:      schema.TypeString,
							Sensitive: true,
							Optional:  true,
						},
						"port": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"project": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"project_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"protocol": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"query_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"read_timeout": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"region": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"replica_set_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"s3_staging_dir": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"sandbox": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"scheduled_query_group": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"schema": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"scheme": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"secret": {
							Type:      schema.TypeString,
							Sensitive: true,
							Optional:  true,
						},
						"secret_key": {
							Type:      schema.TypeString,
							Sensitive: true,
							Optional:  true,
						},
						"server": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"servers": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ssl_cacert": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ssl_cert": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ssl_key": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"sslmode": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"tds_version": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"timeout": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"token": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"total_mbytes_processed_limit": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"trust_certificate": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"url": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"use_standard_sql": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"use_aws_iam_profile": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"use_ldap": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"use_ssl": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"user": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"user_defined_function_resource_uri": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"username": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"verify": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"verify_ssl": {
							Type:     schema.TypeBool,
							Optional: true,
							DefaultFunc: func() (interface{}, error) {
								return nil, nil
							},
						},
						"warehouse": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"work_group": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ssh_tunnel": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ssh_username": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"ssh_port": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"ssh_host": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceRedashDataSourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*redash.Client)

	var diags diag.Diagnostics

	options := d.Get("options").([]interface{})[0].(map[string]interface{})
	payload := redash.DataSource{
		Name:               d.Get("name").(string),
		Type:               d.Get("type").(string),
		ScheduledQueueName: d.Get("scheduled_queue_name").(string),
		PauseReason:        d.Get("pause_reason").(string),
		QueueName:          d.Get("queue_name").(string),
		Syntax:             d.Get("syntax").(string),
		Paused:             d.Get("paused").(int),
		Options:            convertOptions(&options, "redash"),
	}

	dataSource, err := c.CreateDataSource(&payload)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprint(dataSource.ID))

	resourceRedashDataSourceRead(ctx, d, meta)

	return diags
}

func resourceRedashDataSourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*redash.Client)

	var diags diag.Diagnostics

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	dataSource, err := c.GetDataSource(id)
	if err != nil {
		return diag.FromErr(err)
	}

	options := convertOptions(&dataSource.Options, "terraform")
	d.Set("name", &dataSource.Name)
	d.Set("scheduled_queue_name", &dataSource.ScheduledQueueName)
	d.Set("pause_reason", &dataSource.PauseReason)
	d.Set("queue_name", &dataSource.QueueName)
	d.Set("syntax", &dataSource.Syntax)
	d.Set("paused", &dataSource.Paused)
	d.Set("type", &dataSource.Type)
	d.Set("options", &options)

	d.SetId(fmt.Sprint(dataSource.ID))

	return diags
}

func resourceRedashDataSourceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*redash.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	options := d.Get("options").([]interface{})[0].(map[string]interface{})
	payload := redash.DataSource{
		Name:               d.Get("name").(string),
		Type:               d.Get("type").(string),
		ScheduledQueueName: d.Get("scheduled_queue_name").(string),
		PauseReason:        d.Get("pause_reason").(string),
		QueueName:          d.Get("queue_name").(string),
		Syntax:             d.Get("syntax").(string),
		Paused:             d.Get("paused").(int),
		Options:            convertOptions(&options, "redash"),
	}

	_, err = c.UpdateDataSource(id, &payload)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceRedashDataSourceRead(ctx, d, meta)
}

func resourceRedashDataSourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*redash.Client)

	var diags diag.Diagnostics

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = c.DeleteDataSource(id)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}

func convertOptions(options *map[string]interface{}, toFormat string) map[string]interface{} {

	redashConversion := map[string]string{
		"connection_string":                  "connectionString",
		"db_name":                            "dbName",
		"json_key_file":                      "jsonKeyFile",
		"load_schema":                        "loadSchema",
		"maximum_billing_tier":               "maximumBillingTier",
		"project_id":                         "projectId",
		"replica_set_name":                   "replicaSetName",
		"total_mbytes_processed_limit":       "totalMBytesProcessedLimit",
		"use_standard_sql":                   "useStandardSql",
		"user_defined_function_resource_uri": "userDefinedFunctionResourceUri",
		"use_ssl":                            "useSsl",
	}

	terraformConversion := map[string]string{
		"connectionString":               "connection_string",
		"dbName":                         "db_name",
		"jsonKeyFile":                    "json_key_file",
		"loadSchema":                     "load_schema",
		"maximumBillingTier":             "maximum_billing_tier",
		"projectId":                      "project_id",
		"replicaSetName":                 "replica_set_name",
		"totalMBytesProcessedLimit":      "total_mbytes_processed_limit",
		"useStandardSql":                 "use_standard_sql",
		"userDefinedFunctionResourceUri": "user_defined_function_resource_uri",
		"useSsl":                         "use_ssl",
	}

	convertedOptions := map[string]interface{}{}

	for k, v := range *options {
		switch value := v.(type) {
		case []interface{}:
			if k == "ssh_tunnel" && len(value) > 0 {
				convertedOptions[k] = value[0]
			}
		default:
			if toFormat == "redash" {
				if val, ok := redashConversion[k]; ok {
					convertedOptions[val] = v
				} else {
					convertedOptions[k] = v
				}
			}

			if toFormat == "terraform" {
				if val, ok := terraformConversion[k]; ok {
					convertedOptions[val] = v
				} else {
					convertedOptions[k] = v
				}
			}
		}
	}

	return convertedOptions
}
