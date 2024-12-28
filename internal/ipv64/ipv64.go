package ipv64

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"k8s.io/klog"
)

const apiURL = "https://ipv64.net/api.php"

// Record is a struct representing a DNS record
type Record struct {
	recordID int
	content  string
	dnsType  string
	praefix  string
	domain   string
}

type addRecordResponse struct {
	info      string `json:"info"`
	status    string `json:"status"`
	addRecord string `json:"add_record"`
}

// deleteRecordResponse is a struct representing the response of the ipv64 API when deleting a record
type deleteRecordResponse struct {
	info      string `json:"info"`
	status    string `json:"status"`
	delRecord string `json:"del_record"`
}

// Client is a struct representing the ipv64 client
type Client struct {
	apiURL string
	token  string
}

var client Client

// NewClient creates a new ipv64 client
func NewClient(token string) *Client {
	klog.Info("create new ipv64 client")
	if client == (Client{}) {
		client = Client{
			apiURL: apiURL,
			token:  token,
		}
	}
	return &client
}

// AddDNSRecord adds a DNS record to the ipv64 API
func (c *Client) AddDNSRecord(subdomain string, praefix string, content string, recordtype string) error {
	if recordtype != "TXT" && recordtype != "A" && recordtype != "CNAME" && recordtype != "MX" && recordtype != "NS" && recordtype != "PTR" && recordtype != "SRV" && recordtype != "SOA" && recordtype != "AAAA" {
		klog.Error("unsupported record type: ", recordtype)
		return fmt.Errorf("unsupported record type: %s", recordtype)
	}

	params := url.Values{}
	params.Set("add_record", subdomain)
	params.Set("praefix", praefix)
	params.Set("type", recordtype)
	params.Set("content", content)

	req, err := http.NewRequest("POST", c.apiURL, bytes.NewBufferString(params.Encode()))
	if err != nil {
		klog.Error("error creating request: ", err)
		return err
	}
	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+c.token)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		klog.Error("error sending request: ", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		klog.Error("error reading response body: ", err)
		return err
	}

	response := addRecordResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		klog.Error("error unmarshalling response body: ", err)
		return err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		if resp.StatusCode == http.StatusBadRequest && response.addRecord == "dns record already there" {
			klog.Warningln("DNS record already there")
			return nil
		}
		klog.Error("Could not add record: ", response.info, response)
		klog.V(4).Infoln("Response: ", response)
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	klog.Info("Added ", recordtype, " record ", praefix, subdomain)

	return nil
}

// DeleteDNSRecord deletes a DNS record from the ipv64 API
func (c *Client) DeleteDNSRecord(subdomain string, praefix string, content string, recordtype string) error {
	if recordtype != "TXT" && recordtype != "A" && recordtype != "CNAME" && recordtype != "MX" && recordtype != "NS" && recordtype != "PTR" && recordtype != "SRV" && recordtype != "SOA" && recordtype != "AAAA" {
		return fmt.Errorf("unsupported record type: %s", recordtype)
	}

	params := url.Values{}
	params.Set("del_record", subdomain)
	params.Set("praefix", praefix)
	params.Set("type", recordtype)
	params.Set("content", content)

	req, err := http.NewRequest("DELETE", c.apiURL, bytes.NewBufferString(params.Encode()))
	if err != nil {
		klog.Error("error creating request: ", err)
		return err
	}
	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+c.token)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		klog.Error("error sending request: ", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		klog.Error("error reading response body: ", err)
		return err
	}

	response := deleteRecordResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		klog.Error("error unmarshalling response body: ", err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusAccepted && response.delRecord == "del_record" {
			klog.Info("Deleted record ", praefix, ".", subdomain)
			return nil
		}
		klog.Error("Could not delete record: ", response.info, response)
		klog.V(4).Infoln("Response: ", response)
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	klog.Info("Deleted ", recordtype, " record ", praefix, subdomain)

	return nil
}
