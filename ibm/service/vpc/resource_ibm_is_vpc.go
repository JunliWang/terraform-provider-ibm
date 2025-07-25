// Copyright IBM Corp. 2017, 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package vpc

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/vpc-go-sdk/vpcv1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/validate"
)

const (
	isVPCDefaultNetworkACL                    = "default_network_acl"
	isVPCDefaultSecurityGroup                 = "default_security_group"
	isVPCDefaultRoutingTable                  = "default_routing_table"
	isVPCName                                 = "name"
	isVPCDefaultNetworkACLName                = "default_network_acl_name"
	isVPCDefaultNetworkACLCRN                 = "default_network_acl_crn"
	isVPCDefaultSecurityGroupName             = "default_security_group_name"
	isVPCDefaultSecurityGroupCRN              = "default_security_group_crn"
	isVPCDefaultRoutingTableName              = "default_routing_table_name"
	isVPCDefaultRoutingTableCRN               = "default_routing_table_crn"
	isVPCResourceGroup                        = "resource_group"
	isVPCStatus                               = "status"
	isVPCDeleting                             = "deleting"
	isVPCDeleted                              = "done"
	isVPCTags                                 = "tags"
	isVPCClassicAccess                        = "classic_access"
	isVPCAvailable                            = "available"
	isVPCFailed                               = "failed"
	isVPCPending                              = "pending"
	isVPCAddressPrefixManagement              = "address_prefix_management"
	cseSourceAddresses                        = "cse_source_addresses"
	subnetsList                               = "subnets"
	totalIPV4AddressCount                     = "total_ipv4_address_count"
	availableIPV4AddressCount                 = "available_ipv4_address_count"
	isVPCCRN                                  = "crn"
	isVPCSecurityGroupList                    = "security_group"
	isVPCSecurityGroupName                    = "group_name"
	isVPCSgRules                              = "rules"
	isVPCSecurityGroupRuleID                  = "rule_id"
	isVPCSecurityGroupRuleDirection           = "direction"
	isVPCSecurityGroupRuleIPVersion           = "ip_version"
	isVPCSecurityGroupRuleRemote              = "remote"
	isVPCSecurityGroupRuleType                = "type"
	isVPCSecurityGroupRuleCode                = "code"
	isVPCSecurityGroupRulePortMax             = "port_max"
	isVPCSecurityGroupRulePortMin             = "port_min"
	isVPCSecurityGroupRuleProtocol            = "protocol"
	isVPCSecurityGroupID                      = "group_id"
	isVPCAccessTags                           = "access_tags"
	isVPCAccessTagType                        = "access"
	isVPCUserTagType                          = "user"
	isVPCDns                                  = "dns"
	isVPCDnsEnableHub                         = "enable_hub"
	isVPCDnsResolutionBindingCount            = "resolution_binding_count"
	isVPCDnsResolver                          = "resolver"
	isVPCDnsResolverManualServers             = "manual_servers"
	isVPCDnsResolverServers                   = "servers"
	isVPCDnsResolverManualServersAddress      = "address"
	isVPCDnsResolverManualServersZoneAffinity = "zone_affinity"
	isVPCDnsResolverType                      = "type"
	isVPCDnsResolverVpc                       = "vpc"
	isVPCDnsResolverResourceType              = "resource_type"
	isVPCDnsResolverConfiguration             = "configuration"
	isVPCDnsResolverVpcId                     = "id"
	isVPCDnsResolverVpcHref                   = "href"
	isVPCDnsResolverVpcCrn                    = "crn"
	isVPCDnsResolverVpcName                   = "name"
	isVPCDnsResolverVpcDeleted                = "deleted"
	isVPCDnsResolverVpcDeletedMoreInfo        = "more_info"
	isVPCDnsResolverVpcRemote                 = "remote"
	isVPCDnsResolverVpcRemoteAccount          = "account"
	isVPCDnsResolverVpcRemoteRegion           = "region"
	isVPCNoSgAclRules                         = "no_sg_acl_rules"
)

func ResourceIBMISVPC() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIBMISVPCCreate,
		ReadContext:   resourceIBMISVPCRead,
		UpdateContext: resourceIBMISVPCUpdate,
		DeleteContext: resourceIBMISVPCDelete,
		Exists:        resourceIBMISVPCExists,
		Importer:      &schema.ResourceImporter{},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		CustomizeDiff: customdiff.All(
			customdiff.Sequence(
				func(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
					return flex.ResourceTagsCustomizeDiff(diff)
				},
			),
			customdiff.Sequence(
				func(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
					return flex.ResourceValidateAccessTags(diff, v)
				}),
		),

		Schema: map[string]*schema.Schema{

			"default_address_prefixes": {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed:    true,
				Description: "Default address prefixes for each zone.",
			},
			isVPCAddressPrefixManagement: {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "auto",
				DiffSuppressFunc: suppressNullAddPrefix,
				ForceNew:         true,
				ValidateFunc:     validate.InvokeValidator("ibm_is_vpc", isVPCAddressPrefixManagement),
				Description:      "Address Prefix management value",
			},

			isVPCDefaultNetworkACL: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Default network ACL ID",
			},

			isVPCDefaultRoutingTable: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Default routing table associated with VPC",
			},
			isVPCDns: &schema.Schema{
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Description: "The DNS configuration for this VPC.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						isVPCDnsEnableHub: &schema.Schema{
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Indicates whether this VPC is enabled as a DNS name resolution hub.",
						},
						isVPCDnsResolutionBindingCount: &schema.Schema{
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The number of DNS resolution bindings for this VPC.",
						},
						isVPCDnsResolver: &schema.Schema{
							Type:        schema.TypeList,
							MinItems:    0,
							MaxItems:    1,
							Optional:    true,
							Computed:    true,
							Description: "The DNS resolver configuration for the VPC.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									isVPCDnsResolverServers: &schema.Schema{
										Type:        schema.TypeList,
										Computed:    true,
										Description: "The DNS servers for this VPC. The servers are populated:- by the system when `dns.resolver.type` is `system`- using the DNS servers in `dns.resolver.vpc` when `dns.resolver.type` is `delegated`- using `dns.resolver.manual_servers` when the `dns.resolver.type` is `manual`.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												isVPCDnsResolverManualServersAddress: &schema.Schema{
													Type:        schema.TypeString,
													Computed:    true,
													Description: "The IP address.This property may add support for IPv6 addresses in the future. When processing a value in this property, verify that the address is in an expected format. If it is not, log an error. Optionally halt processing and surface the error, or bypass the resource on which the unexpected IP address format was encountered.",
												},
												isVPCDnsResolverManualServersZoneAffinity: &schema.Schema{
													Type:        schema.TypeString,
													Computed:    true,
													Description: "Zone name, if present, DHCP configuration for this zone will have this DNS server listed first.",
												},
											},
										},
									},
									isVPCDnsResolverType: &schema.Schema{
										Type:        schema.TypeString,
										Optional:    true,
										Computed:    true,
										Description: "The type of the DNS resolver used for the VPC.- `delegated`: DNS server addresses are provided by the DNS resolver of the VPC               specified in `dns.resolver.vpc`.- `manual`: DNS server addresses are specified in `dns.resolver.manual_servers`.- `system`: DNS server addresses are provided by the system.",
									},

									"dns_binding_id": &schema.Schema{
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The VPC dns binding id whose DNS resolver provides the DNS server addresses for this VPC.",
									},
									"dns_binding_name": &schema.Schema{
										Type:             schema.TypeString,
										Optional:         true,
										DiffSuppressFunc: suppressNullDnsBindingName,
										Computed:         true,
										Description:      "The VPC dns binding name whose DNS resolver provides the DNS server addresses for this VPC.",
									},
									"vpc_id": &schema.Schema{
										Type:             schema.TypeString,
										Optional:         true,
										DiffSuppressFunc: suppressNullVPC,
										ConflictsWith:    []string{"dns.0.resolver.0.vpc_crn"},
										Description:      "The VPC id whose DNS resolver provides the DNS server addresses for this VPC.The VPC may be remote and therefore may not be directly retrievable.",
									},
									"vpc_name": &schema.Schema{
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The VPC name whose DNS resolver provides the DNS server addresses for this VPC.The VPC may be remote and therefore may not be directly retrievable.",
									},
									"vpc_crn": &schema.Schema{
										Type:             schema.TypeString,
										Optional:         true,
										DiffSuppressFunc: suppressNullVPC,

										ConflictsWith: []string{"dns.0.resolver.0.vpc_id"},
										Description:   "The VPC crn whose DNS resolver provides the DNS server addresses for this VPC.The VPC may be remote and therefore may not be directly retrievable.",
									},
									"vpc_remote_account_id": &schema.Schema{
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The unique identifier for this account.",
									},
									"vpc_remote_region": &schema.Schema{
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Region name. If present, this property indicates that the referenced resource is remote to this region, and identifies the native region.",
									},

									isVPCDnsResolverManualServers: &schema.Schema{
										Type:        schema.TypeSet,
										Optional:    true,
										Computed:    true,
										Set:         hashManualServersList,
										Description: "The manually specified DNS servers for this VPC.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												isVPCDnsResolverManualServersAddress: &schema.Schema{
													Type:        schema.TypeString,
													Optional:    true,
													Computed:    true,
													Description: "The IP address.This property may add support for IPv6 addresses in the future. When processing a value in this property, verify that the address is in an expected format. If it is not, log an error. Optionally halt processing and surface the error, or bypass the resource on which the unexpected IP address format was encountered.",
												},
												isVPCDnsResolverManualServersZoneAffinity: &schema.Schema{
													Type:        schema.TypeString,
													Optional:    true,
													Computed:    true,
													Description: "The name of the zone. If present, DHCP configuration for this zone will have this DNS server listed first.",
												},
											},
										},
									},
									isVPCDnsResolverConfiguration: &schema.Schema{
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The configuration of the system DNS resolver for this VPC.- `custom_resolver`: A custom DNS resolver is configured for this VPC.- `private_resolver`: A private DNS resolver is configured for this VPC. Applicable when  the VPC has either or both of the following:    - at least one endpoint gateway residing in it    - a [DNS Services](https://cloud.ibm.com/docs/dns-svcs) private zone configured for it- `default`: The provider default DNS resolvers are configured for this VPC.  This system DNS resolver configuration is used when the VPC has:  - no custom DNS resolver configured for it, and  - no endpoint gateways residing in it, and  - no [DNS Services](https://cloud.ibm.com/docs/dns-svcs) private zone configured for it.",
									},
								},
							},
						},
					},
				},
			},
			"health_reasons": &schema.Schema{
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The reasons for the current `health_state` (if any).The enumerated reason code values for this property will expand in the future. When processing this property, check for and log unknown values. Optionally halt processing and surface the error, or bypass the resource on which the unexpected reason code was encountered.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"code": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "A snake case string succinctly identifying the reason for this health state.",
						},
						"message": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "An explanation of the reason for this health state.",
						},
						"more_info": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Link to documentation about the reason for this health state.",
						},
					},
				},
			},
			"health_state": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The health of this resource.- `ok`: No abnormal behavior detected- `degraded`: Experiencing compromised performance, capacity, or connectivity- `faulted`: Completely unreachable, inoperative, or otherwise entirely incapacitated- `inapplicable`: The health state does not apply because of the current lifecycle state. A resource with a lifecycle state of `failed` or `deleting` will have a health state of `inapplicable`. A `pending` resource may also have this state.",
			},
			isVPCClassicAccess: {
				Type:        schema.TypeBool,
				ForceNew:    true,
				Deprecated:  "Classic access is deprecated",
				Default:     false,
				Optional:    true,
				Description: "Set to true if classic access needs to enabled to VPC",
			},

			isVPCNoSgAclRules: {
				Type:             schema.TypeBool,
				Default:          false,
				DiffSuppressFunc: flex.ApplyOnce,
				Optional:         true,
				Description:      "Delete all rules attached with default security group and default acl",
			},

			isVPCName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     false,
				ValidateFunc: validate.InvokeValidator("ibm_is_vpc", isVPCName),
				Description:  "VPC name",
			},

			isVPCDefaultNetworkACLName: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validate.InvokeValidator("ibm_is_vpc", isVPCDefaultNetworkACLName),
				Description:  "Default Network ACL name",
			},

			isVPCDefaultSecurityGroupName: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validate.InvokeValidator("ibm_is_vpc", isVPCDefaultSecurityGroupName),
				Description:  "Default security group name",
			},

			isVPCDefaultSecurityGroupCRN: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Default security group CRN",
			},

			isVPCDefaultNetworkACLCRN: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Default Network ACL CRN",
			},

			isVPCDefaultRoutingTableName: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validate.InvokeValidator("ibm_is_vpc", isVPCDefaultRoutingTableName),
				Description:  "Default routing table name",
			},

			isVPCDefaultRoutingTableCRN: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Default routing table CRN",
			},

			isVPCResourceGroup: {
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Computed:    true,
				Description: "Resource group info",
			},

			isVPCStatus: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "VPC status",
			},

			isVPCDefaultSecurityGroup: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Security group associated with VPC",
			},
			isVPCTags: {
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString, ValidateFunc: validate.InvokeValidator("ibm_is_vpc", "tags")},
				Set:         flex.ResourceIBMVPCHash,
				Description: "List of tags",
			},
			isVPCAccessTags: {
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString, ValidateFunc: validate.InvokeValidator("ibm_is_vpc", "accesstag")},
				Set:         flex.ResourceIBMVPCHash,
				Description: "List of access management tags",
			},
			isVPCCRN: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The crn of the resource",
			},

			flex.ResourceControllerURL: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The URL of the IBM Cloud dashboard that can be used to explore and view details about this instance",
			},
			flex.ResourceName: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the resource",
			},

			flex.ResourceCRN: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The crn of the resource",
			},

			flex.ResourceStatus: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the resource",
			},

			flex.ResourceGroupName: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The resource group name in which resource is provisioned",
			},

			cseSourceAddresses: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Cloud service endpoint IP Address",
						},

						"zone_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Location info of CSE Address",
						},
					},
				},
			},

			isVPCSecurityGroupList: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						isVPCSecurityGroupName: {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Security group name",
						},

						isVPCSecurityGroupID: {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Security group id",
						},

						isSecurityGroupRules: {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Security Rules",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{

									isVPCSecurityGroupRuleID: {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Rule ID",
									},

									isVPCSecurityGroupRuleDirection: {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Direction of traffic to enforce, either inbound or outbound",
									},

									isVPCSecurityGroupRuleIPVersion: {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "IP version: ipv4",
									},

									isVPCSecurityGroupRuleRemote: {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Security group id: an IP address, a CIDR block, or a single security group identifier",
									},

									isVPCSecurityGroupRuleType: {
										Type:     schema.TypeInt,
										Computed: true,
									},

									isVPCSecurityGroupRuleCode: {
										Type:     schema.TypeInt,
										Computed: true,
									},

									isVPCSecurityGroupRulePortMin: {
										Type:     schema.TypeInt,
										Computed: true,
									},

									isVPCSecurityGroupRulePortMax: {
										Type:     schema.TypeInt,
										Computed: true,
									},

									isVPCSecurityGroupRuleProtocol: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},

			subnetsList: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "subent name",
						},

						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "subnet ID",
						},

						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "subnet status",
						},

						"zone": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "subnet location",
						},

						totalIPV4AddressCount: {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Total IPv4 address count in the subnet",
						},

						availableIPV4AddressCount: {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Available IPv4 address count in the subnet",
						},
					},
				},
			},
			"public_address_ranges": &schema.Schema{
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The public address ranges attached to this VPC.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"crn": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The CRN for this public address range.",
						},
						"deleted": &schema.Schema{
							Type:        schema.TypeList,
							Computed:    true,
							Description: "If present, this property indicates the referenced resource has been deleted, and providessome supplementary information.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"more_info": &schema.Schema{
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Link to documentation about deleted resources.",
									},
								},
							},
						},
						"href": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The URL for this public address range.",
						},
						"id": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The unique identifier for this public address range.",
						},
						"name": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name for this public address range. The name is unique across all public address ranges in the region.",
						},
						"resource_type": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The resource type.",
						},
					},
				},
			},
		},
	}
}

func ResourceIBMISVPCValidator() *validate.ResourceValidator {

	validateSchema := make([]validate.ValidateSchema, 0)
	address_prefix_management := "auto, manual"

	validateSchema = append(validateSchema,
		validate.ValidateSchema{
			Identifier:                 isVPCAddressPrefixManagement,
			ValidateFunctionIdentifier: validate.ValidateAllowedStringValue,
			Type:                       validate.TypeString,
			Optional:                   true,
			Default:                    "auto",
			AllowedValues:              address_prefix_management})
	validateSchema = append(validateSchema,
		validate.ValidateSchema{
			Identifier:                 "id",
			ValidateFunctionIdentifier: validate.ValidateCloudData,
			Type:                       validate.TypeString,
			CloudDataType:              "is",
			CloudDataRange:             []string{"service:vpc", "resolved_to:id"}})
	validateSchema = append(validateSchema,
		validate.ValidateSchema{
			Identifier:                 isVPCName,
			ValidateFunctionIdentifier: validate.ValidateRegexpLen,
			Type:                       validate.TypeString,
			Required:                   true,
			Regexp:                     `^([a-z]|[a-z][-a-z0-9]*[a-z0-9])$`,
			MinValueLength:             1,
			MaxValueLength:             63})
	validateSchema = append(validateSchema,
		validate.ValidateSchema{
			Identifier:                 isVPCDefaultNetworkACLName,
			ValidateFunctionIdentifier: validate.ValidateRegexpLen,
			Type:                       validate.TypeString,
			Required:                   true,
			Regexp:                     `^([a-z]|[a-z][-a-z0-9]*[a-z0-9])$`,
			MinValueLength:             1,
			MaxValueLength:             63})
	validateSchema = append(validateSchema,
		validate.ValidateSchema{
			Identifier:                 isVPCDefaultSecurityGroupName,
			ValidateFunctionIdentifier: validate.ValidateRegexpLen,
			Type:                       validate.TypeString,
			Required:                   true,
			Regexp:                     `^([a-z]|[a-z][-a-z0-9]*[a-z0-9])$`,
			MinValueLength:             1,
			MaxValueLength:             63})
	validateSchema = append(validateSchema,
		validate.ValidateSchema{
			Identifier:                 isVPCDefaultRoutingTableName,
			ValidateFunctionIdentifier: validate.ValidateRegexpLen,
			Type:                       validate.TypeString,
			Required:                   true,
			Regexp:                     `^([a-z]|[a-z][-a-z0-9]*[a-z0-9])$`,
			MinValueLength:             1,
			MaxValueLength:             63})

	validateSchema = append(validateSchema,
		validate.ValidateSchema{
			Identifier:                 "tags",
			ValidateFunctionIdentifier: validate.ValidateRegexpLen,
			Type:                       validate.TypeString,
			Optional:                   true,
			Regexp:                     `^[A-Za-z0-9:_ .-]+$`,
			MinValueLength:             1,
			MaxValueLength:             128})
	validateSchema = append(validateSchema,
		validate.ValidateSchema{
			Identifier:                 "accesstag",
			ValidateFunctionIdentifier: validate.ValidateRegexpLen,
			Type:                       validate.TypeString,
			Optional:                   true,
			Regexp:                     `^([A-Za-z0-9_.-]|[A-Za-z0-9_.-][A-Za-z0-9_ .-]*[A-Za-z0-9_.-]):([A-Za-z0-9_.-]|[A-Za-z0-9_.-][A-Za-z0-9_ .-]*[A-Za-z0-9_.-])$`,
			MinValueLength:             1,
			MaxValueLength:             128})

	ibmISVPCResourceValidator := validate.ResourceValidator{ResourceName: "ibm_is_vpc", Schema: validateSchema}
	return &ibmISVPCResourceValidator
}

func resourceIBMISVPCCreate(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	log.Printf("[DEBUG] VPC create")
	name := d.Get(isVPCName).(string)
	apm := ""
	rg := ""
	isClassic := false

	if addprefixmgmt, ok := d.GetOk(isVPCAddressPrefixManagement); ok {
		apm = addprefixmgmt.(string)
	}
	if classic, ok := d.GetOk(isVPCClassicAccess); ok {
		isClassic = classic.(bool)
	}

	if grp, ok := d.GetOk(isVPCResourceGroup); ok {
		rg = grp.(string)
	}
	err := vpcCreate(context, d, meta, name, apm, rg, isClassic)
	if err != nil {
		return err
	}
	return resourceIBMISVPCRead(context, d, meta)
}

func vpcCreate(context context.Context, d *schema.ResourceData, meta interface{}, name, apm, rg string, isClassic bool) diag.Diagnostics {
	sess, err := vpcClient(meta)
	if err != nil {
		tfErr := flex.DiscriminatedTerraformErrorf(err, err.Error(), "ibm_is_vpc", "create", "initialize-client")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}
	options := &vpcv1.CreateVPCOptions{
		Name: &name,
	}
	if _, ok := d.GetOk(isVPCDns); ok {
		dnsModel, err := resourceIBMIsVPCMapToVpcdnsPrototype(d.Get("dns.0").(map[string]interface{}))
		if err != nil {
			return flex.DiscriminatedTerraformErrorf(err, err.Error(), "ibm_is_vpc", "create", "parse-dns").GetDiag()
		}
		options.SetDns(dnsModel)
	}
	if rg != "" {
		options.ResourceGroup = &vpcv1.ResourceGroupIdentity{
			ID: &rg,
		}
	}
	if apm != "" {
		options.AddressPrefixManagement = &apm
	}
	options.ClassicAccess = &isClassic

	vpc, _, err := sess.CreateVPCWithContext(context, options)
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("CreateVPCWithContext failed: %s", err.Error()), "ibm_is_vpc", "create")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}
	d.SetId(*vpc.ID)

	if defaultSGName, ok := d.GetOk(isVPCDefaultSecurityGroupName); ok {
		sgNameUpdate(sess, *vpc.DefaultSecurityGroup.ID, defaultSGName.(string))
	}

	if defaultRTName, ok := d.GetOk(isVPCDefaultRoutingTableName); ok {
		rtNameUpdate(sess, *vpc.ID, *vpc.DefaultRoutingTable.ID, defaultRTName.(string))
	}

	if defaultACLName, ok := d.GetOk(isVPCDefaultNetworkACLName); ok {
		nwaclNameUpdate(sess, *vpc.DefaultNetworkACL.ID, defaultACLName.(string))
	}

	log.Printf("[INFO] VPC : %s", *vpc.ID)
	_, err = isWaitForVPCAvailable(sess, d.Id(), d.Timeout(schema.TimeoutCreate))
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("isWaitForVPCAvailable failed: %s", err.Error()), "ibm_is_vpc", "create")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	if dnsresolvertpeOk, ok := d.GetOk("dns.0.resolver.0.type"); ok {
		if dnsresolvertpeOk.(string) == "delegated" && ((d.Get("dns.0.resolver.0.vpc_id").(string) != "") || (d.Get("dns.0.resolver.0.vpc_crn").(string) != "")) {
			vpcId := d.Get("dns.0.resolver.0.vpc_id").(string)
			vpcCrn := d.Get("dns.0.resolver.0.vpc_crn").(string)
			createDnsBindings := &vpcv1.CreateVPCDnsResolutionBindingOptions{
				VPCID: vpc.ID,
			}
			vpcidentity := vpcv1.VPCIdentity{}
			if vpcId != "" {
				vpcidentity.ID = &vpcId
			}
			if vpcCrn != "" {
				vpcidentity.CRN = &vpcCrn
			}
			createDnsBindings.VPC = &vpcidentity
			if bindingNameOk, ok := d.GetOk("dns.0.resolver.0.dns_binding_name"); ok {
				bindingName := bindingNameOk.(string)
				createDnsBindings.Name = &bindingName
			}
			_, _, err := sess.CreateVPCDnsResolutionBindingWithContext(context, createDnsBindings)
			if err != nil {
				tfErr := flex.TerraformErrorf(err, fmt.Sprintf("CreateVPCDnsResolutionBindingWithContext failed: %s", err.Error()), "ibm_is_vpc", "create")
				log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
				return tfErr.GetDiag()
			}
			resolverType := "delegated"
			dnsPatch := &vpcv1.VpcdnsPatch{}
			resolver := &vpcv1.VpcdnsResolverPatch{
				Type: &resolverType,
			}
			vpcPatch := &vpcv1.VpcdnsResolverVPCPatch{}
			if vpcId != "" {
				vpcPatch.ID = &vpcId
			}
			if vpcCrn != "" {
				vpcPatch.CRN = &vpcCrn
			}
			resolver.VPC = vpcPatch
			dnsPatch.Resolver = resolver
			vpcPatchModel := &vpcv1.VPCPatch{}
			vpcPatchModel.Dns = dnsPatch
			vpcPatchModelAsPatch, err := vpcPatchModel.AsPatch()
			if err != nil {
				tfErr := flex.TerraformErrorf(err, fmt.Sprintf("vpcPatchModel.AsPatch() failed: %s", err.Error()), "ibm_is_vpc", "create")
				log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
				return tfErr.GetDiag()
			}
			updateVpcOptions := &vpcv1.UpdateVPCOptions{
				ID: vpc.ID,
			}
			updateVpcOptions.VPCPatch = vpcPatchModelAsPatch
			_, _, err = sess.UpdateVPCWithContext(context, updateVpcOptions)
			if err != nil {
				tfErr := flex.TerraformErrorf(err, fmt.Sprintf("UpdateVPCWithContext failed: %s", err.Error()), "ibm_is_vpc", "create")
				log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
				return tfErr.GetDiag()
			}
		}
	}

	if sgAclRules, ok := d.GetOk(isVPCNoSgAclRules); ok {
		sgAclRules := sgAclRules.(bool)
		if sgAclRules {
			deleteDefaultNetworkACLRules(sess, *vpc.ID)
			deleteDefaultSecurityGroupRules(sess, *vpc.ID)
		}
	}
	v := os.Getenv("IC_ENV_TAGS")
	if _, ok := d.GetOk(isVPCTags); ok || v != "" {
		oldList, newList := d.GetChange(isVPCTags)
		err = flex.UpdateGlobalTagsUsingCRN(oldList, newList, meta, *vpc.CRN, "", isVPCUserTagType)
		if err != nil {
			log.Printf(
				"Error on create of resource vpc (%s) tags: %s", d.Id(), err)
		}
	}
	if _, ok := d.GetOk(isVPCAccessTags); ok {
		oldList, newList := d.GetChange(isVPCAccessTags)
		err = flex.UpdateGlobalTagsUsingCRN(oldList, newList, meta, *vpc.CRN, "", isVPCAccessTagType)
		if err != nil {
			log.Printf(
				"Error on create of resource vpc (%s) access tags: %s", d.Id(), err)
		}
	}
	return nil
}

func isWaitForVPCAvailable(vpc *vpcv1.VpcV1, id string, timeout time.Duration) (interface{}, error) {
	log.Printf("Waiting for VPC (%s) to be available.", id)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{isVPCPending},
		Target:     []string{isVPCAvailable, isVPCFailed},
		Refresh:    isVPCRefreshFunc(vpc, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	return stateConf.WaitForState()
}

func deleteDefaultNetworkACLRules(sess *vpcv1.VpcV1, vpcID string) error {
	getVPCDefaultNetworkACLOptions := sess.NewGetVPCDefaultNetworkACLOptions(vpcID)
	result, detail, err := sess.GetVPCDefaultNetworkACL(getVPCDefaultNetworkACLOptions)
	if err != nil || result == nil {
		log.Printf("Error reading details of VPC Default Network ACL:%s", detail)
		return err
	}

	if result.Rules != nil {
		for _, sourceRule := range result.Rules {
			sourceRuleVal := sourceRule.(*vpcv1.NetworkACLRuleItemNetworkACLRuleProtocolAll)
			if sourceRuleVal.ID != nil {
				getNetworkAclRuleOptions := &vpcv1.GetNetworkACLRuleOptions{
					NetworkACLID: result.ID,
					ID:           sourceRuleVal.ID,
				}
				_, response, err := sess.GetNetworkACLRule(getNetworkAclRuleOptions)

				if err != nil {
					return fmt.Errorf("[ERROR] Error Getting Network ACL Rule  (%s): %s\n%s", *sourceRuleVal.ID, err, response)
				}

				deleteNetworkAclRuleOptions := &vpcv1.DeleteNetworkACLRuleOptions{
					NetworkACLID: result.ID,
					ID:           sourceRuleVal.ID,
				}
				response, err = sess.DeleteNetworkACLRule(deleteNetworkAclRuleOptions)
				if err != nil {
					return fmt.Errorf("[ERROR] Error Deleting Network ACL Rule : %s\n%s", err, response)
				}
			}
		}
	}
	return nil
}

func deleteDefaultSecurityGroupRules(sess *vpcv1.VpcV1, vpcID string) error {
	getVPCDefaultSecurityGroupOptions := sess.NewGetVPCDefaultSecurityGroupOptions(vpcID)
	result, detail, err := sess.GetVPCDefaultSecurityGroup(getVPCDefaultSecurityGroupOptions)
	if err != nil || result == nil {
		log.Printf("Error reading details of VPC Default Security Group:%s", detail)
		return err
	}

	if result.Rules != nil {
		for _, sourceRule := range result.Rules {
			sourceRuleVal := sourceRule.(*vpcv1.SecurityGroupRuleSecurityGroupRuleProtocolAll)
			if sourceRuleVal.ID != nil {
				getSecurityGroupRuleOptions := &vpcv1.GetSecurityGroupRuleOptions{
					SecurityGroupID: result.ID,
					ID:              sourceRuleVal.ID,
				}
				_, response, err := sess.GetSecurityGroupRule(getSecurityGroupRuleOptions)

				if err != nil {
					return fmt.Errorf("[ERROR] Error Getting Security Group Rule  (%s): %s\n%s", *sourceRuleVal.ID, err, response)
				}

				deleteSecurityGroupRuleOptions := &vpcv1.DeleteSecurityGroupRuleOptions{
					SecurityGroupID: result.ID,
					ID:              sourceRuleVal.ID,
				}
				response, err = sess.DeleteSecurityGroupRule(deleteSecurityGroupRuleOptions)
				if err != nil {
					return fmt.Errorf("[ERROR] Error Deleting Security Group Rule : %s\n%s", err, response)
				}
			}
		}
	}
	return nil
}

func isVPCRefreshFunc(vpc *vpcv1.VpcV1, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		getvpcOptions := &vpcv1.GetVPCOptions{
			ID: &id,
		}
		vpc, response, err := vpc.GetVPC(getvpcOptions)
		if err != nil {
			return nil, isVPCFailed, fmt.Errorf("[ERROR] Error getting VPC : %s\n%s", err, response)
		}

		if *vpc.Status == isVPCAvailable || *vpc.Status == isVPCFailed {
			return vpc, *vpc.Status, nil
		}

		return vpc, isVPCPending, nil
	}
}

func resourceIBMISVPCRead(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := d.Id()
	err := vpcGet(context, d, meta, id)
	if err != nil {
		return err
	}
	return nil
}

func vpcGet(context context.Context, d *schema.ResourceData, meta interface{}, id string) diag.Diagnostics {
	sess, err := vpcClient(meta)
	if err != nil {
		tfErr := flex.DiscriminatedTerraformErrorf(err, err.Error(), "ibm_is_vpc", "read", "initialize-client")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}
	getvpcOptions := &vpcv1.GetVPCOptions{
		ID: &id,
	}
	vpc, response, err := sess.GetVPCWithContext(context, getvpcOptions)
	if err != nil {
		if response != nil && response.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("GetVPCWithContext failed: %s", err.Error()), "ibm_is_vpc", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	// address prefixes

	vpcID := id // Assuming the VPC ID is stored in the resource ID

	// Fetch all address prefixes for the VPC
	startAdd := ""
	allRecs := []vpcv1.AddressPrefix{}
	for {
		listVpcAddressPrefixesOptions := &vpcv1.ListVPCAddressPrefixesOptions{
			VPCID: &vpcID,
		}

		if startAdd != "" {
			listVpcAddressPrefixesOptions.Start = &startAdd
		}

		addressPrefixCollection, _, err := sess.ListVPCAddressPrefixesWithContext(context, listVpcAddressPrefixesOptions)
		if err != nil {
			tfErr := flex.TerraformErrorf(err, fmt.Sprintf("ListVPCAddressPrefixesWithContext failed: %s", err.Error()), "ibm_is_vpc", "read")
			log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
			return tfErr.GetDiag()
		}

		allRecs = append(allRecs, addressPrefixCollection.AddressPrefixes...)
		startAdd = flex.GetNext(addressPrefixCollection.Next)
		if startAdd == "" {
			break
		}
	}

	// Process address prefixes
	defaultAddressPrefixes := map[string]string{}

	for _, prefix := range allRecs {
		zoneName := *prefix.Zone.Name
		cidr := *prefix.CIDR
		// Populate default_address_prefixes
		if *prefix.IsDefault {
			defaultAddressPrefixes[zoneName] = cidr
		}
	}

	// Set the default_address_prefixes attribute in the Terraform state
	if err := d.Set("default_address_prefixes", defaultAddressPrefixes); err != nil {
		return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting default_address_prefixes: %s", err), "ibm_is_vpc", "read", "set-default_address_prefixes").GetDiag()
	}
	if err = d.Set(isVPCName, *vpc.Name); err != nil {
		return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting name: %s", err), "ibm_is_vpc", "read", "set-name").GetDiag()
	}

	if err = d.Set(isVPCClassicAccess, *vpc.ClassicAccess); err != nil {
		return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting classic_access: %s", err), "ibm_is_vpc", "read", "set-classic_access").GetDiag()
	}

	if err = d.Set(isVPCStatus, *vpc.Status); err != nil {
		return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting status: %s", err), "ibm_is_vpc", "read", "set-status").GetDiag()
	}
	if vpc.DefaultNetworkACL != nil {
		log.Printf("[DEBUG] vpc default network acl is not null :%s", *vpc.DefaultNetworkACL.ID)
		if err = d.Set(isVPCDefaultNetworkACL, *vpc.DefaultNetworkACL.ID); err != nil {
			return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting default_network_acl: %s", err), "ibm_is_vpc", "read", "set-default_network_acl").GetDiag()
		}
		if err = d.Set(isVPCDefaultNetworkACLName, *vpc.DefaultNetworkACL.Name); err != nil {
			return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting default_network_acl_name: %s", err), "ibm_is_vpc", "read", "set-default_network_acl_name").GetDiag()
		}
		if err = d.Set(isVPCDefaultNetworkACLCRN, vpc.DefaultNetworkACL.CRN); err != nil {
			return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting default_network_acl_crn: %s", err), "ibm_is_vpc", "read", "set-default_network_acl_crn").GetDiag()
		}
	} else {
		log.Printf("[DEBUG] vpc default network acl is  null")
		d.Set(isVPCDefaultNetworkACL, nil)
	}
	if vpc.DefaultSecurityGroup != nil {
		if err = d.Set(isVPCDefaultSecurityGroup, *vpc.DefaultSecurityGroup.ID); err != nil {
			return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting default_security_group: %s", err), "ibm_is_vpc", "read", "set-default_security_group").GetDiag()
		}
		if err = d.Set(isVPCDefaultSecurityGroupName, *vpc.DefaultSecurityGroup.Name); err != nil {
			return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting default_security_group_name: %s", err), "ibm_is_vpc", "read", "set-default_security_group_name").GetDiag()
		}
		if err = d.Set(isVPCDefaultSecurityGroupCRN, vpc.DefaultSecurityGroup.CRN); err != nil {
			return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting default_security_group_crn: %s", err), "ibm_is_vpc", "read", "set-default_security_group_crn").GetDiag()
		}
	} else {
		d.Set(isVPCDefaultSecurityGroup, nil)
	}
	if vpc.DefaultRoutingTable != nil {
		if err = d.Set(isVPCDefaultRoutingTable, *vpc.DefaultRoutingTable.ID); err != nil {
			return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting default_routing_table: %s", err), "ibm_is_vpc", "read", "set-default_routing_table").GetDiag()
		}
		if err = d.Set(isVPCDefaultRoutingTableName, *vpc.DefaultRoutingTable.Name); err != nil {
			return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting default_routing_table_name: %s", err), "ibm_is_vpc", "read", "set-default_routing_table_name").GetDiag()
		}
		if vpc.DefaultRoutingTable.CRN != nil {
			if err = d.Set(isVPCDefaultRoutingTableCRN, *vpc.DefaultRoutingTable.CRN); err != nil {
				return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting default_routing_table_crn: %s", err), "ibm_is_vpc", "read", "set-default_routing_table_crn").GetDiag()
			}
		}
	}
	healthReasons := []map[string]interface{}{}
	if vpc.HealthReasons != nil {
		for _, modelItem := range vpc.HealthReasons {
			modelMap, err := dataSourceIBMIsVPCVPCHealthReasonToMap(&modelItem)
			if err != nil {
				return flex.DiscriminatedTerraformErrorf(err, err.Error(), "ibm_is_vpc", "read", "health_reasons-to-map").GetDiag()
			}
			healthReasons = append(healthReasons, modelMap)
		}
	}
	if err = d.Set("health_reasons", healthReasons); err != nil {
		return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting health_reasons: %s", err), "ibm_is_vpc", "read", "set-health_reasons").GetDiag()
	}

	if err = d.Set("health_state", vpc.HealthState); err != nil {
		return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting health_state: %s", err), "ibm_is_vpc", "read", "set-health_state").GetDiag()
	}
	if !core.IsNil(vpc.Dns) {
		vpcCrn := d.Get("dns.0.resolver.0.vpc_crn").(string)
		vpcId := d.Get("dns.0.resolver.0.vpc_id").(string)

		dnsMap, err := resourceIBMIsVPCVpcdnsToMap(vpc.Dns, vpcId, vpcCrn)
		if err != nil {
			return flex.DiscriminatedTerraformErrorf(err, err.Error(), "ibm_is_vpc", "read", "dns-to-map").GetDiag()
		}
		resolverMapArray := dnsMap["resolver"].([]map[string]interface{})
		resolverMap := resolverMapArray[0]
		if resolverMap["type"] != nil && ((resolverMap["vpc_id"] != nil) || (resolverMap["vpc_crn"] != nil)) {
			resType := resolverMap["type"].(*string)
			resVpc := resolverMap["vpc_id"].(string)
			resVpcCrn := resolverMap["vpc_crn"].(string)
			if *resType == "delegated" {
				listVPCDnsResolutionBindingOptions := &vpcv1.ListVPCDnsResolutionBindingsOptions{
					VPCID: vpc.ID,
				}

				pager, err := sess.NewVPCDnsResolutionBindingsPager(listVPCDnsResolutionBindingOptions)
				if err != nil {
					tfErr := flex.TerraformErrorf(err, fmt.Sprintf("NewVPCDnsResolutionBindingsPager failed: %s", err.Error()), "ibm_is_vpc", "read")
					log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
					return tfErr.GetDiag()
				}
				var allResults []vpcv1.VpcdnsResolutionBinding
				for pager.HasNext() {
					nextPage, err := pager.GetNext()
					if err != nil {
						tfErr := flex.TerraformErrorf(err, fmt.Sprintf("pager.GetNext() failed: %s", err.Error()), "ibm_is_vpc", "read")
						log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
						return tfErr.GetDiag()
					}
					allResults = append(allResults, nextPage...)
				}
				for _, binding := range allResults {
					if (*binding.VPC.ID == resVpc) || (*binding.VPC.CRN == resVpcCrn) {
						resolverMap["dns_binding_id"] = binding.ID
						resolverMap["dns_binding_name"] = binding.Name
						resolverMapArray[0] = resolverMap
						dnsMap["resolver"] = resolverMapArray
					}
				}
			}
		}
		if err = d.Set(isVPCDns, []map[string]interface{}{dnsMap}); err != nil {
			return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting dns: %s", err), "ibm_is_vpc", "read", "set-dns").GetDiag()
		}
	}
	tags, err := flex.GetGlobalTagsUsingCRN(meta, *vpc.CRN, "", isVPCUserTagType)
	if err != nil {
		log.Printf(
			"Error on get of resource vpc (%s) tags: %s", d.Id(), err)
	}

	if err = d.Set(isVPCTags, tags); err != nil {
		return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting tags: %s", err), "ibm_is_vpc", "read", "set-tags").GetDiag()
	}
	accesstags, err := flex.GetGlobalTagsUsingCRN(meta, *vpc.CRN, "", isVPCAccessTagType)
	if err != nil {
		log.Printf(
			"Error on get of resource vpc (%s) access tags: %s", d.Id(), err)
	}
	if err = d.Set(isVPCAccessTags, accesstags); err != nil {
		return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting access_tags: %s", err), "ibm_is_vpc", "read", "set-access_tags").GetDiag()
	}
	controller, err := flex.GetBaseController(meta)
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("GetBaseController failed: %s", err.Error()), "ibm_is_vpc", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}
	if err = d.Set(isVPCCRN, *vpc.CRN); err != nil {
		return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting crn: %s", err), "ibm_is_vpc", "read", "set-crn").GetDiag()
	}
	if err = d.Set(flex.ResourceControllerURL, controller+"/vpc-ext/network/vpcs"); err != nil {
		return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting resource_controller_url: %s", err), "ibm_is_vpc", "read", "set-resource_controller_url").GetDiag()
	}
	if err = d.Set(flex.ResourceName, *vpc.Name); err != nil {
		return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting resource_name: %s", err), "ibm_is_vpc", "read", "set-resource_name").GetDiag()
	}
	if err = d.Set(flex.ResourceCRN, *vpc.CRN); err != nil {
		return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting resource_crn: %s", err), "ibm_is_vpc", "read", "set-resource_crn").GetDiag()
	}
	if err = d.Set(flex.ResourceStatus, *vpc.Status); err != nil {
		return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting resource_status: %s", err), "ibm_is_vpc", "read", "set-resource_status").GetDiag()
	}
	if vpc.ResourceGroup != nil {
		if err = d.Set(isVPCResourceGroup, *vpc.ResourceGroup.ID); err != nil {
			return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting resource_group: %s", err), "ibm_is_vpc", "read", "set-resource_group").GetDiag()
		}
		if err = d.Set(flex.ResourceGroupName, *vpc.ResourceGroup.Name); err != nil {
			return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting resource_group_name: %s", err), "ibm_is_vpc", "read", "set-resource_group_name").GetDiag()
		}
	}

	//public address range
	publicAddressRanges := []map[string]interface{}{}
	for _, publicAddressRangesItem := range vpc.PublicAddressRanges {
		publicAddressRangesItemMap, err := ResourceIBMIsVPCPublicAddressRangeReferenceToMap(&publicAddressRangesItem)
		if err != nil {
			return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting public_address_ranges: %s", err), "ibm_is_vpc", "read", "set-public_address_ranges").GetDiag()
		}
		publicAddressRanges = append(publicAddressRanges, publicAddressRangesItemMap)
	}
	if err = d.Set("public_address_ranges", publicAddressRanges); err != nil {
		return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting public_address_ranges: %s", err), "ibm_is_vpc", "read", "set-public_address_ranges").GetDiag()
	}

	//set the cse ip addresses info
	if vpc.CseSourceIps != nil {
		cseSourceIpsList := make([]map[string]interface{}, 0)
		for _, sourceIP := range vpc.CseSourceIps {
			currentCseSourceIp := map[string]interface{}{}
			if sourceIP.IP != nil {
				currentCseSourceIp[isVPCDnsResolverManualServersAddress] = *sourceIP.IP.Address
				currentCseSourceIp["zone_name"] = *sourceIP.Zone.Name
				cseSourceIpsList = append(cseSourceIpsList, currentCseSourceIp)
			}
		}

		if err = d.Set(cseSourceAddresses, cseSourceIpsList); err != nil {
			return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting cse_source_addresses: %s", err), "ibm_is_vpc", "read", "set-cse_source_addresses").GetDiag()
		}
	}
	// set the subnets list
	start := ""
	allrecs := []vpcv1.Subnet{}
	for {
		options := &vpcv1.ListSubnetsOptions{}
		if start != "" {
			options.Start = &start
		}
		s, _, err := sess.ListSubnetsWithContext(context, options)
		if err != nil {
			tfErr := flex.TerraformErrorf(err, fmt.Sprintf("ListSubnetsWithContext failed: %s", err.Error()), "ibm_is_vpc", "read")
			log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
			return tfErr.GetDiag()
		}
		start = flex.GetNext(s.Next)
		allrecs = append(allrecs, s.Subnets...)
		if start == "" {
			break
		}
	}
	subnetsInfo := make([]map[string]interface{}, 0)
	for _, subnet := range allrecs {
		if *subnet.VPC.ID == d.Id() {
			l := map[string]interface{}{
				"name":                    *subnet.Name,
				"id":                      *subnet.ID,
				"status":                  *subnet.Status,
				"zone":                    *subnet.Zone.Name,
				totalIPV4AddressCount:     *subnet.TotalIpv4AddressCount,
				availableIPV4AddressCount: *subnet.AvailableIpv4AddressCount,
			}
			subnetsInfo = append(subnetsInfo, l)
		}
	}

	if err = d.Set(subnetsList, subnetsInfo); err != nil {
		return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting subnets: %s", err), "ibm_is_vpc", "read", "set-subnets").GetDiag()
	}

	//Set Security group list
	vpcid := d.Id()
	listSgOptions := &vpcv1.ListSecurityGroupsOptions{
		VPCID: &vpcid,
	}
	sgs, _, err := sess.ListSecurityGroupsWithContext(context, listSgOptions)
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("ListSecurityGroupsWithContext failed: %s", err.Error()), "ibm_is_vpc", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	securityGroupList := make([]map[string]interface{}, 0)

	for _, group := range sgs.SecurityGroups {
		if *group.VPC.ID == d.Id() {
			g := make(map[string]interface{})

			g[isVPCSecurityGroupName] = *group.Name
			g[isVPCSecurityGroupID] = *group.ID

			rules := make([]map[string]interface{}, 0)
			for _, sgrule := range group.Rules {
				switch reflect.TypeOf(sgrule).String() {
				case "*vpcv1.SecurityGroupRuleSecurityGroupRuleProtocolIcmp":
					{
						rule := sgrule.(*vpcv1.SecurityGroupRuleSecurityGroupRuleProtocolIcmp)
						r := make(map[string]interface{})
						if rule.Code != nil {
							r[isVPCSecurityGroupRuleCode] = int(*rule.Code)
						}
						if rule.Type != nil {
							r[isVPCSecurityGroupRuleType] = int(*rule.Type)
						}
						r[isVPCSecurityGroupRuleDirection] = *rule.Direction
						r[isVPCSecurityGroupRuleIPVersion] = *rule.IPVersion
						if rule.Protocol != nil {
							r[isVPCSecurityGroupRuleProtocol] = *rule.Protocol
						}
						r[isVPCSecurityGroupRuleID] = *rule.ID
						remote, ok := rule.Remote.(*vpcv1.SecurityGroupRuleRemote)
						if ok {
							if remote != nil && reflect.ValueOf(remote).IsNil() == false {
								if remote.ID != nil {
									r[isVPCSecurityGroupRuleRemote] = remote.ID
								} else if remote.Address != nil {
									r[isVPCSecurityGroupRuleRemote] = remote.Address
								} else if remote.CIDRBlock != nil {
									r[isVPCSecurityGroupRuleRemote] = remote.CIDRBlock
								}
							}
						}

						rules = append(rules, r)
					}

				case "*vpcv1.SecurityGroupRuleSecurityGroupRuleProtocolAll":
					{
						rule := sgrule.(*vpcv1.SecurityGroupRuleSecurityGroupRuleProtocolAll)
						r := make(map[string]interface{})
						r[isVPCSecurityGroupRuleDirection] = *rule.Direction
						r[isVPCSecurityGroupRuleIPVersion] = *rule.IPVersion
						if rule.Protocol != nil {
							r[isVPCSecurityGroupRuleProtocol] = *rule.Protocol
						}
						r[isVPCSecurityGroupRuleID] = *rule.ID
						remote, ok := rule.Remote.(*vpcv1.SecurityGroupRuleRemote)
						if ok {
							if remote != nil && reflect.ValueOf(remote).IsNil() == false {
								if remote.ID != nil {
									r[isVPCSecurityGroupRuleRemote] = remote.ID
								} else if remote.Address != nil {
									r[isVPCSecurityGroupRuleRemote] = remote.Address
								} else if remote.CIDRBlock != nil {
									r[isVPCSecurityGroupRuleRemote] = remote.CIDRBlock
								}
							}
						}
						rules = append(rules, r)
					}

				case "*vpcv1.SecurityGroupRuleSecurityGroupRuleProtocolTcpudp":
					{
						rule := sgrule.(*vpcv1.SecurityGroupRuleSecurityGroupRuleProtocolTcpudp)
						r := make(map[string]interface{})
						r[isVPCSecurityGroupRuleDirection] = *rule.Direction
						r[isVPCSecurityGroupRuleIPVersion] = *rule.IPVersion
						if rule.PortMin != nil {
							r[isVPCSecurityGroupRulePortMin] = int(*rule.PortMin)
						}
						if rule.PortMax != nil {
							r[isVPCSecurityGroupRulePortMax] = int(*rule.PortMax)
						}

						if rule.Protocol != nil {
							r[isVPCSecurityGroupRuleProtocol] = *rule.Protocol
						}

						r[isVPCSecurityGroupRuleID] = *rule.ID
						remote, ok := rule.Remote.(*vpcv1.SecurityGroupRuleRemote)
						if ok {
							if remote != nil && reflect.ValueOf(remote).IsNil() == false {
								if remote.ID != nil {
									r[isVPCSecurityGroupRuleRemote] = remote.ID
								} else if remote.Address != nil {
									r[isVPCSecurityGroupRuleRemote] = remote.Address
								} else if remote.CIDRBlock != nil {
									r[isVPCSecurityGroupRuleRemote] = remote.CIDRBlock
								}
							}
						}
						rules = append(rules, r)
					}
				}
			}
			g[isVPCSgRules] = rules
			securityGroupList = append(securityGroupList, g)
		}
	}

	if err = d.Set(isVPCSecurityGroupList, securityGroupList); err != nil {
		return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting security_group: %s", err), "ibm_is_vpc", "read", "set-security_group").GetDiag()
	}
	return nil
}

func resourceIBMISVPCUpdate(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := d.Id()

	name := ""
	hasChanged := false

	if d.HasChange(isVPCName) {
		name = d.Get(isVPCName).(string)
		hasChanged = true
	}
	err := vpcUpdate(context, d, meta, id, name, hasChanged)
	if err != nil {
		return err
	}
	return resourceIBMISVPCRead(context, d, meta)
}

func vpcUpdate(context context.Context, d *schema.ResourceData, meta interface{}, id, name string, hasChanged bool) diag.Diagnostics {
	sess, err := vpcClient(meta)
	if err != nil {
		tfErr := flex.DiscriminatedTerraformErrorf(err, err.Error(), "ibm_is_vpc", "update", "initialize-client")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	if d.HasChange(isVPCTags) {
		getvpcOptions := &vpcv1.GetVPCOptions{
			ID: &id,
		}
		vpc, _, err := sess.GetVPCWithContext(context, getvpcOptions)
		if err != nil {
			tfErr := flex.TerraformErrorf(err, fmt.Sprintf("GetVPCWithContext failed: %s", err.Error()), "ibm_is_vpc", "update")
			log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
			return tfErr.GetDiag()
		}
		oldList, newList := d.GetChange(isVPCTags)
		err = flex.UpdateGlobalTagsUsingCRN(oldList, newList, meta, *vpc.CRN, "", isVPCUserTagType)
		if err != nil {
			log.Printf(
				"Error on update of resource vpc (%s) tags: %s", d.Id(), err)
		}
	}
	if d.HasChange(isVPCAccessTags) {
		getvpcOptions := &vpcv1.GetVPCOptions{
			ID: &id,
		}
		vpc, _, err := sess.GetVPCWithContext(context, getvpcOptions)
		if err != nil {
			tfErr := flex.TerraformErrorf(err, fmt.Sprintf("GetVPCWithContext failed: %s", err.Error()), "ibm_is_vpc", "update")
			log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
			return tfErr.GetDiag()
		}
		oldList, newList := d.GetChange(isVPCAccessTags)
		err = flex.UpdateGlobalTagsUsingCRN(oldList, newList, meta, *vpc.CRN, "", isVPCAccessTagType)
		if err != nil {
			log.Printf(
				"Error on update of resource VPC (%s) access tags: %s", d.Id(), err)
		}
	}

	if d.HasChange(isVPCDefaultSecurityGroupName) {
		if defaultSGName, ok := d.GetOk(isVPCDefaultSecurityGroupName); ok {
			sgNameUpdate(sess, d.Get(isVPCDefaultSecurityGroup).(string), defaultSGName.(string))
		}
	}
	if d.HasChange(isVPCDefaultRoutingTableName) {
		if defaultRTName, ok := d.GetOk(isVPCDefaultRoutingTableName); ok {
			rtNameUpdate(sess, id, d.Get(isVPCDefaultRoutingTable).(string), defaultRTName.(string))
		}
	}
	if d.HasChange(isVPCDefaultNetworkACLName) {
		if defaultACLName, ok := d.GetOk(isVPCDefaultNetworkACLName); ok {
			nwaclNameUpdate(sess, d.Get(isVPCDefaultNetworkACL).(string), defaultACLName.(string))
		}
	}
	hasDnsChanged := false
	isDnsResolverVPCIDNull := false
	isDnsResolverVPCID := ""
	isDnsResolverVPCCrn := ""
	isDnsResolverVPCCrnNull := false
	isDnsResolverManualServerChange := false
	isDnsResolverManualServerEtag := ""
	deleteBinding := false
	deleteDnsBindings := &vpcv1.DeleteVPCDnsResolutionBindingOptions{}
	var dnsPatch *vpcv1.VpcdnsPatch
	if d.HasChange(isVPCDns) {
		dnsPatch = &vpcv1.VpcdnsPatch{}
		if d.HasChange("dns.0.enable_hub") {
			_, newEH := d.GetChange("dns.0.enable_hub")
			dnsPatch.EnableHub = core.BoolPtr(newEH.(bool))
			hasDnsChanged = true
		}
		if d.HasChange("dns.0.resolver") {
			_, newResolver := d.GetChange("dns.0.resolver")
			if d.HasChange("dns.0.resolver.0.dns_binding_name") && (d.Get("dns.0.resolver.0.dns_binding_id") != nil && d.Get("dns.0.resolver.0.dns_binding_id").(string) != "") && (d.Get("dns.0.resolver.0.dns_binding_name").(string) != "null" || d.Get("dns.0.resolver.0.dns_binding_name").(string) != "") {
				dnsBindingName := d.Get("dns.0.resolver.0.dns_binding_name").(string)
				dnsBindingId := d.Get("dns.0.resolver.0.dns_binding_id").(string)
				vpcdnsResolutionBindingPatch := &vpcv1.VpcdnsResolutionBindingPatch{
					Name: &dnsBindingName,
				}
				vpcdnsResolutionBindingPatchAsPatch, _ := vpcdnsResolutionBindingPatch.AsPatch()
				updateVPCDnsResolutionBinding := &vpcv1.UpdateVPCDnsResolutionBindingOptions{
					ID:                           &dnsBindingId,
					VPCID:                        core.StringPtr(d.Id()),
					VpcdnsResolutionBindingPatch: vpcdnsResolutionBindingPatchAsPatch,
				}
				_, _, err := sess.UpdateVPCDnsResolutionBindingWithContext(context, updateVPCDnsResolutionBinding)
				if err != nil {
					tfErr := flex.TerraformErrorf(err, fmt.Sprintf("UpdateVPCDnsResolutionBindingWithContext failed: %s", err.Error()), "ibm_is_vpc", "update")
					log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
					return tfErr.GetDiag()
				}
			}
			if newResolver != nil && len(newResolver.([]interface{})) > 0 {
				ResolverModel := &vpcv1.VpcdnsResolverPatch{}
				if d.HasChange("dns.0.resolver.0.manual_servers") {

					// getting etag
					getVpcOptions := &vpcv1.GetVPCOptions{
						ID: &id,
					}
					_, response, err := sess.GetVPCWithContext(context, getVpcOptions)
					if err != nil {
						tfErr := flex.TerraformErrorf(err, fmt.Sprintf("GetVPCWithContext failed: %s", err.Error()), "ibm_is_vpc", "update")
						log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
						return tfErr.GetDiag()
					}
					isDnsResolverManualServerChange = true
					isDnsResolverManualServerEtag = response.Headers.Get("ETag") // Getting Etag from the response headers.

					_, newResolverManualServers := d.GetChange("dns.0.resolver.0.manual_servers")

					if newResolverManualServers != nil {
						manualServers := []vpcv1.DnsServerPrototype{}
						for _, manualServersItem := range newResolverManualServers.(*schema.Set).List() {
							manualServersItemModel, err := resourceIBMIsVPCMapToDnsServerPrototype(manualServersItem.(map[string]interface{}))
							if err != nil {
								tfErr := flex.TerraformErrorf(err, fmt.Sprintf("resourceIBMIsVPCMapToDnsServerPrototype failed: %s", err.Error()), "ibm_is_vpc", "update")
								log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
								return tfErr.GetDiag()
							}
							manualServers = append(manualServers, *manualServersItemModel)
						}
						ResolverModel.ManualServers = manualServers
					}

				}
				if d.HasChange("dns.0.resolver.0.type") {
					oldResolverType, newResolverType := d.GetChange("dns.0.resolver.0.type")
					if newResolverType != nil && newResolverType.(string) != "" {
						ResolverModel.Type = core.StringPtr(newResolverType.(string))
					}
					if oldResolverType != nil && newResolverType != nil && oldResolverType.(string) != "" && newResolverType.(string) != "" {
						if oldResolverType.(string) == "system" && newResolverType.(string) == "delegated" {
							vpcId := d.Get("dns.0.resolver.0.vpc_id").(string)
							vpcCrn := d.Get("dns.0.resolver.0.vpc_crn").(string)
							createDnsBindings := &vpcv1.CreateVPCDnsResolutionBindingOptions{
								VPCID: core.StringPtr(d.Id()),
							}
							vpcidentity := &vpcv1.VPCIdentity{}
							if vpcId != "" {
								vpcidentity.ID = &vpcId
							}
							if vpcCrn != "" {
								vpcidentity.CRN = &vpcCrn
							}
							createDnsBindings.VPC = vpcidentity

							if bindingNameOk, ok := d.GetOk("dns.0.resolver.0.dns_binding_name"); ok {
								bindingName := bindingNameOk.(string)
								createDnsBindings.Name = &bindingName
							}
							_, response, err := sess.CreateVPCDnsResolutionBindingWithContext(context, createDnsBindings)
							if err != nil {
								exitError := false
								log.Printf("[DEBUG] CreateVPCDnsResolutionBindingWithContext failed %s\n%s", err, response)
								if response.StatusCode == 400 && strings.Contains(err.Error(), "This VPC already contains DNS Resolution Bindings") {
									listVPCDnsResolutionBindingOptions := &vpcv1.ListVPCDnsResolutionBindingsOptions{
										VPCID: createDnsBindings.VPCID,
									}

									pager, err := sess.NewVPCDnsResolutionBindingsPager(listVPCDnsResolutionBindingOptions)
									if err != nil {
										tfErr := flex.TerraformErrorf(err, fmt.Sprintf("NewVPCDnsResolutionBindingsPager(CreateVPCDnsResolutionBindingWithContext) failed: %s", err.Error()), "ibm_is_vpc", "update")
										log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
										return tfErr.GetDiag()
									}
									var allResults []vpcv1.VpcdnsResolutionBinding
									for pager.HasNext() {
										nextPage, err := pager.GetNext()
										if err != nil {
											tfErr := flex.TerraformErrorf(err, fmt.Sprintf("pager.HasNext()(CreateVPCDnsResolutionBindingWithContext) failed: %s", err.Error()), "ibm_is_vpc", "update")
											log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
											return tfErr.GetDiag()
										}
										allResults = append(allResults, nextPage...)
									}
									for _, binding := range allResults {
										if *binding.VPC.ID == vpcId {
											log.Printf("[DEBUG] CreateVPCDnsResolutionBindingWithContext failed but binding to same vpc exists %s\n%s", err, response)
											exitError = true
										}
									}
								}
								if !exitError {
									tfErr := flex.TerraformErrorf(err, fmt.Sprintf("CreateVPCDnsResolutionBindingWithContext failed: %s", err.Error()), "ibm_is_vpc", "update")
									log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
									return tfErr.GetDiag()
								}
							}
						}
						if _, ok := d.GetOk("dns.0.resolver.0.dns_binding_name"); ok {
						}
						if d.HasChange("dns.0.resolver.0.dns_binding_name") && d.Get("dns.0.resolver.0.dns_binding_name").(string) == "null" {
							dnsid := d.Get("dns.0.resolver.0.dns_binding_id").(string)
							deleteBinding = true
							deleteDnsBindings = &vpcv1.DeleteVPCDnsResolutionBindingOptions{
								VPCID: core.StringPtr(d.Id()),
								ID:    &dnsid,
							}
						}
					}
					hasDnsChanged = true
				}
				if d.HasChange("dns.0.resolver.0.vpc_id") {
					_, newResolverVpc := d.GetChange("dns.0.resolver.0.vpc_id")
					if newResolverVpc != nil && newResolverVpc.(string) != "" {
						isDnsResolverVPCID = newResolverVpc.(string)
						if isDnsResolverVPCID == "null" {
							isDnsResolverVPCIDNull = true
							var nullStringPtr *string
							ResolverModel.VPC = &vpcv1.VpcdnsResolverVPCPatch{
								ID: nullStringPtr,
							}
						} else {
							ResolverModel.VPC = &vpcv1.VpcdnsResolverVPCPatch{
								ID: &isDnsResolverVPCID,
							}
						}
					}
					hasDnsChanged = true
				}
				if d.HasChange("dns.0.resolver.0.vpc_crn") {
					_, newResolverVpc := d.GetChange("dns.0.resolver.0.vpc_crn")
					if newResolverVpc != nil && newResolverVpc.(string) != "" {
						isDnsResolverVPCCrn = newResolverVpc.(string)
						if isDnsResolverVPCCrn == "null" {
							isDnsResolverVPCCrnNull = true
							var nullStringPtr *string
							ResolverModel.VPC = &vpcv1.VpcdnsResolverVPCPatch{
								CRN: nullStringPtr,
							}
						} else {
							ResolverModel.VPC = &vpcv1.VpcdnsResolverVPCPatch{
								CRN: &isDnsResolverVPCCrn,
							}
						}
					}
					hasDnsChanged = true
				}
				dnsPatch.Resolver = ResolverModel
			}
		}
	}
	if hasChanged || hasDnsChanged {
		updateVpcOptions := &vpcv1.UpdateVPCOptions{
			ID: &id,
		}
		vpcPatchModel := &vpcv1.VPCPatch{}
		if hasChanged {
			vpcPatchModel.Name = &name
		}
		if hasDnsChanged {
			vpcPatchModel.Dns = dnsPatch
		}
		if isDnsResolverManualServerChange && isDnsResolverManualServerEtag != "" {
			updateVpcOptions.IfMatch = &isDnsResolverManualServerEtag // if-Match or Etag Change for Patch
		}
		vpcPatch, err := vpcPatchModel.AsPatch()
		if err != nil {
			tfErr := flex.TerraformErrorf(err, fmt.Sprintf("vpcPatchModel.AsPatch() failed: %s", err.Error()), "ibm_is_vpc", "update")
			log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
			return tfErr.GetDiag()
		}
		if isDnsResolverVPCCrnNull || isDnsResolverVPCIDNull {
			dnsMap := vpcPatch["dns"].(map[string]interface{})
			resolverMap := dnsMap["resolver"].(map[string]interface{})
			resolverMap["vpc"] = nil
			dnsMap["resolver"] = resolverMap
			vpcPatch["dns"] = dnsMap
		}

		updateVpcOptions.VPCPatch = vpcPatch
		_, response, err := sess.UpdateVPCWithContext(context, updateVpcOptions)
		if err != nil {
			responsestring := strings.ToLower(response.String())
			if strings.Contains(strings.ToLower(err.Error()), strings.ToLower("The supplied header is not supported for this request")) && strings.Contains(responsestring, "bad_header") && strings.Contains(responsestring, strings.ToLower("If-Match")) {
				log.Printf("[DEBUG] retrying update vpc without If-Match")
				updateVpcOptions.IfMatch = nil
				_, _, nestederr := sess.UpdateVPCWithContext(context, updateVpcOptions)
				if nestederr != nil {
					tfErr := flex.TerraformErrorf(err, fmt.Sprintf("UpdateVPCWithContext(retry) failed: %s", err.Error()), "ibm_is_vpc", "update")
					log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
					return tfErr.GetDiag()
				}
			} else {
				tfErr := flex.TerraformErrorf(err, fmt.Sprintf("UpdateVPCWithContext failed: %s", err.Error()), "ibm_is_vpc", "update")
				log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
				return tfErr.GetDiag()
			}
		}
		if deleteBinding && *deleteDnsBindings.VPCID != "" {
			_, _, err := sess.DeleteVPCDnsResolutionBindingWithContext(context, deleteDnsBindings)
			if err != nil {
				tfErr := flex.TerraformErrorf(err, fmt.Sprintf("DeleteVPCDnsResolutionBindingWithContext failed: %s", err.Error()), "ibm_is_vpc", "update")
				log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
				return tfErr.GetDiag()
			}
		}
		if isDnsResolverVPCCrnNull || isDnsResolverVPCIDNull {

			dnsList := make([]map[string]interface{}, 0)
			currentDns := map[string]interface{}{}
			currentResolverList := make([]map[string]interface{}, 0)
			currentResolver := map[string]interface{}{}
			currentResolver["vpc_id"] = isDnsResolverVPCID
			currentResolver["vpc_crn"] = isDnsResolverVPCCrn
			currentResolverList = append(currentResolverList, currentResolver)
			currentDns["resolver"] = currentResolverList
			dnsList = append(dnsList, currentDns)
			d.Set("dns", dnsList)
		}
	}
	return nil
}

func resourceIBMISVPCDelete(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := d.Id()
	err := vpcDelete(context, d, meta, id)
	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func vpcDelete(context context.Context, d *schema.ResourceData, meta interface{}, id string) diag.Diagnostics {
	sess, err := vpcClient(meta)
	if err != nil {
		tfErr := flex.DiscriminatedTerraformErrorf(err, err.Error(), "ibm_is_vpc", "delete", "initialize-client")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	getVpcOptions := &vpcv1.GetVPCOptions{
		ID: &id,
	}
	_, response, err := sess.GetVPCWithContext(context, getVpcOptions)
	if err != nil {
		if response != nil && response.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("GetVPCWithContext failed: %s", err.Error()), "ibm_is_vpc", "delete")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	deletevpcOptions := &vpcv1.DeleteVPCOptions{
		ID: &id,
	}
	response, err = sess.DeleteVPCWithContext(context, deletevpcOptions)
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("DeleteVPCWithContext failed: %s", err.Error()), "ibm_is_vpc", "delete")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}
	_, err = isWaitForVPCDeleted(sess, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("isWaitForVPCDeleted failed: %s", err.Error()), "ibm_is_vpc", "delete")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}
	d.SetId("")
	return nil
}

func isWaitForVPCDeleted(vpc *vpcv1.VpcV1, id string, timeout time.Duration) (interface{}, error) {
	log.Printf("Waiting for VPC (%s) to be deleted.", id)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"retry", isVPCDeleting},
		Target:     []string{isVPCDeleted, isVPCFailed},
		Refresh:    isVPCDeleteRefreshFunc(vpc, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	return stateConf.WaitForState()
}

func isVPCDeleteRefreshFunc(vpc *vpcv1.VpcV1, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		log.Printf("[DEBUG] is vpc delete function here")
		getvpcOptions := &vpcv1.GetVPCOptions{
			ID: &id,
		}
		vpc, response, err := vpc.GetVPC(getvpcOptions)
		if err != nil {
			if response != nil && response.StatusCode == 404 {
				return vpc, isVPCDeleted, nil
			}
			return nil, isVPCFailed, fmt.Errorf("[ERROR] The VPC %s failed to delete: %s\n%s", id, err, response)
		}

		return vpc, isVPCDeleting, nil
	}
}

func resourceIBMISVPCExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	id := d.Id()
	exists, err := vpcExists(d, meta, id)
	return exists, err
}

func vpcExists(d *schema.ResourceData, meta interface{}, id string) (bool, error) {
	sess, err := vpcClient(meta)
	if err != nil {
		tfErr := flex.DiscriminatedTerraformErrorf(err, err.Error(), "ibm_is_vpc", "exists", "initialize-client")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return false, tfErr
	}
	getvpcOptions := &vpcv1.GetVPCOptions{
		ID: &id,
	}
	_, response, err := sess.GetVPC(getvpcOptions)
	if err != nil {
		if response != nil && response.StatusCode == 404 {
			return false, nil
		}
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("GetVPC failed: %s", err.Error()), "ibm_is_vpc", "exists")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return false, tfErr
	}
	return true, nil
}

// func ResourceIBMVPCHash(v interface{}) int {
// 	var buf bytes.Buffer
// 	buf.WriteString(fmt.Sprintf("%s",
// 		strings.ToLower(v.(string))))
// 	return conns.String(buf.String())
// }

func nwaclNameUpdate(sess *vpcv1.VpcV1, id, name string) error {
	updateNetworkACLOptions := &vpcv1.UpdateNetworkACLOptions{
		ID: &id,
	}
	networkACLPatchModel := &vpcv1.NetworkACLPatch{
		Name: &name,
	}
	networkACLPatch, err := networkACLPatchModel.AsPatch()
	if err != nil {
		return fmt.Errorf("[ERROR] Error calling asPatch for NetworkACLPatch: %s", err)
	}
	updateNetworkACLOptions.NetworkACLPatch = networkACLPatch
	_, response, err := sess.UpdateNetworkACL(updateNetworkACLOptions)
	if err != nil {
		return fmt.Errorf("[ERROR] Error Updating Network ACL(%s) name : %s\n%s", id, err, response)
	}
	return nil
}

func sgNameUpdate(sess *vpcv1.VpcV1, id, name string) error {
	updateSecurityGroupOptions := &vpcv1.UpdateSecurityGroupOptions{
		ID: &id,
	}
	securityGroupPatchModel := &vpcv1.SecurityGroupPatch{
		Name: &name,
	}
	securityGroupPatch, err := securityGroupPatchModel.AsPatch()
	if err != nil {
		return fmt.Errorf("[ERROR] Error calling asPatch for SecurityGroupPatch: %s", err)
	}
	updateSecurityGroupOptions.SecurityGroupPatch = securityGroupPatch
	_, response, err := sess.UpdateSecurityGroup(updateSecurityGroupOptions)
	if err != nil {
		return fmt.Errorf("[ERROR] Error Updating Security Group name : %s\n%s", err, response)
	}
	return nil
}

func rtNameUpdate(sess *vpcv1.VpcV1, vpcID, id, name string) error {
	updateVpcRoutingTableOptions := new(vpcv1.UpdateVPCRoutingTableOptions)
	updateVpcRoutingTableOptions.VPCID = &vpcID
	updateVpcRoutingTableOptions.ID = &id
	routingTablePatchModel := new(vpcv1.RoutingTablePatch)
	routingTablePatchModel.Name = &name
	routingTablePatchModelAsPatch, asPatchErr := routingTablePatchModel.AsPatch()
	if asPatchErr != nil {
		return fmt.Errorf("[ERROR] Error calling asPatch for RoutingTablePatchModel: %s", asPatchErr)
	}
	updateVpcRoutingTableOptions.RoutingTablePatch = routingTablePatchModelAsPatch
	_, response, err := sess.UpdateVPCRoutingTable(updateVpcRoutingTableOptions)
	if err != nil {
		return fmt.Errorf("[ERROR] Error Updating Routing table name %s\n%s", err, response)
	}
	return nil
}

func suppressNullAddPrefix(k, old, new string, d *schema.ResourceData) bool {
	// During import
	if old == "" && d.Id() != "" {
		return true
	}
	return false
}

func suppressNullVPC(k, old, new string, d *schema.ResourceData) bool {
	if new != old && new == "null" && old == "" && d.Id() != "" {
		return true
	}
	return false
}

func suppressNullDnsBindingName(k, old, new string, d *schema.ResourceData) bool {
	if new != old && new == "null" && old == "" && d.Id() != "" {
		return true
	}
	return false
}

func hashManualServersList(v interface{}) int {
	var buf bytes.Buffer
	a := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", a["address"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", a["zone_affinity"].(string)))
	return conns.String(buf.String())
}

// for create dns
func resourceIBMIsVPCMapToVpcdnsPrototype(modelMap map[string]interface{}) (*vpcv1.VpcdnsPrototype, error) {
	model := &vpcv1.VpcdnsPrototype{}
	if modelMap["enable_hub"] != nil {
		model.EnableHub = core.BoolPtr(modelMap["enable_hub"].(bool))
	}
	if modelMap["resolver"] != nil && len(modelMap["resolver"].([]interface{})) > 0 {
		ResolverModel, err := resourceIBMIsVPCMapToVpcdnsResolverPrototype(modelMap["resolver"].([]interface{})[0].(map[string]interface{}))
		if err != nil {
			return model, err
		}
		model.Resolver = ResolverModel
	}
	return model, nil
}

func resourceIBMIsVPCMapToVpcdnsResolverPrototype(modelMap map[string]interface{}) (vpcv1.VpcdnsResolverPrototypeIntf, error) {
	model := &vpcv1.VpcdnsResolverPrototype{}
	if modelMap["type"] != nil && modelMap["type"].(string) != "" {
		if modelMap["type"].(string) == "delegated" {
			model.Type = core.StringPtr("system")
		} else {
			model.Type = core.StringPtr(modelMap["type"].(string))
		}
	}
	if modelMap["manual_servers"] != nil && modelMap["manual_servers"].(*schema.Set).Len() > 0 {
		model.Type = core.StringPtr("manual")
		manualServers := []vpcv1.DnsServerPrototype{}
		for _, manualServersItem := range modelMap["manual_servers"].(*schema.Set).List() {
			manualServersItemModel, err := resourceIBMIsVPCMapToDnsServerPrototype(manualServersItem.(map[string]interface{}))
			if err != nil {
				return model, err
			}
			manualServers = append(manualServers, *manualServersItemModel)
		}
		model.ManualServers = manualServers
	}
	return model, nil
}

func resourceIBMIsVPCMapToDnsServerPrototype(modelMap map[string]interface{}) (*vpcv1.DnsServerPrototype, error) {
	model := &vpcv1.DnsServerPrototype{}
	model.Address = core.StringPtr(modelMap["address"].(string))
	if modelMap[isVPCDnsResolverManualServersZoneAffinity] != nil && modelMap[isVPCDnsResolverManualServersZoneAffinity].(string) != "" {
		ZoneAffinityModel := &vpcv1.ZoneIdentity{
			Name: core.StringPtr(modelMap[isVPCDnsResolverManualServersZoneAffinity].(string)),
		}
		model.ZoneAffinity = ZoneAffinityModel
	}
	return model, nil
}

// for dns read

func resourceIBMIsVPCVpcdnsToMap(model *vpcv1.Vpcdns, vpcId, vpcCrn string) (map[string]interface{}, error) {
	modelMap := make(map[string]interface{})
	modelMap["enable_hub"] = model.EnableHub
	modelMap["resolution_binding_count"] = flex.IntValue(model.ResolutionBindingCount)
	resolverMap, err := resourceIBMIsVPCVpcdnsResolverToMap(model.Resolver, vpcId, vpcCrn)
	if err != nil {
		return modelMap, err
	}
	modelMap["resolver"] = []map[string]interface{}{resolverMap}

	return modelMap, nil
}

func resourceIBMIsVPCVpcdnsResolverToMap(model vpcv1.VpcdnsResolverIntf, vpcId, vpcCrn string) (map[string]interface{}, error) {
	if _, ok := model.(*vpcv1.VpcdnsResolverTypeDelegated); ok {
		return resourceIBMIsVPCVpcdnsResolverTypeDelegatedToMap(model.(*vpcv1.VpcdnsResolverTypeDelegated), vpcId, vpcCrn)
	} else if _, ok := model.(*vpcv1.VpcdnsResolverTypeManual); ok {
		return resourceIBMIsVPCVpcdnsResolverTypeManualToMap(model.(*vpcv1.VpcdnsResolverTypeManual), vpcId, vpcCrn)
	} else if _, ok := model.(*vpcv1.VpcdnsResolverTypeSystem); ok {
		return resourceIBMIsVPCVpcdnsResolverTypeSystemToMap(model.(*vpcv1.VpcdnsResolverTypeSystem), vpcId, vpcCrn)
	} else if _, ok := model.(*vpcv1.VpcdnsResolver); ok {
		modelMap := make(map[string]interface{})
		model := model.(*vpcv1.VpcdnsResolver)
		servers := []map[string]interface{}{}
		for _, serversItem := range model.Servers {
			serversItemMap, err := resourceIBMIsVPCDnsServerToMap(&serversItem)
			if err != nil {
				return modelMap, err
			}
			servers = append(servers, serversItemMap)
		}
		modelMap["servers"] = servers
		modelMap["type"] = model.Type
		if model.VPC != nil {
			modelMap["vpc_id"] = vpcId
			modelMap["vpc_crn"] = vpcCrn
			modelMap["vpc_name"] = *model.VPC.Name
			if model.VPC.Remote != nil {
				if model.VPC.Remote.Account != nil {
					modelMap["vpc_remote_account_id"] = *model.VPC.Remote.Account.ID
				}
				if model.VPC.Remote.Region != nil {
					modelMap["vpc_remote_region"] = *model.VPC.Remote.Region.Name
				}
			}

		}

		if model.ManualServers != nil {
			manualServers := []map[string]interface{}{}
			for _, manualServersItem := range model.ManualServers {
				manualServersItemMap, err := resourceIBMIsVPCDnsServerToMap(&manualServersItem)
				if err != nil {
					return modelMap, err
				}
				manualServers = append(manualServers, manualServersItemMap)
			}
			modelMap["manual_servers"] = manualServers
		}
		if model.Configuration != nil {
			modelMap["configuration"] = model.Configuration
		}
		return modelMap, nil
	} else {
		return nil, fmt.Errorf("Unrecognized vpcv1.VpcdnsResolverIntf subtype encountered")
	}
}

func resourceIBMIsVPCDnsServerToMap(model *vpcv1.DnsServer) (map[string]interface{}, error) {
	modelMap := make(map[string]interface{})
	modelMap["address"] = model.Address
	if model.ZoneAffinity != nil {
		zoneAffinity := *model.ZoneAffinity.Name
		modelMap["zone_affinity"] = zoneAffinity
	}
	return modelMap, nil
}

func resourceIBMIsVPCVpcdnsResolverTypeDelegatedToMap(model *vpcv1.VpcdnsResolverTypeDelegated, vpcId, vpcCrn string) (map[string]interface{}, error) {
	modelMap := make(map[string]interface{})
	servers := []map[string]interface{}{}
	for _, serversItem := range model.Servers {
		serversItemMap, err := resourceIBMIsVPCDnsServerToMap(&serversItem)
		if err != nil {
			return modelMap, err
		}
		servers = append(servers, serversItemMap)
	}
	modelMap["servers"] = servers
	modelMap["type"] = model.Type
	if model.VPC != nil {
		modelMap["vpc_id"] = vpcId
		modelMap["vpc_crn"] = vpcCrn
		modelMap["vpc_name"] = *model.VPC.Name
		if model.VPC.Remote != nil {
			if model.VPC.Remote.Account != nil {
				modelMap["vpc_remote_account_id"] = *model.VPC.Remote.Account.ID
			}
			if model.VPC.Remote.Region != nil {
				modelMap["vpc_remote_region"] = *model.VPC.Remote.Region.Name
			}
		}
	}
	return modelMap, nil
}

func resourceIBMIsVPCVpcdnsResolverTypeManualToMap(model *vpcv1.VpcdnsResolverTypeManual, vpcId, vpcCrn string) (map[string]interface{}, error) {
	modelMap := make(map[string]interface{})
	servers := []map[string]interface{}{}
	for _, serversItem := range model.Servers {
		serversItemMap, err := resourceIBMIsVPCDnsServerToMap(&serversItem)
		if err != nil {
			return modelMap, err
		}
		servers = append(servers, serversItemMap)
	}
	modelMap["servers"] = servers
	manualServers := []map[string]interface{}{}
	for _, manualServersItem := range model.ManualServers {
		manualServersItemMap, err := resourceIBMIsVPCDnsServerToMap(&manualServersItem)
		if err != nil {
			return modelMap, err
		}
		manualServers = append(manualServers, manualServersItemMap)
	}
	modelMap["manual_servers"] = manualServers
	modelMap["type"] = model.Type
	if vpcId != "" {
		modelMap["vpc_id"] = vpcId
	}
	if vpcCrn != "" {
		modelMap["vpc_crn"] = vpcCrn
	}
	return modelMap, nil
}

func resourceIBMIsVPCVpcdnsResolverTypeSystemToMap(model *vpcv1.VpcdnsResolverTypeSystem, vpcId, vpcCrn string) (map[string]interface{}, error) {
	modelMap := make(map[string]interface{})
	servers := []map[string]interface{}{}
	for _, serversItem := range model.Servers {
		serversItemMap, err := resourceIBMIsVPCDnsServerToMap(&serversItem)
		if err != nil {
			return modelMap, err
		}
		servers = append(servers, serversItemMap)
	}
	modelMap["servers"] = servers
	modelMap["configuration"] = model.Configuration
	modelMap["type"] = model.Type
	if vpcId != "" {
		modelMap["vpc_id"] = vpcId
	}
	if vpcCrn != "" {
		modelMap["vpc_crn"] = vpcCrn
	}
	return modelMap, nil
}

func resourceIBMIsVPCVPCReferenceDnsResolverContextToMap(model *vpcv1.VPCReferenceDnsResolverContext) (map[string]interface{}, error) {
	modelMap := make(map[string]interface{})
	modelMap["crn"] = model.CRN
	if model.Deleted != nil {
		deletedMap, err := resourceIBMIsVPCVPCReferenceDnsResolverContextDeletedToMap(model.Deleted)
		if err != nil {
			return modelMap, err
		}
		modelMap["deleted"] = []map[string]interface{}{deletedMap}
	}
	modelMap["href"] = model.Href
	modelMap["id"] = model.ID
	modelMap["name"] = model.Name
	if model.Remote != nil {
		remoteMap, err := resourceIBMIsVPCVPCRemoteToMap(model.Remote)
		if err != nil {
			return modelMap, err
		}
		modelMap["remote"] = []map[string]interface{}{remoteMap}
	}
	modelMap["resource_type"] = model.ResourceType
	return modelMap, nil
}

func resourceIBMIsVPCVPCReferenceDnsResolverContextDeletedToMap(model *vpcv1.Deleted) (map[string]interface{}, error) {
	modelMap := make(map[string]interface{})
	modelMap["more_info"] = model.MoreInfo
	return modelMap, nil
}

func resourceIBMIsVPCVPCRemoteToMap(model *vpcv1.VPCRemote) (map[string]interface{}, error) {
	modelMap := make(map[string]interface{})
	if model.Account != nil {
		accountMap, err := resourceIBMIsVPCAccountReferenceToMap(model.Account)
		if err != nil {
			return modelMap, err
		}
		modelMap["account"] = []map[string]interface{}{accountMap}
	}
	if model.Region != nil {
		modelMap["region"] = model.Region.Name
	}
	return modelMap, nil
}
func resourceIBMIsVPCAccountReferenceToMap(model *vpcv1.AccountReference) (map[string]interface{}, error) {
	modelMap := make(map[string]interface{})
	modelMap["id"] = model.ID
	modelMap["resource_type"] = model.ResourceType
	return modelMap, nil
}

func ResourceIBMIsVPCPublicAddressRangeReferenceToMap(model *vpcv1.PublicAddressRangeReference) (map[string]interface{}, error) {
	modelMap := make(map[string]interface{})
	modelMap["crn"] = *model.CRN
	if model.Deleted != nil {
		deletedMap, err := ResourceIBMIsVPCDeletedToMap(model.Deleted)
		if err != nil {
			return modelMap, err
		}
		modelMap["deleted"] = []map[string]interface{}{deletedMap}
	}
	modelMap["href"] = *model.Href
	modelMap["id"] = *model.ID
	modelMap["name"] = *model.Name
	modelMap["resource_type"] = *model.ResourceType
	return modelMap, nil
}
func ResourceIBMIsVPCDeletedToMap(model *vpcv1.Deleted) (map[string]interface{}, error) {
	modelMap := make(map[string]interface{})
	modelMap["more_info"] = *model.MoreInfo
	return modelMap, nil
}
