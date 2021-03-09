package paymail

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

// newTestClient will return a client for testing purposes
func newTestClient() (*Client, error) {
	// Create a Resty Client
	client := resty.New()

	// Get the underlying HTTP Client and set it to Mock
	httpmock.ActivateNonDefault(client.GetClient())

	// Create a new client
	newClient, err := NewClient(WithRequestTracing(), WithDNSTimeout(15*time.Second))
	if err != nil {
		return nil, err
	}
	newClient.WithCustomHTTPClient(client)
	// Set the customer resolver with known defaults
	r := newCustomResolver(
		newClient.resolver,
		map[string][]string{
			testDomain:      {"44.225.125.175", "35.165.117.200", "54.190.182.236"},
			"norecords.com": {},
		},
		map[string][]*net.SRV{
			DefaultServiceName + DefaultProtocol + testDomain:      {{Target: "www." + testDomain, Port: 443, Priority: 10, Weight: 10}},
			"invalid" + DefaultProtocol + testDomain:               {{Target: "www." + testDomain, Port: 443, Priority: 10, Weight: 10}},
			DefaultServiceName + DefaultProtocol + "relayx.io":     {{Target: "relayx.io", Port: 443, Priority: 10, Weight: 10}},
			DefaultServiceName + DefaultProtocol + "norecords.com": {},
		},
		map[string][]net.IPAddr{
			"example.com": {net.IPAddr{IP: net.ParseIP("8.8.8.8"), Zone: "eth0"}},
		},
	)

	// Set the custom resolver
	newClient.WithCustomResolver(r)
	// Return the mocking client
	return newClient, nil
}

// TestNewClient will test the method NewClient()
func TestNewClient(t *testing.T) {
	t.Parallel()

	t.Run("default client", func(t *testing.T) {
		client, err := NewClient()
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, defaultDNSTimeout, client.options.dnsTimeout)
		assert.Equal(t, defaultDNSPort, client.options.dnsPort)
		assert.Equal(t, defaultUserAgent, client.options.userAgent)
		assert.Equal(t, defaultNameServerNetwork, client.options.nameServerNetwork)
		assert.Equal(t, defaultNameServer, client.options.nameServer)
		assert.Equal(t, defaultSSLTimeout, client.options.sslTimeout)
		assert.Equal(t, defaultSSLDeadline, client.options.sslDeadline)
		assert.Equal(t, defaultHTTPTimeout, client.options.httpTimeout)
		assert.Equal(t, defaultRetryCount, client.options.retryCount)
		assert.Equal(t, false, client.options.requestTracing)
		assert.NotEqual(t, 0, len(client.options.brfcSpecs))
		assert.Greater(t, len(client.options.brfcSpecs), 6)
	})

	t.Run("custom http client", func(t *testing.T) {
		customHTTPClient := resty.New()
		customHTTPClient.SetTimeout(defaultHTTPTimeout)
		client, err := NewClient()
		assert.NoError(t, err)
		assert.NotNil(t, client)
		client.WithCustomHTTPClient(customHTTPClient)
	})

	t.Run("custom options", func(t *testing.T) {
		var client *Client
		client, err := NewClient(WithUserAgent("custom user agent"))
		assert.NotNil(t, client)
		assert.NoError(t, err)
	})

	t.Run("custom resolver", func(t *testing.T) {
		r := newCustomResolver(nil, nil, nil, nil)
		client, err := NewClient()
		assert.NotNil(t, client)
		assert.NoError(t, err)
		client.WithCustomResolver(r)
	})

	t.Run("no brfcs", func(t *testing.T) {
		var client *Client
		client, err := NewClient(WithBRFCSpecs(nil))
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})
}

// ExampleNewClient example using NewClient()
//
// See more examples in /examples/
func ExampleNewClient() {
	client, err := NewClient()
	if err != nil {
		fmt.Printf("error loading client: %s", err.Error())
		return
	}
	fmt.Printf("loaded client: %s", client.options.userAgent)
	// Output:loaded client: go-paymail: v0.1.6
}

// BenchmarkNewClient benchmarks the method NewClient()
func BenchmarkNewClient(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = NewClient(nil, nil, nil)
	}
}

// TestDefaultClientOptions will test the method DefaultClientOptions()
func TestDefaultClientOptions(t *testing.T) {
	t.Parallel()

	options, err := defaultClientOptions()
	assert.NoError(t, err)
	assert.NotNil(t, options)

	assert.Equal(t, defaultDNSTimeout, options.dnsTimeout)
	assert.Equal(t, defaultDNSPort, options.dnsPort)
	assert.Equal(t, defaultUserAgent, options.userAgent)
	assert.Equal(t, defaultNameServerNetwork, options.nameServerNetwork)
	assert.Equal(t, defaultNameServer, options.nameServer)
	assert.Equal(t, defaultSSLTimeout, options.sslTimeout)
	assert.Equal(t, defaultSSLDeadline, options.sslDeadline)
	assert.Equal(t, defaultHTTPTimeout, options.httpTimeout)
	assert.Equal(t, defaultRetryCount, options.retryCount)
	assert.Equal(t, false, options.requestTracing)
	assert.NotEqual(t, 0, len(options.brfcSpecs))
	assert.Greater(t, len(options.brfcSpecs), 6)
}

// ExampleDefaultClientOptions example using DefaultClientOptions()
//
// See more examples in /examples/
// nolint:govet // opts are now private
func ExampleDefaultClientOptions() {
	options, err := defaultClientOptions()
	if err != nil {
		fmt.Printf("error loading options: %s", err.Error())
		return
	}
	fmt.Printf("loaded options: %s", options.userAgent)
	// Output:loaded options: go-paymail: v0.1.6
}

// BenchmarkDefaultClientOptions benchmarks the method DefaultClientOptions()
func BenchmarkDefaultClientOptions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = defaultClientOptions()
	}
}
