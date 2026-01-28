package api

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/netdisco-tui/netdisco-tui/internal/config"
)

type Client struct {
	cfg        *config.Config
	httpClient *http.Client
}

func New(cfg *config.Config) *Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout:   cfg.HTTPTimeout,
			Transport: transport,
		},
	}
}

func (c *Client) doRequest(path string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", c.cfg.BaseURL, path)

	var lastErr error
	for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Authorization", c.cfg.APIKey)
		req.Header.Set("Accept", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if attempt < c.cfg.MaxRetries {
				time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
				continue
			}
			return nil, fmt.Errorf("request failed after %d retries: %w", c.cfg.MaxRetries, lastErr)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, fmt.Errorf("authentication failed — check NETDISCO_TOKEN")
		}
		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(body))
		}
		return body, nil
	}
	return nil, lastErr
}

func (c *Client) parseArray(body []byte) ([]map[string]interface{}, error) {
	var arr []map[string]interface{}
	if err := json.Unmarshal(body, &arr); err == nil {
		return arr, nil
	}
	var single map[string]interface{}
	if err := json.Unmarshal(body, &single); err == nil {
		return []map[string]interface{}{single}, nil
	}
	return nil, fmt.Errorf("failed to parse response")
}

func (c *Client) parseObject(body []byte) (map[string]interface{}, error) {
	var obj map[string]interface{}
	if err := json.Unmarshal(body, &obj); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return obj, nil
}

// SearchNode searches by IP or MAC
func (c *Client) SearchNode(query string) ([]map[string]interface{}, error) {
	body, err := c.doRequest(fmt.Sprintf("search/node?q=%s&archived=false", query))
	if err != nil {
		return nil, err
	}
	return c.parseArray(body)
}

// GetDevice gets device details by IP
func (c *Client) GetDevice(ip string) (map[string]interface{}, error) {
	body, err := c.doRequest(fmt.Sprintf("object/device/%s", ip))
	if err != nil {
		return nil, err
	}
	return c.parseObject(body)
}

// GetDevicePorts gets all ports for a device
func (c *Client) GetDevicePorts(ip string) ([]map[string]interface{}, error) {
	body, err := c.doRequest(fmt.Sprintf("object/device/%s/ports", ip))
	if err != nil {
		return nil, err
	}
	return c.parseArray(body)
}

// GetPortActiveNodes gets active nodes on a specific port
func (c *Client) GetPortActiveNodes(ip, port string) ([]map[string]interface{}, error) {
	body, err := c.doRequest(fmt.Sprintf("object/device/%s/port/%s/active_nodes", ip, port))
	if err != nil {
		return nil, err
	}
	return c.parseArray(body)
}

// GetDeviceNeighbors gets CDP/LLDP neighbors
func (c *Client) GetDeviceNeighbors(ip string, hops int) (map[string]interface{}, error) {
	body, err := c.doRequest(fmt.Sprintf("object/device/%s/neighbors?scope=depth&hops=%d", ip, hops))
	if err != nil {
		return nil, err
	}
	return c.parseObject(body)
}

// GetDeviceVlans gets VLANs configured on a device
func (c *Client) GetDeviceVlans(ip string) ([]map[string]interface{}, error) {
	body, err := c.doRequest(fmt.Sprintf("object/device/%s/vlans", ip))
	if err != nil {
		return nil, err
	}
	return c.parseArray(body)
}

// SearchDevice searches for devices
func (c *Client) SearchDevice(query, location, vendor, model string) ([]map[string]interface{}, error) {
	path := "search/device?"
	params := []string{}
	if query != "" {
		params = append(params, "q="+query)
	}
	if location != "" {
		params = append(params, "location="+location)
	}
	if vendor != "" {
		params = append(params, "vendor="+vendor)
	}
	if model != "" {
		params = append(params, "model="+model)
	}
	for i, p := range params {
		if i > 0 {
			path += "&"
		}
		path += p
	}
	body, err := c.doRequest(path)
	if err != nil {
		return nil, err
	}
	return c.parseArray(body)
}

// SearchPort searches for ports
func (c *Client) SearchPort(query string, includeUplinks bool) ([]map[string]interface{}, error) {
	uplink := "false"
	if includeUplinks {
		uplink = "true"
	}
	body, err := c.doRequest(fmt.Sprintf("search/port?q=%s&uplink=%s", query, uplink))
	if err != nil {
		return nil, err
	}
	return c.parseArray(body)
}

// SearchVlan searches for VLANs
func (c *Client) SearchVlan(query string) ([]map[string]interface{}, error) {
	body, err := c.doRequest(fmt.Sprintf("search/vlan?q=%s", query))
	if err != nil {
		return nil, err
	}
	return c.parseArray(body)
}

// GetDeviceInventory gets full device inventory
func (c *Client) GetDeviceInventory() ([]map[string]interface{}, error) {
	body, err := c.doRequest("report/device/deviceinventory")
	if err != nil {
		return nil, err
	}
	return c.parseArray(body)
}

// GetVlanInventory gets VLAN inventory
func (c *Client) GetVlanInventory() ([]map[string]interface{}, error) {
	body, err := c.doRequest("report/vlan/vlaninventory")
	if err != nil {
		return nil, err
	}
	return c.parseArray(body)
}

// GetRecentDevices gets recently added devices
func (c *Client) GetRecentDevices(days int) ([]map[string]interface{}, error) {
	body, err := c.doRequest(fmt.Sprintf("report/device/recentlyaddeddevices?since=%d", days))
	if err != nil {
		return nil, err
	}
	return c.parseArray(body)
}

// GetSubnetUtilization gets subnet utilization report
func (c *Client) GetSubnetUtilization(subnet string) ([]map[string]interface{}, error) {
	body, err := c.doRequest(fmt.Sprintf("report/ip/subnets?subnet=%s", subnet))
	if err != nil {
		return nil, err
	}
	return c.parseArray(body)
}

// GetIpInventory gets IP inventory for a subnet
func (c *Client) GetIpInventory(subnet string, limit int) ([]map[string]interface{}, error) {
	body, err := c.doRequest(fmt.Sprintf("report/ip/ipinventory?subnet=%s&limit=%d", subnet, limit))
	if err != nil {
		return nil, err
	}
	return c.parseArray(body)
}
