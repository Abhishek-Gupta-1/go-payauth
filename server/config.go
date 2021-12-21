package server

import (
	"errors"
	"strings"
	"time"

	"github.com/mrz1836/go-sanitize"
	"github.com/tonicpow/go-paymail"
)

// Configuration paymail server configuration object
type Configuration struct {
	APIVersion              string        `json:"api_version"`
	BasicRoutes             *basicRoutes  `json:"basic_routes"`
	BSVAliasVersion         string        `json:"bsv_alias_version"`
	Capabilities            *Capabilities `json:"capabilities"`
	PaymailDomains          []*Domain     `json:"paymail_domains"`
	Port                    int           `json:"port"`
	Prefix                  string        `json:"prefix"`
	SenderValidationEnabled bool          `json:"sender_validation_enabled"`
	ServiceName             string        `json:"service_name"`
	Timeout                 time.Duration `json:"timeout"`

	// private
	actions PaymailServiceProvider
}

// Domain is the Paymail Domain information
type Domain struct {
	Name string `json:"name"`
}

// Validate will check that the configuration meets a minimum requirement to run the server
func (c *Configuration) Validate() error {

	// Requires domains for the server to run
	if len(c.PaymailDomains) == 0 {
		return errors.New("missing a paymail domain")
	}

	// Requires a port
	if c.Port <= 0 {
		return errors.New("missing a port")
	}

	// todo: validate the []domains

	// Sanitize and standardize the service name
	c.ServiceName = sanitize.PathName(c.ServiceName)
	if len(c.ServiceName) == 0 {
		return errors.New("missing service name")
	}

	// Validate (basic checks for existence of capabilities)
	if c.Capabilities == nil {
		return errors.New("missing capabilities struct")
	} else if len(c.Capabilities.BsvAlias) == 0 {
		return errors.New("missing bsv alias version")
	} else if len(c.Capabilities.Capabilities) == 0 {
		return errors.New("missing capabilities")
	}

	return nil
}

// IsAllowedDomain will return true if it's an allowed paymail domain
func (c *Configuration) IsAllowedDomain(domain string) (success bool) {

	// Sanitize the domain (standard)
	var err error
	if domain, err = sanitize.Domain(
		domain, false, true,
	); err != nil {
		return
	}

	// Loop all domains check
	// todo: make this faster with an init that creates a hash map?
	for _, d := range c.PaymailDomains {
		if strings.EqualFold(d.Name, domain) {
			success = true
			break
		}
	}

	return
}

// AddDomain will add the domain if it does not exist
func (c *Configuration) AddDomain(domain string) (err error) {

	// Sanity check
	if len(domain) == 0 {
		return errors.New("domain is missing")
	}

	// Sanitize and standardize
	domain, err = sanitize.Domain(domain, false, true)
	if err != nil {
		return
	}

	// Already exists?
	if c.IsAllowedDomain(domain) {
		return
	}

	// Create the domain
	c.PaymailDomains = append(c.PaymailDomains, &Domain{Name: domain})
	return
}

// EnrichCapabilities will update the capabilities with the appropriate service url
func (c *Configuration) EnrichCapabilities(domain string) {
	for key, val := range c.Capabilities.Capabilities {
		if w, ok := val.(string); ok {
			c.Capabilities.Capabilities[key] = GenerateServiceURL(c.Prefix, domain, c.APIVersion, c.ServiceName) + w
		}
	}
}

// GenerateServiceURL will create the service URL
func GenerateServiceURL(prefix, domain, apiVersion, serviceName string) string {
	return prefix + domain + "/" + apiVersion + "/" + serviceName
}

// NewConfig will make a new server configuration
func NewConfig(serviceProvider PaymailServiceProvider, opts ...ConfigOps) (*Configuration, error) {

	// Create the base configuration
	config := defaultConfigOptions()

	// Overwrite defaults with any set by user
	for _, opt := range opts {
		opt(config)
	}

	// todo: generic capabilities

	// Validate the configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Set the service provider
	config.actions = serviceProvider

	return config, nil
}

// NewConfiguration create a new Configuration for the paymail server
func NewConfiguration(paymailDomain string, serverInterface PaymailServiceProvider) *Configuration {
	config := &Configuration{
		BasicRoutes:             &basicRoutes{},
		actions:                 serverInterface,
		Port:                    DefaultServerPort,
		SenderValidationEnabled: DefaultSenderValidation,
		ServiceName:             paymail.DefaultServiceName,
		// ServiceURL:              "https://" + paymailDomain + "/" + DefaultAPIVersion + "/" + paymail.DefaultServiceName + "/",
		Timeout: DefaultTimeout,
	}

	// config.Capabilities = createCapabilities(config)

	return config
}
