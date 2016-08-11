package ddcloud

import (
	"fmt"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

const (
	resourceKeyVirtualListenerName                   = "name"
	resourceKeyVirtualListenerDescription            = "description"
	resourceKeyVirtualListenerType                   = "type"
	resourceKeyVirtualListenerProtocol               = "protocol"
	resourceKeyVirtualListenerIPv4Address            = "ipv4"
	resourceKeyVirtualListenerPort                   = "port"
	resourceKeyVirtualListenerEnabled                = "enabled"
	resourceKeyVirtualListenerConnectionLimit        = "connection_limit"
	resourceKeyVirtualListenerConnectionRateLimit    = "connection_rate_limit"
	resourceKeyVirtualListenerSourcePortPreservation = "source_port_preservation"
	resourceKeyVirtualListenerPoolID                 = "pool"
	resourceKeyVirtualListenerPersistenceProfileName = "persistence_profile"
	resourceKeyVirtualListenerIRuleIDs               = "irules"
	resourceKeyVirtualListenerOptimizationProfiles   = "optimization_profiles"
	resourceKeyVirtualListenerNetworkDomainID        = "networkdomain"
)

func resourceVirtualListener() *schema.Resource {
	return &schema.Resource{
		Create: resourceVirtualListenerCreate,
		Read:   resourceVirtualListenerRead,
		Exists: resourceVirtualListenerExists,
		Update: resourceVirtualListenerUpdate,
		Delete: resourceVirtualListenerDelete,

		Schema: map[string]*schema.Schema{
			resourceKeyVirtualListenerName: &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			resourceKeyVirtualListenerDescription: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			resourceKeyVirtualListenerType: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  compute.VirtualListenerTypeStandard,
				ValidateFunc: func(data interface{}, fieldName string) (messages []string, errors []error) {
					listenerType := data.(string)
					switch listenerType {
					case compute.VirtualListenerTypeStandard:
					case compute.VirtualListenerTypePerformanceLayer4:
						return
					default:
						errors = append(errors, fmt.Errorf("Invalid virtual listener type '%s'.", listenerType))
					}

					return
				},
			},
			resourceKeyVirtualListenerProtocol: &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			resourceKeyVirtualListenerIPv4Address: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  nil,
			},
			resourceKeyVirtualListenerPort: &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
				ForceNew: true,
			},
			resourceKeyVirtualListenerEnabled: &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			resourceKeyVirtualListenerConnectionLimit: &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  20000,
				ValidateFunc: func(data interface{}, fieldName string) (messages []string, errors []error) {
					connectionRate := data.(int)
					if connectionRate > 0 {
						return
					}

					errors = append(errors,
						fmt.Errorf("Connection rate ('%s') must be greater than 0.", fieldName),
					)

					return
				},
			},
			resourceKeyVirtualListenerConnectionRateLimit: &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  2000,
				ValidateFunc: func(data interface{}, fieldName string) (messages []string, errors []error) {
					connectionRate := data.(int)
					if connectionRate > 0 {
						return
					}

					errors = append(errors,
						fmt.Errorf("Connection rate limit ('%s') must be greater than 0.", fieldName),
					)

					return
				},
			},
			resourceKeyVirtualListenerSourcePortPreservation: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  compute.SourcePortPreservationEnabled,
			},
			resourceKeyVirtualListenerPoolID: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  nil,
			},
			resourceKeyVirtualListenerPersistenceProfileName: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			resourceKeyVirtualListenerIRuleIDs: &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set: func(item interface{}) int {
					iRuleID := item.(string)

					return schema.HashString(iRuleID)
				},
			},
			resourceKeyVirtualListenerOptimizationProfiles: &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set: func(item interface{}) int {
					optimizationProfile := item.(string)

					return schema.HashString(optimizationProfile)
				},
			},
			resourceKeyVirtualListenerNetworkDomainID: &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			// TODO: Add remaining properties.
		},
	}
}

func resourceVirtualListenerCreate(data *schema.ResourceData, provider interface{}) error {
	networkDomainID := data.Get(resourceKeyVirtualListenerNetworkDomainID).(string)
	name := data.Get(resourceKeyVirtualListenerName).(string)
	description := data.Get(resourceKeyVirtualListenerDescription).(string)

	log.Printf("Create virtual listener '%s' ('%s') in network domain '%s'.", name, description, networkDomainID)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	persistenceProfileID, err := getPersistenceProfileID(apiClient, data)
	if err != nil {
		return err
	}

	propertyHelper := propertyHelper(data)

	virtualListenerID, err := apiClient.CreateVirtualListener(compute.NewVirtualListenerConfiguration{
		Name:                   name,
		Description:            description,
		Type:                   data.Get(resourceKeyVirtualListenerType).(string),
		Protocol:               data.Get(resourceKeyVirtualListenerProtocol).(string),
		Port:                   data.Get(resourceKeyVirtualListenerPort).(int),
		ListenerIPAddress:      propertyHelper.GetOptionalString(resourceKeyVirtualListenerIPv4Address, false),
		Enabled:                data.Get(resourceKeyVirtualListenerEnabled).(bool),
		ConnectionLimit:        data.Get(resourceKeyVirtualListenerConnectionLimit).(int),
		ConnectionRateLimit:    data.Get(resourceKeyVirtualListenerConnectionRateLimit).(int),
		SourcePortPreservation: data.Get(resourceKeyVirtualListenerSourcePortPreservation).(string),
		PoolID:                 propertyHelper.GetOptionalString(resourceKeyVirtualListenerPoolID, false),
		PersistenceProfileID:   persistenceProfileID,
		IRuleIDs:               propertyHelper.GetStringSetItems(resourceKeyVirtualListenerIRuleIDs),
		OptimizationProfiles:   propertyHelper.GetStringSetItems(resourceKeyVirtualListenerOptimizationProfiles),
		NetworkDomainID:        networkDomainID,
	})
	if err != nil {
		return err
	}

	data.SetId(virtualListenerID)

	log.Printf("Successfully created virtual listener '%s'.", virtualListenerID)

	virtualListener, err := apiClient.GetVirtualListener(virtualListenerID)
	if err != nil {
		return err
	}
	if virtualListener == nil {
		return fmt.Errorf("Cannot find newly-created virtual listener with Id '%s'.", virtualListenerID)
	}

	// TODO: Populate computed properties.

	return nil
}

func resourceVirtualListenerExists(data *schema.ResourceData, provider interface{}) (bool, error) {
	id := data.Id()

	log.Printf("Check if virtual listener '%s' exists...", id)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	vipPool, err := apiClient.GetVirtualListener(id)
	if err != nil {
		return false, err
	}

	exists := vipPool != nil

	log.Printf("virtual listener '%s' exists: %t.", id, exists)

	return exists, nil
}

func resourceVirtualListenerRead(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()

	log.Printf("Read virtual listener '%s'...", id)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	virtualListener, err := apiClient.GetVirtualListener(id)
	if err != nil {
		return err
	}
	if virtualListener == nil {
		data.SetId("") // Virtual listener has been deleted

		return nil
	}

	data.Set(resourceKeyVirtualListenerDescription, virtualListener.Description)
	data.Set(resourceKeyVirtualListenerEnabled, virtualListener.Enabled)
	data.Set(resourceKeyVirtualListenerConnectionLimit, virtualListener.ConnectionLimit)
	data.Set(resourceKeyVirtualListenerConnectionRateLimit, virtualListener.ConnectionRateLimit)
	data.Set(resourceKeyVirtualListenerSourcePortPreservation, virtualListener.SourcePortPreservation)

	propertyHelper := propertyHelper(data)
	propertyHelper.SetVirtualListenerIRuleIDs(virtualListener.IRules)

	// TODO: Capture other properties.

	return nil
}

func resourceVirtualListenerUpdate(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	log.Printf("Update virtual listener '%s'...", id)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	configuration := &compute.EditVirtualListenerConfiguration{}

	propertyHelper := propertyHelper(data)
	if data.HasChange(resourceKeyVirtualListenerDescription) {
		configuration.Description = propertyHelper.GetOptionalString(resourceKeyVirtualListenerDescription, true)
	}

	if data.HasChange(resourceKeyVirtualListenerEnabled) {
		configuration.Enabled = propertyHelper.GetOptionalBool(resourceKeyVirtualListenerEnabled)
	}

	if data.HasChange(resourceKeyVirtualListenerConnectionLimit) {
		configuration.ConnectionLimit = propertyHelper.GetOptionalInt(resourceKeyVirtualListenerConnectionLimit, false)
	}

	if data.HasChange(resourceKeyVirtualListenerConnectionRateLimit) {
		configuration.ConnectionRateLimit = propertyHelper.GetOptionalInt(resourceKeyVirtualListenerConnectionRateLimit, false)
	}

	if data.HasChange(resourceKeyVirtualListenerSourcePortPreservation) {
		configuration.SourcePortPreservation = propertyHelper.GetOptionalString(resourceKeyVirtualListenerSourcePortPreservation, true)
	}

	if data.HasChange(resourceKeyVirtualListenerPoolID) {
		configuration.PoolID = propertyHelper.GetOptionalString(resourceKeyVirtualListenerPoolID, true)
	}

	if data.HasChange(resourceKeyVirtualListenerPersistenceProfileName) {
		persistenceProfileID, err := getPersistenceProfileID(apiClient, data)
		if err != nil {
			return err
		}

		configuration.PersistenceProfileID = persistenceProfileID
	}

	return apiClient.EditVirtualListener(id, *configuration)
}

func resourceVirtualListenerDelete(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	name := data.Get(resourceKeyVirtualListenerName).(string)
	networkDomainID := data.Get(resourceKeyVirtualListenerNetworkDomainID)

	log.Printf("Delete virtual listener '%s' ('%s') from network domain '%s'...", name, id, networkDomainID)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	return apiClient.DeleteVirtualListener(id)
}

func getPersistenceProfileID(apiClient *compute.Client, data *schema.ResourceData) (persistenceProfileID *string, err error) {
	value, ok := data.GetOk(resourceKeyVirtualListenerPersistenceProfileName)
	if !ok {
		return
	}
	persistenceProfileName := value.(string)

	networkDomainID := data.Get(resourceKeyVirtualListenerNetworkDomainID).(string)

	page := compute.DefaultPaging()
	for {
		var persistenceProfiles *compute.PersistenceProfiles
		persistenceProfiles, err = apiClient.ListDefaultPersistenceProfiles(networkDomainID, page)
		if err != nil {
			return
		}
		if persistenceProfiles.IsEmpty() {
			break // We're done
		}

		for _, profile := range persistenceProfiles.Items {
			if profile.Name == persistenceProfileName {
				persistenceProfileID = &profile.ID

				return
			}
		}

		page.Next()
	}

	return
}
